package macroproc

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMacroModifiersDeletion(t *testing.T) {
	conv := New()

	tobj := []byte(`{
		"Kind": "some",
		"kubegen.Null.Delete": { "aFoodOrder": "YES" },
		"order": {
			"potatoe": {
				"mash": { "count": 1 },
				"chips": { "count": 2 },
				"sausages": true,
				"gravy": {
					"beef": 1,
					"chicken": 0,
					"other": [
						{
							"kind": "sweetAndSour",
							"kubegen.Null.Delete": "someOfThatMayBe"
						},
						{
							"kind": "sourCream",
							"kubegen.Null.Delete": "yesExtraOfThatPleaseThankYou"
						}
					]
				}
			}
		}
	}`)

	umobj := []byte(`{
		"Kind": "some",
		"order": {
			"potatoe": {
				"mash": { "count": 1 },
				"chips": { "count": 2 },
				"sausages": true,
				"gravy": {
					"beef": 1,
					"chicken": 0,
					"other": [
						{
							"kind": "sweetAndSour"
						},
						{
							"kind": "sourCream"
						}
					]
				}
			}
		}
	}`)

	assert := assert.New(t)

	assert.NotEqual(tobj, umobj,
		"object without modifier macros should be different")
	assert.True(len(tobj) > len(umobj),
		"object without modifier macros should be larger")

	conv.DefineMacro(&Macro{
		ReturnType: Null,
		EvalPhase:  MacrosEvalPhaseA,
		VerbName:   "Delete",
	}, func(c *Converter, branch *BranchLocator, _ *Macro) (ModifierCallback, error) {
		switch branch.kind {
		case String:
			fallthrough
		case Object:
			// TODO panic if key exists or find a way to have unique keys
			cb := func(_ *Modifier, c *Converter) error {
				return c.Delete(branch)
			}
			return cb, nil
		}
		return nil, nil
	})

	if err := conv.loadStrict(tobj); err != nil {
		t.Fatalf("failed to laod – %v\ntree=%s", err, conv.tree)
	}

	if err := conv.Run(); err != nil {
		t.Fatalf("failed to convert – %v\ntree=%s\nlocator.self=%#v", err, conv.tree, conv.locator.self)
	}

	unmodified := New()
	if err := unmodified.loadStrict(umobj); err != nil {
		t.Fatalf("failed to laod – %v\ntree=%s", err, unmodified.tree)
	}

	assert.Equal(unmodified.tree.Len(), conv.tree.Len(),
		"new object should be the same len as one without modifier macros")

	assert.JSONEq(unmodified.tree.String(), conv.tree.String(),
		"new object should be exactly the same as one without modifier macros")

	assert.Equal(0, len(unmodified.macros[MacrosEvalPhaseA]),
		"object without modifier macros has no macro handlers")
	assert.Equal(0, len(unmodified.modifiers),
		"object without modifier macros has no macro callbacks")

	reloaded := New()
	if err := unmodified.loadStrict(conv.tree.Bytes()); err != nil {
		t.Fatalf("failed to laod – %v\ndata=%s", err, reloaded.tree)
	}
}

