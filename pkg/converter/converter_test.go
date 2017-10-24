package converter

import (
	"fmt"
	"strings"
	"testing"

	"github.com/buger/jsonparser"

	_ "github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func assertPathKeys(pks [][]string, conv *Converter, t *testing.T) {
	assert := assert.New(t)

	for _, pk := range pks {
		x := conv.get(pk...)
		if assert.NotNil(x, fmt.Sprintf("path %v should exist", pk)) {
			assert.Equal(len(pk)+1, len(x.path),
				fmt.Sprintf("lenght of path to %v should be len(pk)+1", pk))
			assert.Equal("."+strings.Join(pk, "."), strings.Join(x.path, "."),
				fmt.Sprintf("dot-joined path to %v should be the same", pk))
		}
	}
}

func TestConverterBasic(t *testing.T) {

	assert := assert.New(t)

	tobj := []byte(`{
		"Kind": "Some",		
		"this":  true,
		"that":  false,
		"things": [
			{ "a": 1, "b": 2, "c": 3 }
		],
		"nothing": { "empty1": [], "empty2": [] },
		"other": {
			"moreThings": [
				{ "a": 1, "b": 2, "c": 3 },
				{ "a": 1, "b": 2, "c": 3 },
				{ "a": 1, "b": 2, "c": 3 }
			],
			"number": 1.0,
			"string": "foobar"
		},
		"and more": {
			"Kind": "Some",
			"this":  true,
			"that":  false,
			"things": [
				{ "a": 1, "b": 2, "c": 3 }
			],
			"nothing": { "empty1": [], "empty2": [] },
			"other": {
				"moreThings": [
					{ "a": 1, "b": 2, "c": 3 },
					{ "a": 1, "b": 2, "c": 3 },
					{ "a": 1, "b": 2, "c": 3 }
				],
				"number": 1.0,
				"string": "foobar"
			}
		}
	}`)

	conv := New()

	if err := conv.loadStrict(tobj); err != nil {
		t.Fatalf("failed to laod – %v", err)
	}

	if err := conv.run(KeywordEvalPhaseA); err != nil {
		t.Fatalf("failed to covert – %v", err)
	}

	assert.Equal(7, len(conv.tree.self),
		"converter should have a tree of length 7")
	assert.Equal(6, len(conv.tree.self["and more"].self),
		"converter should have `and more` subtree of length 6")

	//t.Log(spew.Sdump(conv.tree))

	assert.Equal(jsonparser.Object, conv.get("other", "moreThings", "[0]").kind)
	assert.Equal(jsonparser.Object, conv.get("other", "moreThings", "[1]").kind)
	assert.Equal(jsonparser.Object, conv.get("other", "moreThings", "[2]").kind)
	assert.Nil(conv.get("other", "moreThings", "[9]"))
}

func TestConverterOnlyObjects(t *testing.T) {

	assert := assert.New(t)

	tobj := []byte(`{
			"Kind": "Some",		
			"this":  true,
			"that":  false,
			"things": { "a": 1, "b": 2, "c": 3 },
			"other": {
				"moreThings": {
					"1": { "a": 1, "b": 2, "c": 3 },
					"2": { "a": 1, "b": 2, "c": 3 },
					"3": { "a": 1, "b": 2, "c": 3 }
				},
				"number": 1.0,
				"string": "foobar"
			},
			"and more": {
				"Kind": "Some",
				"this":  true,
				"that":  false,
				"things": {
					"0": { "a": 1, "b": 2, "c": 3 }
				},
				"other": {
					"moreThings": {
						"1": { "a": 1, "b": 2, "c": 3 },
						"2": { "a": 1, "b": 2, "c": 3 },
						"3": { "a": 1, "b": 2, "c": 3 }
					},
					"number": 1.0,
					"string": "foobar"
				}
			}
		}`)

	conv := New()

	if err := conv.loadStrict(tobj); err != nil {
		t.Fatalf("failed to laod – %v", err)
	}

	if err := conv.run(KeywordEvalPhaseA); err != nil {
		t.Fatalf("failed to covert – %v", err)
	}

	assert.Equal(6, len(conv.tree.self),
		"converter should have a tree of length 7")
	assert.Equal(5, len(conv.tree.self["and more"].self),
		"converter should have `and more` subtree of length 6")

	//t.Log(spew.Sdump(conv.tree))

	assert.Equal(jsonparser.Object, conv.get("other", "moreThings", "1").kind)
	assert.Equal(jsonparser.Object, conv.get("other", "moreThings", "2").kind)
	assert.Equal(jsonparser.Object, conv.get("other", "moreThings", "3").kind)
	assert.Nil(conv.get("other", "moreThings", "0"))

	pathKeys := [][]string{
		{"things"},
		{"things", "a"},
		{"things", "b"},
		{"things", "c"},
		{"other", "number"},
		{"other", "string"},
		{"other", "moreThings"},
		{"other", "moreThings", "1"},
		{"other", "moreThings", "1", "a"},
		{"other", "moreThings", "1", "b"},
		{"other", "moreThings", "1", "c"},
		{"other", "moreThings", "2"},
		{"other", "moreThings", "2", "a"},
		{"other", "moreThings", "2", "b"},
		{"other", "moreThings", "2", "c"},
		{"other", "moreThings", "3"},
		{"other", "moreThings", "3", "a"},
		{"other", "moreThings", "3", "b"},
		{"other", "moreThings", "3", "c"},
		{"and more", "Kind"},
		{"and more", "this"},
		{"and more", "that"},
		{"and more", "other", "number"},
		{"and more", "other", "string"},
	}

	assertPathKeys(pathKeys, conv, t)

}

