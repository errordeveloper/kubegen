package macroproc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPathCutoffIndex(t *testing.T) {
	assert := assert.New(t)

	tobj := []byte(`{
		"a1": ["0", 1, 2, 3],
		"a2": [10, 20, 30],
		"o": {
			"a3": ["100", "200", "300"],
			"a4": [null, true]
		}
	}`)

	o, err := loadObject(tobj)
	if err != nil {
		t.Fatal(err)
	}

	{
		i := o.getPathCutoffIndex("x")
		assert.Equal(-1, i)
	}

	{
		i := o.getPathCutoffIndex("x", 1, 2)
		assert.Equal(-1, i)
	}

	{
		i := o.getPathCutoffIndex("xxx", 123, 456, 789, "x", "x", "x")
		assert.Equal(-1, i)
	}

	{
		i := o.getPathCutoffIndex("a2", 0, "xxx", 123, 456, 789, "x", "x", "x")
		assert.Equal(1, i)
	}

	{
		i := o.getPathCutoffIndex(0, "x")
		assert.Equal(-1, i)
	}

	{
		i := o.getPathCutoffIndex("a1")
		assert.Equal(0, i)
	}

	{
		i := o.getPathCutoffIndex("a1", 0)
		assert.Equal(1, i)
	}

	{
		i := o.getPathCutoffIndex("o", "a3", 0)
		assert.Equal(2, i)
	}

	{
		i := o.getPathCutoffIndex("o", "a3", 10)
		assert.Equal(1, i)
	}

	{
		i := o.getPathCutoffIndex("o", "a3", 0, 10)
		assert.Equal(2, i)
	}

	{
		i := o.getPathCutoffIndex("o", "a3", 0, "a2", 0, "xxx", 123, 456, 789, "x", "x", "x")
		assert.Equal(2, i)
	}

}

func TestTreeOverlayBasic(t *testing.T) {
	assert := assert.New(t)

	tobj := []byte(`{
		"Kind": "Some",
		"that":  {},
		"things": [
			{ "a": "one", "b": "two", "c": "three" }
		],
		"test": {
			"kubegen.Object.Lookup": "foo"
		}
	}`)

	sobj := []byte(`{
		"a": "one1",
		"x": "twoX",
		"z": "threeZ",
		"y": { "foo": 1 },
		"n": [ 0, { "a": "one"}, 2, { "c": "three"}, 4, {} ]
	}`)

	target, err := loadObject(tobj)
	if err != nil {
		t.Fatal(err)
	}

	source, err := loadObject(sobj)
	if err != nil {
		t.Fatal(err)
	}

	{
		err := target.Overlay(source, "things", 0)
		assert.Nil(err)
	}

	{
		v, err := target.GetString("things", 0, "a")
		assert.Nil(err)
		assert.Equal("one", v)
	}

	{
		v, err := target.GetString("things", 0, "b")
		assert.Nil(err)
		assert.Equal("two", v)
	}

	{
		v, err := target.GetString("things", 0, "c")
		assert.Nil(err)
		assert.Equal("three", v)
	}

	{
		v, err := target.GetString("things", 0, "x")
		assert.Nil(err)
		assert.Equal("twoX", v)
	}

	{
		v, err := target.GetString("things", 0, "z")
		assert.Nil(err)
		assert.Equal("threeZ", v)
	}

	{
		err := target.Overlay(source, "that")
		assert.Nil(err)

		v, err := target.GetObject("that")
		assert.Nil(err)

		js, err := json.Marshal(v)
		assert.Nil(err)

		assert.JSONEq(string(sobj), string(js))
	}

	{
		err := target.Overlay(source, "test")
		assert.Nil(err)

		err = target.Delete("test", "kubegen.Object.Lookup")
		assert.Nil(err)

		v, err := target.GetObject("test")
		assert.Nil(err)

		js, err := json.Marshal(v)
		assert.Nil(err)

		assert.JSONEq(string(sobj), string(js))
	}

}