func TestMacroErrorsAndModifiersLookup(t *testing.T) {
	conv := New()

	assert := assert.New(t)

	tobj := []byte(`{
		"Kind": "Some",
		"test1s": {
			"kubegen.String.Lookup": "test1val"
		},
		"test2s": {
			"kubegen.String.Lookup": "test2val"
		},
		"test3n": {
			"kubegen.Number.Lookup": "test3val"
		},
		"test4n": {
			"kubegen.Number.Lookup": "test4val"
		}
	}`)

	badSyntax := [][]byte{
		[]byte(`{ { "kubegen.String.Lookup": []`),
		[]byte(`foo: {}`),
		[]byte(`{ "Kind" "SOME" }`),
	}

	badModfiersOrObjecs := [][]byte{
		[]byte(`{ "Something": null }`),
		[]byte(`{ "Kind": "Some", "test5sf": { "kubegen.String.Lookup": [] } }`),
		[]byte(`{ "Kind": "Some", "test6nf": { "kubegen.Number.Lookup": {} } }`),
	}

	conv.DefineMacro(MacroStringLookup,
		func(c *Converter, branch *BranchLocator, _ *Macro) (ModifierCallback, error) {
			p := branch.PathToString()
			switch branch.kind {
			case String:
				cb := func(_ *Modifier, c *Converter) error {
					if err := c.Set(branch, "TEST"); err != nil {
						return fmt.Errorf("could not set string – %v", err)
					}
					return nil
				}
				return cb, nil
			default:
				return nil, fmt.Errorf("in %q value is a %s, but must be a string", p, branch.kind)
			}
		})

	conv.DefineMacro(MacroNumberLookup,
		func(c *Converter, branch *BranchLocator, _ *Macro) (ModifierCallback, error) {
			p := branch.PathToString()
			switch branch.kind {
			case String:
				cb := func(_ *Modifier, c *Converter) error {
					if err := c.Set(branch, 1); err != nil {
						return fmt.Errorf("could not set number – %v", err)
					}
					return nil
				}
				return cb, nil
			default:
				return nil, fmt.Errorf("in %q value is a %s, but must be a string", p, branch.kind)
			}
		})

	for _, v := range badSyntax {
		conv2 := New()
		conv2.macros = conv.macros
		var err error
		err = conv2.loadStrict(v)
		assert.NotNil(err)
	}

	for _, v := range badModfiersOrObjecs {
		conv2 := New()
		conv2.macros = conv.macros
		var err error
		err = conv2.loadStrict(v)
		assert.Nil(err)
		err = conv2.run(MacrosEvalPhaseB)
		assert.NotNil(err)
	}

	if err := conv.loadStrict(tobj); err != nil {
		t.Fatalf("failed to laod – %v\ntree=%s", err, conv.tree)
	}

	if err := conv.Run(); err != nil {
		t.Fatalf("failed to convert – %v\ntree=%s\nc.locator.self=%#v", err, conv.tree, conv.locator.self)
	}

	{
		v, err := conv.tree.GetString("test1s")
		assert.Nil(err)
		assert.Equal("TEST", v)
	}

	{
		v, err := conv.tree.GetString("test2s")
		assert.Nil(err)
		assert.Equal("TEST", v)
	}

	{
		v, err := conv.tree.GetInt("test3n")
		assert.Nil(err)
		assert.Equal(1, v)
	}

	{
		v, err := conv.tree.GetInt("test4n")
		assert.Nil(err)
		assert.Equal(1, v)
	}
}

func TestMacroLookupSimpleObjectOnly(t *testing.T) {
	conv := New()

	assert := assert.New(t)

	tobj := []byte(`{
		"Kind": "Some",
		"test1s": [{
			"kubegen.Object.Lookup": "testInsertObj6"
		}],
		"test2s": {
			"kubegen.Object.Lookup": "testInsertObj6"
		}
	}`)

	objs := map[string][]byte{
		"testInsertObj6": []byte(`{ "testObj": { "test1": 1, "test2": 2 }, "testStr": "str" }`),
	}

	conv.DefineMacro(MacroObjectLookup,
		func(c *Converter, branch *BranchLocator, _ *Macro) (ModifierCallback, error) {
			p := branch.PathToString()
			var js []byte
			k := branch.StringValue()
			if k == nil {
				return nil, fmt.Errorf("unexpect nil string value of %s", p)
			}
			if v, ok := objs[*k]; ok {
				js = v
			} else {
				js = []byte(`{ }`)
			}

			x, err := loadObject(js)
			if err != nil {
				return nil, err
			}
			switch branch.kind {
			case String:
				cb := func(_ *Modifier, c *Converter) error {
					if err := c.Overlay(branch, x.self); err != nil {
						return fmt.Errorf("%v – c.tree=%s", err, c.tree)
					}
					return nil
				}
				return cb, nil
			default:
				return nil, fmt.Errorf("in %q value is a %s, but must be a string", p, branch.kind)
			}
		})

	if err := conv.loadStrict(tobj); err != nil {
		t.Fatalf("failed to laod – %v\ntree=%s", err, conv.tree)
	}

	if err := conv.Run(); err != nil {
		t.Fatalf("failed to run converter – %v", err)
	}

	{
		v, err := conv.tree.GetObject("test1s", 0)
		assert.Nil(err)

		js, err := json.Marshal(v)
		assert.Nil(err)

		assert.JSONEq(string(objs["testInsertObj6"]), string(js))
	}

	{
		v, err := conv.tree.GetObject("test2s")
		assert.Nil(err)

		js, err := json.Marshal(v)
		assert.Nil(err)

		assert.JSONEq(string(objs["testInsertObj6"]), string(js))
	}
}

