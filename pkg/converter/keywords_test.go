package converter

import (
	"fmt"
	"testing"

	"github.com/buger/jsonparser"

	"github.com/stretchr/testify/assert"
)

func TestKeywordModifiersDeletion(t *testing.T) {
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
		"object without modifier keywords should be different")
	assert.True(len(tobj) > len(umobj),
		"object without modifier keywords should be larger")

	conv.DefineKeyword(&Keyword{
		ReturnType: jsonparser.Null,
		EvalPhase:  KeywordEvalPhaseA,
		FuncName:   "Delete",
	}, func(c *Converter, branch *BranchInfo) error {
		p := branch.PathToString()
		switch branch.kind {
		case jsonparser.String:
			fallthrough
		case jsonparser.Object:
			// TODO panic if key exists or find a way to have unique keys
			c.modifiers[p] = func(c *Converter) error {
				c.data = jsonparser.Delete(c.data, branch.path[1:]...)
				return nil
			}
		}
		return nil
	})

	if err := conv.loadStrict(tobj); err != nil {
		t.Fatalf("failed to laod – %v\nc.data=%s", err, string(conv.data))
	}

	if err := conv.run(KeywordEvalPhaseA); err != nil {
		t.Fatalf("failed to convert – %v\nc.data=%s\nc.tree.self=%#v", err, string(conv.data), conv.tree.self)
	}

	if err := conv.callModifiersOnce(); err != nil {
		t.Fatalf("failed to run modifiers – %v", err)
	}

	unmodified := New()
	if err := unmodified.loadStrict(umobj); err != nil {
		t.Fatalf("failed to laod – %v\nc.data=%s", err, string(unmodified.data))
	}

	assert.Equal(len(unmodified.data), len(conv.data),
		"new object should be the same len as one without modifier keywords")

	assert.Equal(unmodified.data, conv.data,
		"new object should be exactly the same as one without modifier keywords")

	assert.Equal(0, len(unmodified.keywords[KeywordEvalPhaseA]),
		"object without modifier keywords has no keyword handlers")
	assert.Equal(0, len(unmodified.modifiers),
		"object without modifier keywords has no keyword callbacks")

	reloaded := New()
	if err := unmodified.loadStrict(conv.data); err != nil {
		t.Fatalf("failed to laod – %v\nc.data=%s", err, string(reloaded.data))
	}
}

func TestKeywordErrorsAndModifiersLookup(t *testing.T) {
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

	conv.DefineKeyword(KeywordStringLookup,
		func(c *Converter, branch *BranchInfo) error {
			p := branch.PathToString()
			switch branch.kind {
			case jsonparser.String:
				c.modifiers[p] = func(c *Converter) error {
					v, err := jsonparser.Set(c.data, []byte("\"TEST\""), branch.parent.path[1:]...)
					if err != nil {
						return fmt.Errorf("could not set string – %v", err)
					}
					c.data = v
					return nil
				}
			default:
				return fmt.Errorf("in %q value is a %s, but must be a string", p, branch.kind)
			}
			return nil
		})

	conv.DefineKeyword(KeywordNumberLookup,
		func(c *Converter, branch *BranchInfo) error {
			p := branch.PathToString()
			switch branch.kind {
			case jsonparser.String:
				c.modifiers[p] = func(c *Converter) error {
					v, err := jsonparser.Set(c.data, []byte("0"), branch.parent.path[1:]...)
					if err != nil {
						return fmt.Errorf("could not set number – %v", err)
					}
					c.data = v
					return nil
				}
			default:
				return fmt.Errorf("in %q value is a %s, but must be a string", p, branch.kind)
			}
			return nil
		})

	for _, v := range badSyntax {
		conv2 := New()
		conv2.keywords = conv.keywords
		var err error
		err = conv2.loadStrict(v)
		assert.NotNil(err)
	}

	for _, v := range badModfiersOrObjecs {
		conv2 := New()
		conv2.keywords = conv.keywords
		var err error
		err = conv2.loadStrict(v)
		assert.Nil(err)
		err = conv2.run(KeywordEvalPhaseB)
		assert.NotNil(err)
	}

	if err := conv.loadStrict(tobj); err != nil {
		t.Fatalf("failed to laod – %v\nc.data=%s", err, string(conv.data))
	}

	if err := conv.run(KeywordEvalPhaseB); err != nil {
		t.Fatalf("failed to convert – %v\nc.data=%s\nc.tree.self=%#v", err, string(conv.data), conv.tree.self)
	}

	if err := conv.callModifiersOnce(); err != nil {
		t.Fatalf("failed to run modifiers – %v", err)
	}

	{
		v, err := jsonparser.GetString(conv.data, "test1s")
		assert.Nil(err)
		assert.Equal("TEST", v)
	}

	{
		v, err := jsonparser.GetString(conv.data, "test2s")
		assert.Nil(err)
		assert.Equal("TEST", v)
	}

	{
		v, err := jsonparser.GetInt(conv.data, "test3n")
		assert.Nil(err)
		assert.Equal(int64(0), v)
	}

	{
		v, err := jsonparser.GetInt(conv.data, "test4n")
		assert.Nil(err)
		assert.Equal(int64(0), v)
	}
}

