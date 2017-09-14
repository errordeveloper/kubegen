package converter

// TODO try using jsonparser, assume that root is a map
// build a nested map without values to track the location
// try to write some tests first

import (
	"encoding/json"
	"fmt"

	"github.com/buger/jsonparser"
)

type branchInfo struct {
	selfType jsonparser.ValueType
	selfInfo branchData
	myParent *branchData
}

type branchData = map[string]branchInfo

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

type objectIterator func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error

func (c *Converter) makeObjectIterator(parentBranch *branchInfo) objectIterator {
	callback := func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		newBranch := branchInfo{
			myParent: &parentBranch.selfInfo,
			selfType: dataType,
			selfInfo: make(branchData),
		}

		parentBranch.selfInfo[string(key)] = newBranch

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
		return nil
	}
	return callback
}

type arrayIterator func(value []byte, dataType jsonparser.ValueType, offset int, err error)

func (c *Converter) makeArrayIterator(parentBranch *branchInfo) arrayIterator {
	callback := func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		newBranch := branchInfo{
			myParent: &parentBranch.selfInfo,
			selfType: dataType,
			selfInfo: make(branchData),
		}

		parentBranch.selfInfo[fmt.Sprintf("[[%d]]", offset)] = newBranch

		switch dataType {
		case jsonparser.Object:
			if err := jsonparser.ObjectEach(value, c.makeObjectIterator(&newBranch)); err != nil {
				panic(err)
			}
		case jsonparser.Array:
			jsonparser.ArrayEach(value, c.makeArrayIterator(&newBranch))
		}
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
		myParent: nil,
		selfType: jsonparser.Object,
		selfInfo: make(branchData),
	}

	if err := jsonparser.ObjectEach(c.data, c.makeObjectIterator(&c.tree)); err != nil {
		return err
	}
	return nil
}

func (c *Converter) Dumps() []string {
	s := []string{}
	for k, v := range c.tree.selfInfo {
		s = append(s, fmt.Sprintf("k=%#v v=%#v\n", k, v))
	}
	return s
}

func (c *Converter) helperLoader(obj interface{}) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	c.data = data
	return nil
}