func TestMacroLookupRecursive(t *testing.T) {
	conv := New()

	assert := assert.New(t)

	tobj := []byte(`{
		"Kind": "Some",
		"test1s": {
			"kubegen.String.Lookup": "test1val"
		},
		"test2n": {
			"kubegen.Number.Lookup": "test2val"
		},
		"test3m": {
			"kubegen.Array.Lookup": "testInsertArray1"
		},
		"test4o": {
			"kubegen.Object.Lookup": "testInsertObj1"
		},
		"test5o": {
			"kubegen.Object.Lookup": "testInsertObj9"
		},
		"test6o": [
			{
				"kubegen.Object.Lookup": "testInsertObj6",
				"testObj": { "test0": 0 }
			},
			{
				"kubegen.Object.Lookup": "testInsertObj6"
			},
			{
				"kubegen.Object.Lookup": "testInsertObj6",
				"testObj": {}
			},
			{
				"kubegen.Object.Lookup": "testInsertObj6",
				"testStr": ""
			}
		]
	}`)

	objs := map[string][]byte{
		"testInsertObj1": []byte(`{
			"foo": [
				{ "kubegen.Number.Lookup": "test4val" },
				{ "kubegen.String.Lookup": "test4val" },
				{ "kubegen.Object.Lookup": "testInsertObj2" }
			],
			"bar": [
				{ "kubegen.Object.Lookup": "testInsertObj3" }
			]
		}`),
		"testInsertObj2":   []byte(`{ "bar": "TEST_VAL4" }`),
		"testInsertObj3":   []byte(`{ "kubegen.Object.Lookup": "testInsertObj4" }`),
		"testInsertArray1": []byte(`[{ "a": { "kubegen.Number.Lookup": "test3valN" }},{ "kubegen.String.Lookup": "test3valS" },{ "kubegen.Object.Lookup": "testInsertObj2" }]`),
		"testInsertObj6":   []byte(`{ "testObj": { "test1": 1, "test2": 2 }, "testStr": "str" }`),
	}

	conv.DefineMacro(MacroStringLookup,
		func(c *Converter, branch *BranchLocator, _ *Macro) (ModifierCallback, error) {
			p := branch.PathToString()
			switch branch.kind {
			case String:
				cb := func(_ *Modifier, c *Converter) error {
					if err := c.Set(branch, "TEST_STRING"); err != nil {
						return fmt.Errorf("could not set string – %v", err)
					}
					return nil
				}
				return cb, nil
			default:
				return nil, fmt.Errorf("in %q value is a %s, but must be a string", p, branch.kind)
			}
		})

	conv.DefineMacro(MacroNumberLookup,
		func(c *Converter, branch *BranchLocator, _ *Macro) (ModifierCallback, error) {
			p := branch.PathToString()
			switch branch.kind {
			case String:
				cb := func(_ *Modifier, c *Converter) error {
					if err := c.Set(branch, 12345); err != nil {
						return fmt.Errorf("could not set number – %v", err)
					}
					return nil
				}
				return cb, nil
			default:
				return nil, fmt.Errorf("in %q value is a %s, but must be a string", p, branch.kind)
			}
		})

	conv.DefineMacro(MacroObjectLookup,
		func(c *Converter, branch *BranchLocator, _ *Macro) (ModifierCallback, error) {
			p := branch.PathToString()
			var js []byte
			k := branch.StringValue()
			if k == nil {
				return nil, fmt.Errorf("unexpect nil string value of %s", p)
			}
			if v, ok := objs[*k]; ok {
				js = v
			} else {
				js = []byte(`{ "empty": true }`)
			}
			x, err := loadObject(js)
			if err != nil {
				return nil, err
			}

			switch branch.kind {
			case String:
				cb := func(_ *Modifier, c *Converter) error {
					if err := c.Overlay(branch, x.self); err != nil {
						return fmt.Errorf("%v – c.tree=%s", err, c.tree)
					}
					return nil
				}
				return cb, nil
			default:
				return nil, fmt.Errorf("in %q value is a %s, but must be a string", p, branch.kind)
			}
		})

	conv.DefineMacro(MacroArrayLookup,
		func(c *Converter, branch *BranchLocator, _ *Macro) (ModifierCallback, error) {
			p := branch.PathToString()
			var js []byte
			k := branch.StringValue()
			if k == nil {
				return nil, fmt.Errorf("unexpect nil string value of %s", p)
			}
			if v, ok := objs[*k]; ok {
				js = v
			} else {
				js = []byte("[ 1, 2, 3 ]")
			}
			x, err := loadObject(js)
			if err != nil {
				return nil, err
			}

			switch branch.kind {
			case String:
				cb := func(_ *Modifier, c *Converter) error {
					if err := c.Overlay(branch, x.self); err != nil {
						return fmt.Errorf("%v – c.tree=%s", err, c.tree)
					}
					return nil
				}
				return cb, nil
			default:
				return nil, fmt.Errorf("in %q value is a %s, but must be a string", p, branch.kind)
			}
		})

	if err := conv.loadStrict(tobj); err != nil {
		t.Fatalf("failed to laod – %v\ntree=%s", err, conv.tree)
	}

	if err := conv.Run(); err != nil {
		t.Fatalf("failed to run converter – %v", err)
	}

	{
		v, err := conv.tree.GetString("test1s")
		assert.Nil(err)
		assert.Equal("TEST_STRING", v)
	}

	{
		v, err := conv.tree.GetInt("test2n")
		assert.Nil(err)
		assert.Equal(12345, v)
	}

	{
		v, err := conv.tree.GetValue("test3m")
		assert.Nil(err)

		a := fmt.Sprintf(`[{"a":12345},"TEST_STRING",%s]`, objs["testInsertObj2"])

		js, err := json.Marshal(v)
		assert.Nil(err)

		assert.JSONEq(a, string(js))
	}

	{
		v, err := conv.tree.GetValue("test5o")
		assert.Nil(err)

		js, err := json.Marshal(v)
		assert.Nil(err)

		assert.JSONEq(`{ "empty": true }`, string(js))
	}

	{
		v, err := conv.tree.GetString("test1s")
		assert.Nil(err)
		assert.Equal("TEST_STRING", v)
	}

	{
		v, err := conv.tree.GetInt("test2n")
		assert.Nil(err)
		assert.Equal(12345, v)
	}

	{
		v, err := conv.tree.GetArray("test3m")
		assert.Nil(err)

		a := fmt.Sprintf(`[{"a":12345},"TEST_STRING",%s]`, objs["testInsertObj2"])

		js, err := json.Marshal(v)
		assert.Nil(err)

		assert.JSONEq(a, string(js))
	}

	{
		_, err := conv.tree.GetObject("test4o")
		assert.Nil(err)
	}

	{
		v, err := conv.tree.GetInt("test4o", "foo", 0)
		assert.Nil(err)
		assert.Equal(12345, v)
	}

	{
		v, err := conv.tree.GetString("test4o", "foo", 1)
		assert.Nil(err)
		assert.Equal("TEST_STRING", v)
	}

	{
		v, err := conv.tree.GetObject("test4o", "foo", 2)
		assert.Nil(err)

		js, err := json.Marshal(v)
		assert.Nil(err)

		assert.JSONEq(string(objs["testInsertObj2"]), string(js))

	}

	{
		v, err := conv.tree.GetObject("test5o")
		assert.Nil(err)

		js, err := json.Marshal(v)
		assert.Nil(err)

		assert.JSONEq(`{ "empty": true }`, string(js))
	}

	{
		v, err := conv.tree.GetObject("test4o", "bar", 0)
		assert.Nil(err)

		js, err := json.Marshal(v)
		assert.Nil(err)

		assert.JSONEq(`{ "empty": true }`, string(js))
	}

	{
		v, err := conv.tree.GetInt("test4o", "foo", 0)
		assert.Nil(err)
		assert.Equal(12345, v)
	}

	{
		v, err := conv.tree.GetString("test4o", "foo", 1)
		assert.Nil(err)
		assert.Equal("TEST_STRING", v)
	}

	{
		v, err := conv.tree.GetObject("test4o", "foo", 2)
		assert.Nil(err)

		js, err := json.Marshal(v)
		assert.Nil(err)

		assert.JSONEq(string(objs["testInsertObj2"]), string(js))
	}

	{
		v, err := conv.tree.GetObject("test5o")
		assert.Nil(err)

		js, err := json.Marshal(v)
		assert.Nil(err)

		assert.JSONEq(`{ "empty": true }`, string(js))
	}

	{
		v, err := conv.tree.GetObject("test4o", "bar", 0)
		assert.Nil(err)

		js, err := json.Marshal(v)
		assert.Nil(err)

		assert.JSONEq(`{ "empty": true }`, string(js))
	}

	{
		v, err := conv.tree.GetObject("test6o", 0)
		assert.Nil(err)

		js, err := json.Marshal(v)
		assert.Nil(err)

		assert.JSONEq(`{ "testObj": { "test0": 0, "test1": 1, "test2": 2 }, "testStr": "str" }`, string(js))
	}
}