func TestBasicKubegenAsset(t *testing.T) {

	assert := assert.New(t)

	tobj := []byte(`{
		"Kind": "kubegen.k8s.io/Module.v1alpha2",
		"Deployments": [
			{
				"name": "cart",
				"replicas": 1,
				"containers": [
					{
						"name": "cart",
						"image": "<image_registry>/cart:0.4.0",
						"ports": [
							{
								"name": "http",
								"containerPort": 80
							}
						],
						"securityContext": {
							"runAsNonRoot": true,
							"runAsUser": 10001,
							"capabilities": {
								"drop": [
									"all"
								],
								"add": [
									"NET_BIND_SERVICE"
								]
							},
							"readOnlyRootFilesystem": true
						},
						"volumeMounts": [
							{
								"mountPath": "/tmp",
								"name": "tmp-volume"
							}
						],
						"livenessProbe": {
							"httpGet": {
								"path": "/health"
							},
							"initialDelaySeconds": 300,
							"periodSeconds": 3
						},
						"readinessProbe": {
							"httpGet": {
								"path": "/health"
							},
							"initialDelaySeconds": 180,
							"periodSeconds": 3
						}
					}
				],
				"volumes": [
					{
						"name": "tmp-volume",
						"emptyDir": {
							"medium": "Memory"
						}
					}
				]
			},
			{
				"name": "cart-db",
				"kubegen.MapMerge.Lookup": "mongo",
				"replicas": 2
			}
		],
		"Services": [
			{
				"name": "cart",
				"annotations": {
					"prometheus.io/path": "/prometheus"
				},
				"ports": [
					{
						"name": "http"
					}
				]
			},
			{
				"name": "cart-db",
				"ports": [
					{
						"name": "mongo"
					}
				]
			}
		]
	}`)

	conv := New()

	if err := conv.loadStrict(tobj); err != nil {
		t.Fatalf("failed to laod – %v", err)
	}

	if err := conv.run(KeywordEvalPhaseA); err != nil {
		t.Fatalf("failed to covert – %v", err)
	}

	//t.Log(spew.Sdump(conv.tree))

	assert.Equal(2, len(conv.tree.self["Deployments"].self),
		"there are two Deployments")
	assert.Equal(jsonparser.String, conv.tree.self["Deployments"].self["[0]"].self["name"].kind,
		"there should be name in a Deployments")
	assert.Equal(jsonparser.String, conv.tree.self["Deployments"].self["[1]"].self["name"].kind,
		"there should be name in a Deployments")

	assert.Equal(2, len(conv.tree.self["Services"].self),
		"there are two Services")
	assert.Equal(jsonparser.String, conv.tree.self["Services"].self["[0]"].self["name"].kind,
		"there should be cart in Services")
	assert.Equal(jsonparser.String, conv.tree.self["Services"].self["[1]"].self["name"].kind,
		"there should be cart in Services")

	pathKeys := [][]string{
		{"Deployments"},

		{"Deployments", "[0]"},
		{"Deployments", "[0]", "name"},
		{"Deployments", "[0]", "replicas"},

		{"Deployments", "[0]", "containers"},
		{"Deployments", "[0]", "containers", "[0]"},
		{"Deployments", "[0]", "volumes"},
		{"Deployments", "[0]", "volumes", "[0]"},

		{"Deployments", "[1]"},
		{"Deployments", "[1]", "name"},
		{"Deployments", "[1]", "replicas"},

		{"Services"},

		{"Services", "[0]", "ports"},
		{"Services", "[0]", "ports", "[0]", "name"},
		{"Services", "[0]", "annotations"},
		{"Services", "[0]", "annotations", "prometheus.io/path"},

		{"Services", "[1]", "ports"},
		{"Services", "[1]", "ports", "[0]", "name"},
	}

	assertPathKeys(pathKeys, conv, t)

	//t.Log(string(conv.data))

}