func TestTreeOverlayWithArraysAndObjects(t *testing.T) {
	assert := assert.New(t)

	tobj := []byte(`{
		"Kind": "Some",
		"test1s": [
			{
				"kubegen.Object.Lookup": "testInsertObj6"
			},
			{
				"testObj": {},
				"testStr": "rts"
			},
			{
				"testObj": { "test0": 0 }
			},
			{
				"testObj": { "test0": 0, "test2": "2x2" }
			},
			{
				"testObj": { "test3": { "bar": "foo" } }
			},
			{
				"testObj": { "test3": { "foo": "baz" } },
				"testAry": []
			}
		]
	}`)

	sobj := []byte(`{ "testObj": { "test1": 1, "test2": 2, "test3": { "foo": "bar" } }, "testStr": "str", "testAry": [ { "name": "eman" } ] }`)

	target, err := loadObject(tobj)
	if err != nil {
		t.Fatal(err)
	}

	source, err := loadObject(sobj)
	if err != nil {
		t.Fatal(err)
	}

	{
		{
			err := target.Overlay(source, "test1s", 0)
			if err != nil {
				t.Fatal(err)
			}
		}

		{
			v, err := target.GetString("test1s", 0, "kubegen.Object.Lookup")
			assert.Nil(err)
			assert.Equal("testInsertObj6", v)
		}

		{
			v, err := target.GetFloat("test1s", 0, "testObj", "test1")
			assert.Nil(err)
			assert.Equal(float64(1), v)
		}

		{
			v, err := target.GetFloat("test1s", 0, "testObj", "test2")
			assert.Nil(err)
			assert.Equal(float64(2), v)
		}

		{
			v, err := target.GetString("test1s", 0, "testStr")
			assert.Nil(err)
			assert.Equal("str", v)
		}
	}

	{
		{
			err := target.Overlay(source, "test1s", 1)
			if err != nil {
				t.Fatal(err)
			}
		}

		{
			v, err := target.GetString("test1s", 1, "testStr")
			assert.Nil(err)
			assert.Equal("rts", v)
		}

		{
			v, err := target.GetObject("test1s", 1, "testObj")
			assert.Nil(err)

			js, err := json.Marshal(v)
			assert.Nil(err)

			assert.JSONEq(`{ "test1": 1, "test2": 2, "test3": { "foo": "bar" } }`, string(js))
		}

		{
			v, err := target.GetObject("test1s", 1, "testAry", 0)
			assert.Nil(err)

			js, err := json.Marshal(v)
			assert.Nil(err)

			assert.JSONEq(`{ "name": "eman" }`, string(js))
		}
	}

	{
		{
			err := target.Overlay(source, "test1s", 2)
			if err != nil {
				t.Fatal(err)
			}
		}

		{
			v, err := target.GetString("test1s", 2, "testStr")
			assert.Nil(err)
			assert.Equal("str", v)
		}

		{
			v, err := target.GetObject("test1s", 2, "testObj")
			assert.Nil(err)

			js, err := json.Marshal(v)
			assert.Nil(err)

			assert.JSONEq(`{ "test0": 0, "test1": 1, "test2": 2, "test3": { "foo": "bar" } }`, string(js))
		}
	}

	{
		{
			err := target.Overlay(source, "test1s", 3)
			if err != nil {
				t.Fatal(err)
			}
		}

		{
			v, err := target.GetString("test1s", 3, "testStr")
			assert.Nil(err)
			assert.Equal("str", v)
		}

		{
			v, err := target.GetObject("test1s", 3, "testObj")
			assert.Nil(err)

			js, err := json.Marshal(v)
			assert.Nil(err)

			assert.JSONEq(`{ "test0": 0, "test1": 1, "test2": "2x2", "test3": { "foo": "bar" } }`, string(js))
		}
	}

	{
		{
			err := target.Overlay(source, "test1s", 4)
			if err != nil {
				t.Fatal(err)
			}
		}

		{
			v, err := target.GetString("test1s", 4, "testStr")
			assert.Nil(err)
			assert.Equal("str", v)
		}

		{
			v, err := target.GetObject("test1s", 4, "testObj")
			assert.Nil(err)

			js, err := json.Marshal(v)
			assert.Nil(err)

			assert.JSONEq(`{"test1": 1, "test2":2, "test3": { "bar": "foo", "foo": "bar" } }`, string(js))
		}
	}

	{
		{
			err := target.Overlay(source, "test1s", 5)
			if err != nil {
				t.Fatal(err)
			}
		}

		{
			v, err := target.GetObject("test1s", 5, "testObj")
			assert.Nil(err)

			js, err := json.Marshal(v)
			assert.Nil(err)

			assert.JSONEq(`{"test1": 1, "test2":2, "test3": { "foo": "baz" } }`, string(js))
		}

		{
			v, err := target.GetArray("test1s", 5, "testAry")
			assert.Nil(err)

			js, err := json.Marshal(v)
			assert.Nil(err)

			assert.JSONEq(`[ { "name": "eman" } ]`, string(js))
		}
	}
}

