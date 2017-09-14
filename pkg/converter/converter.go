package converter

// TODO try using jsonparser, assume that root is a map
// build a nested map without values to track the location
// try to write some tests first
// TODO also consider https://godoc.org/github.com/json-iterator/go
// it should have similar methods, it'd better cause it's what
// Kubernetes uses, and Tim has contributed a ton of tests etc :)

import (
	"encoding/json"
	"fmt"

	"github.com/buger/jsonparser"
)

type branchInfo struct {
	kind   jsonparser.ValueType
	self   branch
	value  *[]byte
	parent *branch
}

type branch = map[string]branchInfo

type Converter struct {
	tree branchInfo
	data []byte
}

func New() *Converter {
	return &Converter{}
}

func (c *Converter) Load(data []byte) {
	c.data = data
}

// LoadStrict ensure we validate the data, as out jsonparser doesn't produce very helpful errors,
// namely it doesn't indicate line numbers and seems to tolerate leading commas etc
func (c *Converter) LoadStrict(data []byte) error {
	obj := new(interface{})
	if err := json.Unmarshal(data, obj); err != nil {
		return fmt.Errorf(
			"error while re-decoding %q (%q): %v",
			"<TODO:instanceName>", "<TODO:sourcePath", err)
	}
	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf(
			"error while re-encoding %q (%q): %v",
			"<TODO:instanceName>", "<TODO:sourcePath>", err)
	}
	c.data = data
	return nil
}

func (c *Converter) makeIterator(parentBranch *branchInfo, key string, value []byte, dataType jsonparser.ValueType, index *int) {
	newBranch := branchInfo{
		parent: &parentBranch.self,
		kind:   dataType,
		value:  &value,
		self:   make(branch),
	}
	*index = *index + 1
	parentBranch.self[key] = newBranch

	switch dataType {
	case jsonparser.Object:
		handler := c.makeObjectIterator(&newBranch)

		if err := jsonparser.ObjectEach(value, handler); err != nil {
			panic(err)
		}
	case jsonparser.Array:
		handler := c.makeArrayIterator(&newBranch)

		jsonparser.ArrayEach(value, handler)
	}
}

type objectIterator func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error

func (c *Converter) makeObjectIterator(parentBranch *branchInfo) objectIterator {
	index := 0
	callback := func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		c.makeIterator(parentBranch, string(key), value, dataType, &index)
		return nil
	}
	return callback
}

type arrayIterator func(value []byte, dataType jsonparser.ValueType, offset int, err error)

func (c *Converter) makeArrayIterator(parentBranch *branchInfo) arrayIterator {
	index := 0
	callback := func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		key := fmt.Sprintf("[[%d]]", index)
		c.makeIterator(parentBranch, key, value, dataType, &index)
	}
	return callback
}

func (c *Converter) Run() error {
	kind, err := jsonparser.GetString(c.data, "Kind")
	if err != nil {
		return err
	}
	if kind == "" {
		return fmt.Errorf("kind is bank")
	}

	c.tree = branchInfo{
		parent: nil,
		kind:   jsonparser.Object,
		value:  &c.data,
		self:   make(branch),
	}

	if err := jsonparser.ObjectEach(c.data, c.makeObjectIterator(&c.tree)); err != nil {
		return err
	}
	return nil
}