func TestMacroJoinStrings(t *testing.T) {
	conv := New()

	assert := assert.New(t)

	tobj1 := []byte(`{
			"Kind": "Some",
			"foobar": {
				"kubegen.String.Join": [
					"foo",
					"bar"
				]
			}
	}`)

	if err := conv.LoadObject(tobj1, "tobj1.json", ""); err != nil {
		t.Fatalf("failed to load – %v", err)
	}

	conv.DefineMacro(MacroStringJoin, MakeModifierStringJoin)

	if err := conv.Run(); err != nil {
		t.Logf("tree=%s", conv.tree)
		t.Fatalf("failed to convert – %v", err)
	}

	assert.Equal(0, len(conv.modifiers))

	{
		v, err := conv.tree.GetString("foobar")
		assert.Nil(err)
		assert.Equal("foobar", v)
	}
}

func TestMacroObjectToJSON(t *testing.T) {
	conv := New()

	assert := assert.New(t)

	tobj1 := []byte(`{
			"Kind": "Some",
			"foobar1": {
				"kubegen.String.AsJSON": [
					"foo",
					"bar"
				]
			},
			"foobar2": {
				"kubegen.String.AsYAML": [
					"foo",
					"bar"
				]
			},
			"foobar3": {
				"kubegen.String.AsYAML": {
					"foo": [],
					"bar": {}
				}
			},
			"foobar4": {
				"kubegen.String.Join": [
					"---",
					{
						"kubegen.String.AsYAML": {
							"foo": 1,
							"bar": 2
						}
					}
				]
			}
	}`)

	if err := conv.LoadObject(tobj1, "tobj1.json", ""); err != nil {
		t.Fatalf("failed to load – %v", err)
	}

	conv.DefineMacro(MacroStringAsJSON, MakeModifierStringAsJSON)
	conv.DefineMacro(MacroStringAsYAML, MakeModifierStringAsYAML)

	if err := conv.Run(); err != nil {
		t.Logf("tree=%s", conv.tree)
		t.Fatalf("failed to convert �� %v", err)
	}

	assert.Equal(0, len(conv.modifiers))

	{
		v, err := conv.tree.GetString("foobar1")
		assert.Nil(err)
		assert.Equal("[\"foo\",\"bar\"]", v)
	}

	{
		v, err := conv.tree.GetString("foobar2")
		assert.Nil(err)
		assert.Equal("- foo\n- bar\n", v)
	}

	{
		v, err := conv.tree.GetString("foobar3")
		assert.Nil(err)
		assert.Equal("bar: {}\nfoo: []\n", v)
	}
}