func TestKeywordLookupRecursive(t *testing.T) {
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
		}
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
		"testInsertArray1": []byte(`[{ "kubegen.Number.Lookup": "test3val" },{ "kubegen.String.Lookup": "test3val" },{ "kubegen.Object.Lookup": "testInsertObj2" }]`),
		// "testInsertObj4": []byte(`{}`),
	}

	conv.DefineKeyword(KeywordStringLookup,
		func(c *Converter, branch *BranchInfo) error {
			p := branch.PathToString()

			switch branch.kind {
			case jsonparser.String:
				c.modifiers[p] = func(c *Converter) error {
					v, err := jsonparser.Set(c.data, []byte("\"TEST_STRING\""), branch.parent.path[1:]...)
					if err != nil {
						return fmt.Errorf("could not set string – %v", err)
					}
					c.data = v
					return nil
				}
			default:
				return fmt.Errorf("in %q value is a %s, but must be a string", p, branch.kind)
			}
			return nil
		})

	conv.DefineKeyword(KeywordNumberLookup,
		func(c *Converter, branch *BranchInfo) error {
			p := branch.PathToString()

			switch branch.kind {
			case jsonparser.String:

				c.modifiers[p] = func(c *Converter) error {
					v, err := jsonparser.Set(c.data, []byte("12345"), branch.parent.path[1:]...)
					if err != nil {
						return fmt.Errorf("could not set number – %v", err)
					}
					c.data = v
					return nil
				}
			default:
				return fmt.Errorf("in %q value is a %s, but must be a string", p, branch.kind)
			}
			return nil
		})

	conv.DefineKeyword(KeywordObjectLookup,
		func(c *Converter, branch *BranchInfo) error {
			p := branch.PathToString()
			var x []byte
			if v, ok := objs[string(branch.value)]; ok {
				x = v
			} else {
				x = []byte("{ }")
			}
			switch branch.kind {
			case jsonparser.String:
				c.modifiers[p] = func(c *Converter) error {
					v, err := jsonparser.Set(c.data, x, branch.parent.path[1:]...)
					if err != nil {
						return fmt.Errorf("could not set object – %v", err)
					}
					c.data = v
					return nil
				}
			default:
				return fmt.Errorf("in %q value is a %s, but must be a string", p, branch.kind)
			}
			return nil
		})

	conv.DefineKeyword(KeywordArrayLookup,
		func(c *Converter, branch *BranchInfo) error {
			p := branch.PathToString()
			var x []byte
			if v, ok := objs[string(branch.value)]; ok {
				x = v
			} else {
				x = []byte("[ 1, 2, 3 ]")
			}
			switch branch.kind {
			case jsonparser.String:
				c.modifiers[p] = func(c *Converter) error {
					v, err := jsonparser.Set(c.data, x, branch.parent.path[1:]...)
					if err != nil {
						return fmt.Errorf("could not set array – %v", err)
					}
					c.data = v
					return nil
				}
			default:
				return fmt.Errorf("in %q value is a %s, but must be a string", p, branch.kind)
			}
			return nil
		})

	if err := conv.loadStrict(tobj); err != nil {
		t.Fatalf("failed to laod – %v\nc.data=%s", err, string(conv.data))
	}

	if err := conv.Run(); err != nil {
		t.Fatalf("failed to run converter – %v", err)
	}

	{
		v, err := jsonparser.GetString(conv.data, "test1s")
		assert.Nil(err)
		assert.Equal("TEST_STRING", v)
	}

	{
		v, err := jsonparser.GetInt(conv.data, "test2n")
		assert.Nil(err)
		assert.Equal(int64(12345), v)
	}

	{
		v, t, _, err := jsonparser.Get(conv.data, "test3m")
		a := []byte(fmt.Sprintf(`[12345,"TEST_STRING",%s]`, objs["testInsertObj2"]))
		assert.Nil(err)
		assert.Equal(jsonparser.Array, t)
		assert.Equal(a, v)
	}

	{
		v, t, _, err := jsonparser.Get(conv.data, "test5o")
		a := []byte("{ }")
		assert.Nil(err)
		assert.Equal(jsonparser.Object, t)
		assert.Equal(a, v)
	}

	{
		v, err := jsonparser.GetString(conv.data, "test1s")
		assert.Nil(err)
		assert.Equal("TEST_STRING", v)
	}

	{
		v, err := jsonparser.GetInt(conv.data, "test2n")
		assert.Nil(err)
		assert.Equal(int64(12345), v)
	}

	{
		v, t, _, err := jsonparser.Get(conv.data, "test3m")
		a := []byte(fmt.Sprintf(`[12345,"TEST_STRING",%s]`, objs["testInsertObj2"]))
		assert.Nil(err)
		assert.Equal(jsonparser.Array, t)
		assert.Equal(a, v)
	}

	{
		_, t, _, err := jsonparser.Get(conv.data, "test4o")
		assert.Nil(err)
		assert.Equal(jsonparser.Object, t)
	}

	{
		v, err := jsonparser.GetInt(conv.data, "test4o", "foo", "[0]")
		assert.Nil(err)
		assert.Equal(int64(12345), v)
	}

	{
		v, err := jsonparser.GetString(conv.data, "test4o", "foo", "[1]")
		assert.Nil(err)
		assert.Equal("TEST_STRING", v)
	}

	{
		v, t, _, err := jsonparser.Get(conv.data, "test4o", "foo", "[2]")
		assert.Nil(err)
		assert.Equal(jsonparser.Object, t)
		assert.Equal(v, objs["testInsertObj2"])
	}

	{
		v, t, _, err := jsonparser.Get(conv.data, "test5o")
		a := []byte("{ }")
		assert.Nil(err)
		assert.Equal(jsonparser.Object, t)
		assert.Equal(a, v)
	}

	{
		v, t, _, err := jsonparser.Get(conv.data, "test4o", "bar", "[0]")
		a := []byte("{ }")
		assert.Nil(err)
		assert.Equal(jsonparser.Object, t)
		assert.Equal(a, v)
	}

	{
		v, err := jsonparser.GetInt(conv.data, "test4o", "foo", "[0]")
		assert.Nil(err)
		assert.Equal(int64(12345), v)
	}

	{
		v, err := jsonparser.GetString(conv.data, "test4o", "foo", "[1]")
		assert.Nil(err)
		assert.Equal("TEST_STRING", v)
	}

	{
		v, t, _, err := jsonparser.Get(conv.data, "test4o", "foo", "[2]")
		assert.Nil(err)
		assert.Equal(jsonparser.Object, t)
		assert.Equal(v, objs["testInsertObj2"])
	}

	{
		v, t, _, err := jsonparser.Get(conv.data, "test5o")
		a := []byte("{ }")
		assert.Nil(err)
		assert.Equal(jsonparser.Object, t)
		assert.Equal(a, v)
	}

	{
		v, t, _, err := jsonparser.Get(conv.data, "test4o", "bar", "[0]")
		a := []byte("{ }")
		assert.Nil(err)
		assert.Equal(jsonparser.Object, t)
		assert.Equal(a, v)
	}
}

