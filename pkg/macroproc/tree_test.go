package macroproc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTreeGet(t *testing.T) {
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

	tree, err := loadObject(tobj)
	if err != nil {
		t.Fatal(err)
	}

	{
		v, err := tree.GetValue("Kind")
		assert.Nil(err)
		assert.Equal("Some", v)
	}

	{
		v, err := tree.GetBoolean("this")
		assert.Nil(err)
		assert.Equal(true, v)
	}

	{
		v, err := tree.GetBoolean("that")
		assert.Nil(err)
		assert.Equal(false, v)
	}

	{
		v, err := tree.GetFloat("things", 0, "a")
		assert.Nil(err)
		assert.Equal(float64(1), v)
	}

	{
		v, err := tree.GetFloat("things", 0, "b")
		assert.Nil(err)
		assert.Equal(float64(2), v)
	}

	{
		v, err := tree.GetFloat("things", 0, "c")
		assert.Nil(err)
		assert.Equal(float64(3), v)
	}

	err = tree.ObjectEach(func(key string, value interface{}, valueType ValueType) error {
		switch key {
		case "Kind":
			assert.Equal(String, valueType)
			assert.Equal("Some", value)
		case "this":
			assert.Equal(Boolean, valueType)
			assert.Equal(true, value)
		case "that":
			assert.Equal(Boolean, valueType)
			assert.Equal(false, value)
		case "things":
			assert.Equal(Array, valueType)
			v := `[{ "a": 1, "b": 2, "c": 3 }]`
			x, _ := json.Marshal(value)
			assert.JSONEq(v, string(x))
		case "nothing":
			assert.Equal(Object, valueType)
			v := `{ "empty1": [], "empty2": [] }`
			x, _ := json.Marshal(value)
			assert.JSONEq(v, string(x))
		case "other":
			assert.Equal(Object, valueType)
			v := `{
					"moreThings": [
						{ "a": 1, "b": 2, "c": 3 },
						{ "a": 1, "b": 2, "c": 3 },
						{ "a": 1, "b": 2, "c": 3 }
					],
					"number": 1.0,
					"string": "foobar"
				}`
			x, _ := json.Marshal(value)
			assert.JSONEq(v, string(x))

			{
				tree := &Tree{}
				tree.self = value.(map[string]interface{})["moreThings"]
				err := tree.ArrayEach(func(_ int, value interface{}, valueType ValueType) error {
					assert.Equal(Object, valueType)
					v := `{ "a": 1, "b": 2, "c": 3 }`
					x, _ := json.Marshal(value)
					assert.JSONEq(v, string(x))
					return nil
				})
				if err != nil {
					t.Fatal(err)
				}
			}

		case "and more":
			assert.Equal(Object, valueType)
			v := `{
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
				}`
			x, err := json.Marshal(value)
			assert.Nil(err)
			assert.JSONEq(v, string(x))

			{

				tree := &Tree{}
				tree.self = value.(map[string]interface{})["nothing"]
				err := tree.ArrayEach(func(_ int, value interface{}, valueType ValueType) error {
					return nil
				})
				assert.NotNil(err)
			}

			{
				tree := &Tree{}
				tree.self = value.(map[string]interface{})["nothing"].(map[string]interface{})["empty1"]
				err := tree.ArrayEach(func(_ int, value interface{}, valueType ValueType) error {
					t.Fatal("iterrator callback shouldn't get called")
					return nil
				})
				assert.Nil(err)
			}

			{
				tree := &Tree{}
				tree.self = value.(map[string]interface{})["nothing"].(map[string]interface{})["empty1"]
				err := tree.ArrayEach(func(_ int, value interface{}, valueType ValueType) error {
					t.Fatal("iterrator callback shouldn't get called")
					return nil
				})
				assert.Nil(err)
			}

		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestTreeSet(t *testing.T) {
	assert := assert.New(t)

	tobj := []byte(`{
		"Kind": "Some",
		"EmptyArray": [],
		"SomeArray": ["foo", 1],
		"EmptyObject": {},
		"SomeObject": {"foo": 1, "anotherObject": {"SomeString": "foo1", "SomeArray": [{"bar": 2}]}}
	}`)

	tree, err := loadObject(tobj)
	if err != nil {
		t.Fatal(err)
	}

	{
		err := tree.SetString("OKAY", "Kind")
		assert.Nil(err)
		v, err := tree.GetValue("Kind")
		assert.Nil(err)
		assert.Equal("OKAY", v)
	}

	{
		v, err := tree.GetArray("EmptyArray")
		assert.Nil(err)
		assert.Equal(0, len(v))

		x := make([]interface{}, 3)
		x[0] = 123.0
		x[1] = 345
		x[2] = "789"

		err = tree.SetArray(x, "EmptyArray")
		assert.Nil(err)

		v, err = tree.GetArray("EmptyArray")
		assert.Nil(err)
		assert.Equal(3, len(v))
		assert.Equal(123.0, v[0])
		assert.Equal(345, v[1])
		assert.Equal("789", v[2])
	}

	{
		err = tree.SetInt(2, "SomeObject", "foo")
		assert.Nil(err)
	}

	{
		v, err := tree.GetString("SomeObject", "anotherObject", "SomeString")
		assert.Nil(err)
		assert.Equal("foo1", v)
	}

	{
		err = tree.Set(false, "SomeObject", "anotherObject", "SomeString")
		assert.Nil(err)
	}

	{
		v, err := tree.GetBoolean("SomeObject", "anotherObject", "SomeString")
		assert.Nil(err)
		assert.Equal(false, v)
	}

	{
		v, err := tree.GetArray("SomeObject", "anotherObject", "SomeArray")
		assert.Nil(err)
		x := map[string]bool{
			"true":  true,
			"false": false,
		}
		v = append(v, x)
		err = tree.SetArray(v, "SomeObject", "anotherObject", "SomeArray")
		assert.Nil(err)
	}

	{
		v, err := tree.GetObject("SomeObject", "anotherObject")
		assert.Nil(err)
		v["moreStuff"] = map[string]string{
			"EmptyString": "",
		}
		err = tree.SetObject(v, "SomeObject", "anotherObject")
		assert.Nil(err)
	}

	result := `{
		"EmptyObject":{},
		"Kind":"OKAY",
		"SomeArray":["foo", 1],
		"SomeObject":{
			"anotherObject":{
				"SomeArray":[
					{"bar":2},
					{"false":false, "true":true}
				],
				"SomeString":false,
				"moreStuff":{"EmptyString":""}
			},
			"foo":2
		},
		"EmptyArray":[123, 345, "789"]
	}`

	js, err := json.Marshal(tree.self)
	assert.Nil(err)
	assert.JSONEq(result, string(js))
}

func TestTreeDelete(t *testing.T) {
	assert := assert.New(t)

	tobj := []byte(`{
		"Kind": "Some",
		"EmptyArray": [],
		"SomeArray": ["foo", 1, "bar", 2],
		"EmptyObject": {},
		"SomeObject": {"foo": 1, "anotherObject": {"SomeString": "foo1", "SomeArray": [{"bar": 2}]}}
	}`)

	tree, err := loadObject(tobj)
	if err != nil {
		t.Fatal(err)
	}

	{
		err := tree.Delete("Kind")
		assert.Nil(err)
		_, err = tree.GetValue("Kind")
		assert.NotNil(err)
	}

	{
		err := tree.Delete("SomeObject", "foo")
		assert.Nil(err)
	}

	{
		err := tree.SetString("FOO", "SomeArray", 0)
		assert.Nil(err)
	}

	{
		err := tree.Delete("SomeArray", 1)
		assert.Nil(err)
	}

	{
		err := tree.Delete("EmptyArray")
		assert.Nil(err)
	}

	{
		err := tree.Delete("EmptyObject")
		assert.Nil(err)
	}

	{
		err := tree.Delete("SomeObject", "anotherObject", "SomeArray", 0, "bar")
		assert.Nil(err)
	}

	{
		err := tree.Delete("SomeObject", "anotherObject", "SomeArray", 0)
		assert.Nil(err)
	}

	result := `{
		"SomeArray": ["FOO", "bar", 2],
		"SomeObject": {"anotherObject": {"SomeString": "foo1", "SomeArray": []}}
	}`

	js, err := json.Marshal(tree.self)
	assert.Nil(err)
	assert.JSONEq(result, string(js))
}

func TestTreeDeleteArrayElements(t *testing.T) {
	assert := assert.New(t)

	tobj := []byte(`{
		"a1": ["0", 1, 2, 3],
		"a2": [10, 20, 30],
		"o": {
			"a3": ["100", "200", "300"],
			"a4": [null, true]
		}
	}`)

	tree, err := loadObject(tobj)
	if err != nil {
		t.Fatal(err)
	}

	{
		err = tree.Delete("a1", 1)
		assert.Nil(err)

		v, err := tree.GetFloat("a1", 1)
		assert.Nil(err)

		assert.Equal(float64(2), v)
	}

	result := `{
		"a1": ["0", 2, 3],
		"a2": [10, 20, 30],
		"o": {
			"a3": ["100", "200", "300"],
			"a4": [null, true]
		}
	}`

	js, err := json.Marshal(tree.self)
	assert.Nil(err)
	assert.JSONEq(result, string(js))

}