func TestMacroToString(t *testing.T) {
	assert := assert.New(t)

	m := &Macro{
		ReturnType: String,
		EvalPhase:  MacrosEvalPhaseA,
		VerbName:   "FooBar",
	}

	assert.Equal("kubegen.String.FooBar", m.String())
}

func _TestMacroLoadJSON(t *testing.T) {
	conv := New()

	assert := assert.New(t)

	tfiles := map[string][]byte{
		"TRUEO":  []byte(`{ "test": true }`),
		"FALSEO": []byte(`{ "test": false }`),
		"TRUEA":  []byte(`[ "test", true ]`),
		"FALSEA": []byte(`[ "test", false ]`),
		"RECURSIVE": []byte(`[
			{ "kubegen.Array.LoadJSON": "TRUEA" },
			{ "kubegen.Array.LoadJSON": "FALSEA" }
		]`),
	}

	tobj := []byte(`{
			"Kind": "Some",
			"kubegen.Object.LoadJSON": "TRUEO",
			"other": {
				"kubegen.Object.LoadJSON": "FALSEO"
			},
			"another": {
				"kubegen.Object.LoadJSON": "TRUEO",
				"something": [
					{ "kubegen.Array.LoadJSON": "TRUEA" },
					{ "kubegen.Array.LoadJSON": "FLASEA" }
				]
			},
			"more": {
				"kubegen.Array.LoadJSON": "RECURSIVE"
			}
	}`)

	conv.DefineMacro(LoadObjectJSON,
		func(c *Converter, branch *BranchLocator, _ *Macro) (ModifierCallback, error) {
			var newData []byte
			k := branch.StringValue()
			if v, ok := tfiles[*k]; ok {
				newData = v
			} else {
				newData = []byte("{ }")
			}
			return MakeObjectLoadJSON(c, branch, newData)
		})

	conv.DefineMacro(LoadArrayJSON,
		func(c *Converter, branch *BranchLocator, _ *Macro) (ModifierCallback, error) {
			var newData []byte
			k := branch.StringValue()
			if v, ok := tfiles[*k]; ok {
				newData = v
			} else {
				newData = []byte("[ ]")
			}
			return MakeArrayLoadJSON(c, branch, newData)
		})

	if err := conv.LoadObject(tobj, "tobj1.json", ""); err != nil {
		t.Fatalf("failed to load – %v", err)
	}

	if err := conv.Run(); err != nil {
		t.Logf("tree=%s", conv.tree)
		t.Fatalf("failed to convert – %v", err)
	}

	{
		v, err := conv.tree.GetString("Kind")
		assert.Nil(err)
		assert.Equal("Some", v)
	}

	{
		v, err := conv.tree.GetBoolean("test")
		assert.Nil(err)
		assert.Equal(true, v)
	}

	{
		v, err := conv.tree.GetObject("other")
		assert.Nil(err)

		js, err := json.Marshal(v)
		assert.Nil(err)

		assert.JSONEq(`{"test":false}`, string(js))
	}

	{
		v, err := conv.tree.GetBoolean("another", "test")
		assert.Nil(err)
		assert.Equal(true, v)
	}

	{
		v, err := conv.tree.GetArray("more")
		assert.Nil(err)

		js, err := json.Marshal(v)
		assert.Nil(err)

		assert.JSONEq(`[ [ "test", true ], [ "test", false ] ]`, string(js))
	}
}