func TestKeywordJoinStrings(t *testing.T) {
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

	if err := conv.LoadObj(tobj1, "tobj1.json", ""); err != nil {
		t.Fatalf("failed to load – %v", err)
	}

	conv.DefineKeyword(KeywordStringJoin, StringJoin)

	if err := conv.Run(); err != nil {
		t.Logf("c.data=%s", string(conv.data))
		t.Fatalf("failed to convert – %v", err)
	}

	assert.Equal(0, len(conv.modifiers))

	{
		v, t, _, err := jsonparser.Get(conv.data, "foobar")
		a := []byte("foobar")
		assert.Nil(err)
		assert.Equal(jsonparser.String, t)
		assert.Equal(a, v)
	}
}

func TestKeywordObjectToJSON(t *testing.T) {
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

	if err := conv.LoadObj(tobj1, "tobj1.json", ""); err != nil {
		t.Fatalf("failed to load – %v", err)
	}

	conv.DefineKeyword(KeywordStringAsJSON, StringAsJSON)
	conv.DefineKeyword(KeywordStringAsYAML, StringAsYAML)

	if err := conv.Run(); err != nil {
		t.Logf("c.data=%s", string(conv.data))
		t.Fatalf("failed to convert – %v", err)
	}

	assert.Equal(0, len(conv.modifiers))

	{
		v, t, _, err := jsonparser.Get(conv.data, "foobar1")
		a := []byte(`[\"foo\",\"bar\"]`)
		assert.Nil(err)
		assert.Equal(jsonparser.String, t)
		assert.Equal(a, v)
	}

	{
		v, t, _, err := jsonparser.Get(conv.data, "foobar2")
		a := []byte(`- foo\n- bar\n`)
		assert.Nil(err)
		assert.Equal(jsonparser.String, t)
		assert.Equal(a, v)
	}

	{
		v, t, _, err := jsonparser.Get(conv.data, "foobar3")
		a := []byte(`bar: {}\nfoo: []\n`)
		assert.Nil(err)
		assert.Equal(jsonparser.String, t)
		assert.Equal(a, v)
	}
}