func TestConverterGet(t *testing.T) {
	conv := New()
	tobj := []byte(`{ "Kind": "some" }`)

	if err := conv.loadStrict(tobj); err != nil {
		t.Fatalf("failed to laod – %v", err)
	}

	if err := conv.run(KeywordEvalPhaseA); err != nil {
		t.Fatalf("failed to covert – %v", err)
	}

	errors := make(chan error)
	go func() {
		for _, x := range []string{"1", "2", "3"} {
			conv.doIterate(&conv.tree, x, tobj, jsonparser.Object, errors)
		}

		conv.doIterate(&conv.tree, "order", []byte(`{}`), jsonparser.Object, errors)

		conv.doIterate(conv.get("order"), "potatoe",
			[]byte(`{ "mash": { "count": 1 }, "chips": { "count": 2 }, "sausages": true, "gravy": { "beef": 1, "chicken": 0 }  }`),
			jsonparser.Object, errors)

		errors <- nil
	}()

	if err := <-errors; err != nil {
		t.Fatalf("%v", err)
	}

	pathKeys := [][]string{
		{"1", "Kind"},
		{"2", "Kind"},
		{"3", "Kind"},
		{"order", "potatoe"},
		{"order", "potatoe", "mash"},
		{"order", "potatoe", "chips"},
	}

	assertPathKeys(pathKeys, conv, t)
}

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
		err = conv2.run(KeywordEvalPhaseA)
		assert.NotNil(err)
	}

	if err := conv.loadStrict(tobj); err != nil {
		t.Fatalf("failed to laod – %v\nc.data=%s", err, string(conv.data))
	}

	if err := conv.run(KeywordEvalPhaseA); err != nil {
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
		"test3m": [
			{ "kubegen.Number.Lookup": "test3val" },
			{ "kubegen.String.Lookup": "test3val" },
			{ "kubegen.Object.Lookup": "testInsertObj2" }
		],
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
		"testInsertObj2": []byte(`{ "bar": "TEST_VAL4" }`),
		"testInsertObj3": []byte(`{ "kubegen.Object.Lookup": "testInsertObj4" }`),
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

	if err := conv.loadStrict(tobj); err != nil {
		t.Fatalf("failed to laod – %v\nc.data=%s", err, string(conv.data))
	}

	// first pass

	if err := conv.run(KeywordEvalPhaseA); err != nil {
		t.Fatalf("failed to convert – %v\nc.data=%s\nc.tree.self=%#v", err, string(conv.data), conv.tree.self)
	}

	if err := conv.callModifiersOnce(); err != nil {
		t.Fatalf("failed to run modifiers – %v", err)
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
		v, t, _, err := jsonparser.Get(conv.data, "test4o")
		a := []byte(objs["testInsertObj1"])
		assert.Nil(err)
		assert.Equal(jsonparser.Object, t)
		assert.Equal(a, v)
	}

	{
		v, t, _, err := jsonparser.Get(conv.data, "test5o")
		a := []byte("{ }")
		assert.Nil(err)
		assert.Equal(jsonparser.Object, t)
		assert.Equal(a, v)
	}

	assert.Equal(0, len(conv.modifiers))

	// second pass

	if err := conv.run(KeywordEvalPhaseA); err != nil {
		t.Fatalf("failed to convert – %v\nc.data=%s\nc.tree.self=%#v", err, string(conv.data), conv.tree.self)
	}

	assert.Equal(4, len(conv.modifiers))

	if err := conv.callModifiersOnce(); err != nil {
		t.Fatalf("failed to run modifiers – %v", err)
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
		v, t, _, err := jsonparser.Get(conv.data, "test4o", "bar", "[0]")
		assert.Nil(err)
		assert.Equal(jsonparser.Object, t)
		assert.Equal(v, objs["testInsertObj3"])
	}

	{
		v, t, _, err := jsonparser.Get(conv.data, "test5o")
		a := []byte("{ }")
		assert.Nil(err)
		assert.Equal(jsonparser.Object, t)
		assert.Equal(a, v)
	}

	assert.Equal(0, len(conv.modifiers))

	// third pass

	if err := conv.run(KeywordEvalPhaseA); err != nil {
		t.Fatalf("failed to convert – %v\nc.data=%s\nc.tree.self=%#v", err, string(conv.data), conv.tree.self)
	}

	assert.Equal(1, len(conv.modifiers))

	if err := conv.callModifiersOnce(); err != nil {
		t.Fatalf("failed to run modifiers – %v", err)
	}

	{
		v, t, _, err := jsonparser.Get(conv.data, "test4o", "bar", "[0]")
		a := []byte("{ }")
		assert.Nil(err)
		assert.Equal(jsonparser.Object, t)
		assert.Equal(a, v)
	}

	assert.Equal(0, len(conv.modifiers))

	// final pass

	if err := conv.run(KeywordEvalPhaseA); err != nil {
		t.Fatalf("failed to convert – %v\nc.data=%s\nc.tree.self=%#v", err, string(conv.data), conv.tree.self)
	}

	assert.Equal(0, len(conv.modifiers))

	// now do the same with exported run method
	if err := conv.loadStrict(tobj); err != nil {
		t.Fatalf("failed to laod – %v\nc.data=%s", err, string(conv.data))
	}

	if err := conv.Run(); err != nil {
		t.Fatalf("failed to convert – %v\nc.data=%s\nc.tree.self=%#v", err, string(conv.data), conv.tree.self)
	}

	assert.Equal(0, len(conv.modifiers))

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

// TODO:
// - kubegen.String.ReadFile
// - kubegen.Object.LoadJSON
// - kubegen.Object.LoadYAML
// Also:
// - kubegen.Array.ReadBytes
// - kubegen.String.AsBASE64