func TestTreeOverlayWithObjectOnly(t *testing.T) {
	assert := assert.New(t)

	tobj := []byte(`{
		"Kind": "Some",
		"test1s": {
			"0": {
				"kubegen.Object.Lookup": "testInsertObj6"
			},
			"1": {
				"testObj": {},
				"testStr": "rts"
			},
			"2": {
				"testObj": { "test0": 0 }
			},
			"3": {
				"testObj": { "test0": 0, "test2": "2x2" }
			},
			"4": {
				"testObj": { "test3": { "bar": "foo" } }
			}
		}
	}`)

	sobj := []byte(`{ "testObj": { "test1": 1, "test2": 2, "test3": { "foo": "bar" } }, "testStr": "str" }`)

	target, err := loadObject(tobj)
	if err != nil {
		t.Fatal(err)
	}

	source, err := loadObject(sobj)
	if err != nil {
		t.Fatal(err)
	}

	{
		{
			err := target.Overlay(source, "test1s", "0")
			assert.Nil(err)
		}

		{
			v, err := target.GetString("test1s", "0", "kubegen.Object.Lookup")
			assert.Nil(err)
			assert.Equal("testInsertObj6", v)
		}

		{
			v, err := target.GetFloat("test1s", "0", "testObj", "test1")
			assert.Nil(err)
			assert.Equal(float64(1), v)
		}

		{
			v, err := target.GetFloat("test1s", "0", "testObj", "test2")
			assert.Nil(err)
			assert.Equal(float64(2), v)
		}

		{
			v, err := target.GetString("test1s", "0", "testStr")
			assert.Nil(err)
			assert.Equal("str", v)
		}
	}

	{
		{
			err := target.Overlay(source, "test1s", "1")
			assert.Nil(err)
		}

		{
			v, err := target.GetString("test1s", "1", "testStr")
			assert.Nil(err)
			assert.Equal("rts", v)
		}

		{
			v, err := target.GetObject("test1s", "1", "testObj")
			assert.Nil(err)

			js, err := json.Marshal(v)
			assert.Nil(err)

			assert.JSONEq(`{ "test1": 1, "test2": 2, "test3": { "foo": "bar" } }`, string(js))
		}
	}

	{
		{
			err := target.Overlay(source, "test1s", "2")
			assert.Nil(err)
		}

		{
			v, err := target.GetString("test1s", "2", "testStr")
			assert.Nil(err)
			assert.Equal("str", v)
		}

		{
			v, err := target.GetObject("test1s", "2", "testObj")
			assert.Nil(err)

			js, err := json.Marshal(v)
			assert.Nil(err)

			assert.JSONEq(`{ "test0": 0, "test1": 1, "test2": 2, "test3": { "foo": "bar" } }`, string(js))
		}
	}

	{
		{
			err := target.Overlay(source, "test1s", "3")
			assert.Nil(err)
		}

		{
			v, err := target.GetString("test1s", "3", "testStr")
			assert.Nil(err)
			assert.Equal("str", v)
		}

		{
			v, err := target.GetObject("test1s", "3", "testObj")
			assert.Nil(err)

			js, err := json.Marshal(v)
			assert.Nil(err)

			assert.JSONEq(`{ "test0": 0, "test1": 1, "test2": "2x2", "test3": {"foo":"bar"} }`, string(js))
		}
	}

	{
		{
			err := target.Overlay(source, "test1s", "4")
			assert.Nil(err)
		}

		{
			v, err := target.GetString("test1s", "4", "testStr")
			assert.Nil(err)
			assert.Equal("str", v)
		}

		{
			v, err := target.GetObject("test1s", "4", "testObj")
			assert.Nil(err)

			js, err := json.Marshal(v)
			assert.Nil(err)

			assert.JSONEq(`{"test1": 1, "test2": 2, "test3": { "bar": "foo", "foo": "bar"  }}`, string(js))
		}
	}
}