func TestKeywordToString(t *testing.T) {
	assert := assert.New(t)

	kw := &Keyword{
		ReturnType: String,
		EvalPhase:  KeywordEvalPhaseA,
		FuncName:   "FooBar",
	}

	assert.Equal("kubegen.String.FooBar", kw.String())
}

/*
func TestKeywordLoadJSON(t *testing.T) {
	conv := New()

	//assert := assert.New(t)

	loadObject := &Keyword{
		ReturnType: Object,
		EvalPhase:  KeywordEvalPhaseA,
		FuncName:   "LoadJSON",
	}

	loadArray := &Keyword{
		ReturnType: Array,
		EvalPhase:  KeywordEvalPhaseA,
		FuncName:   "LoadJSON",
	}

	tfiles := map[string][]byte{
		"file1.json": []byte(`{ "test": true }`),
	}

	tobj := []byte(`{
		"Kind": "Some",
		"kubegen.Object.LoadJSON": "file1.json"
	}`)

	// TODO parse and re-marshal, check if we get desired type

	conv.DefineKeyword(loadObject, func(c *Converter, branch *BranchInfo) error {
		p := branch.PathToString()

		switch branch.kind {
		case jsonparser.String:

			c.modifiers[p] = func(c *Converter) error {
				v, err := jsonparser.Set(c.data, tfiles[string(branch.Value())], branch.parent.path[0:]...)
				if err != nil {
					return fmt.Errorf("could not set object – %v", err)
				}
				c.data = v
				return nil
			}
		default:
			return fmt.Errorf("in %q value is a %s, but must be a string", p, branch.kind)
		}
		return nil
	})

	conv.DefineKeyword(loadArray, func(c *Converter, branch *BranchInfo) error {
		p := branch.PathToString()

		switch branch.kind {
		case jsonparser.String:

			c.modifiers[p] = func(c *Converter) error {
				v, err := jsonparser.Set(c.data, tfiles[string(branch.Value())], branch.parent.path[1:]...)
				if err != nil {
					return fmt.Errorf("could not set object – %v", err)
				}
				c.data = v
				return nil
			}
		default:
			return fmt.Errorf("in %q value is a %s, but must be a string", p, branch.kind)
		}
		return nil
	})
	if err := conv.LoadObj(tobj, "tobj1.json", ""); err != nil {
		t.Fatalf("failed to load – %v", err)
	}
	if err := conv.Run(); err != nil {
		t.Logf("c.data=%s", string(conv.data))
		t.Fatalf("failed to convert – %v", err)
	}
}
*/

// TODO:
// - kubegen.String.ReadFile
// - kubegen.Object.LoadJSON
// - kubegen.Object.LoadYAML
// Also:
// - kubegen.Array.ReadBytes
// - kubegen.String.AsBASE64