/*
func TestAllAttributes(t *testing.T) {
	tobj := []byte(`{
		"Kind": "Some",
		"kubegen.Object.Select(.test.foo)": {
		  "kubegen.Object.LoadJSON": "fixture1"
		}
	}`)
}
*/

/*
func TestMacroSelect(t *testing.T) {
	conv := New()

	assert := assert.New(t)

	tfiles := map[string][]byte{
		"fixture1": []byte(`{ "test": { "foo": "bar" } }`),
	}

	tobj := []byte(`{
		"Kind": "Some",
		"kubegen.Object.Select(.test.foo)": {
		  "kubegen.Object.LoadJSON": "fixture1"
		}
	}`)

	conv.DefineMacro(LoadObjectJSON,
		func(c *Converter, branch *BranchLocator, _ *Macro) (ModifierCallback, error) {
			var newData []byte
			if v, ok := tfiles[string(branch.value)]; ok {
				newData = v
			} else {
				newData = []byte("{ }")
			}
			return MakeObjectLoadJSON(c, branch, newData)
		})

	conv.DefineMacro(LoadArrayJSON,
		func(c *Converter, branch *BranchLocator, _ *Macro) (ModifierCallback, error) {
			var newData []byte
			if v, ok := tfiles[string(branch.value)]; ok {
				newData = v
			} else {
				newData = []byte("[ ]")
			}
			return MakeArrayLoadJSON(c, branch, newData)
		})

	objectSelect := &Macro{
		ReturnType: Array,
		EvalPhase:  MacrosEvalPhaseA,
		VerbName:   "LoadJSON",
		Argument:   true,
	}

	conv.DefineMacro(objectSelect,
		func(c *Converter, branch *BranchLocator, _ *Macro) (ModifierCallback, error) {
			return nil, nil
		})

	if err := conv.LoadObject(tobj, "tobj1.json", ""); err != nil {
		t.Fatalf("failed to load – %v", err)
	}

	if err := conv.Run(); err != nil {
		t.Logf("tree=%s", conv.tree)
		t.Fatalf("failed to convert – %v", err)
	}

	{
		v, err := conv.tree.GetString("Kind")
		assert.Nil(err)
		assert.Equal("Some", v)
	}
}
*/

// TODO:
// - kubegen.String.ReadFile
// - kubegen.Object.LoadYAML
// Also:
// - kubegen.Array.ReadBytes
// - kubegen.String.AsBASE64
