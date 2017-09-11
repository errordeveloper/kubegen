package converter

import (
	"encoding/json"
	"fmt"
	"reflect"

	_ "github.com/buger/jsonparser"
)

type converterBranch struct {
	self   reflect.Value
	parent *converterBranch // pointer to the parent, so we know where to go back once done
	key    reflect.Value    // hash key to use for the current value (if it is in a hash)
	value  *reflect.Value   // current value we are handling
	index  int              // our index in the tree
}

type Converter struct {
	branches    []converterBranch // list of tree branches – hashes and arrays
	branchIndex int               // maps to the current position in the tree
	values      []reflect.Value   // used to store all values we find
	isRoot      bool              // idiates we are at the root of the tree
}

func New() *Converter {
	return &Converter{isRoot: true}
}

func (c *Converter) Convert(obj interface{}) error {
	if len(c.values) > 0 || len(c.branches) > 0 || !c.isRoot {
		return fmt.Errorf("Convert: don't call me again, I'm stupid!")
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	// Re-encode JSON data into an interface and store it
	var tree interface{}
	if err := json.Unmarshal(data, &tree); err != nil {
		return err
	}

	c.walkTree(tree)
	return nil
}

func (c *Converter) Value() reflect.Value {
	if len(c.values) > 0 {
		return c.values[0]
	}
	panic("Value: uninitialised converter cannot be used to obtain a value!")
}

func (c *Converter) Dump() {
	for k, v := range c.values {
		fmt.Printf("value %d: %#v\n", k, v)
	}
	for k, v := range c.branches {
		fmt.Printf("branch %d: (itself) %#v\n", k, v.self)
		if v.parent != nil {
			fmt.Printf("branch %d: (parent) %#v\n", k, v.parent.self)
		}
	}
}

func (c *Converter) appendSimpleValue(v reflect.Value) {
	if c.isRoot {
		c.values = append(c.values, v)
		return
	}
	c.values = append(c.values, v) // keep this for consistency
	c.thisBranch().value = &v
}

func (c *Converter) appendBranch(newBranch *converterBranch) {
	if c.isRoot {
		c.isRoot = false
	} else {
		newBranch.parent = c.thisBranch()
	}
	if newBranch.parent != nil {
		fmt.Printf("appendBranch: parent index %d\n", newBranch.parent.index)
		switch newBranch.parent.self.Kind() {
		case reflect.Map:
			if newBranch.parent.key != reflect.ValueOf(nil) {
				fmt.Println("have parent key!")
				newBranch.parent.self.SetMapIndex(newBranch.parent.key, newBranch.self)
				fmt.Printf("appendBranch: set %q to %#v in %#v [branch index %d]...\n", newBranch.parent.key, newBranch.self, newBranch.parent.self, newBranch.parent.index)
			}
		case reflect.Slice:
			newBranch.parent.self = reflect.Append(newBranch.parent.self, newBranch.self)
			fmt.Printf("appendBranch: appended %q to parent slice (%#v) [branch index %d]...\n", newBranch.self, newBranch.parent.self, newBranch.parent.index)
		default:
			panic("appendBranch: unknown kind of parent!")
		}
	}
	c.values = append(c.values, newBranch.self)
	c.branches = append(c.branches, *newBranch)
	i := len(c.branches) - 1
	fmt.Printf("appendBranch: new index %d\n", i)
	c.branchIndex, newBranch.index = i, i
}

func (c *Converter) thisBranch() *converterBranch {
	return &c.branches[c.branchIndex]
}

func (c *Converter) flipBranch() {
	if c.thisBranch().parent != nil {
		fmt.Printf("c.branchIndex = (%d => %d)\n", c.branchIndex, c.thisBranch().parent.index)
		c.branchIndex = c.thisBranch().parent.index
		c.thisBranch().value = nil // reset the value, so we don't inherit it unitentionally
	}
}

func (c *Converter) walkTree(v interface{}) {
	fmt.Printf("v=%#v\n", v)
	switch vv := v.(type) {
	case map[string]interface{}:
		fmt.Printf("mapBranch=%#v\n", vv)
		c.convertMapBranch(vv)
	case []interface{}:
		fmt.Printf("sliceBranch=%#v\n", vv)
		c.convertSliceBranch(vv)
	default:
		c.appendSimpleValue(reflect.ValueOf(v))
	}
}

func (c *Converter) newMapBranch() *converterBranch {
	newBranch := &converterBranch{}
	newBranch.self = reflect.ValueOf(make(map[string]interface{}))
	c.appendBranch(newBranch)
	return newBranch
}

func (c *Converter) convertMapBranch(x map[string]interface{}) {
	thisBranch := c.newMapBranch()
	for k, v := range x {
		thisBranch.key = reflect.ValueOf(k)
		c.walkTree(v)
		if thisBranch.value != nil {
			fmt.Printf("convertMapBranch: thisBranch.value=%#v\n", *thisBranch.value)
			switch thisBranch.self.Kind() {
			case reflect.Map:
				thisBranch.self.SetMapIndex(thisBranch.key, *thisBranch.value)
				fmt.Printf("convertMapBranch: set %q to %#v in (%#v)...\n", thisBranch.key, *thisBranch.value, thisBranch.self)
			default:
				panic(fmt.Sprintf("convertMapBranch: unknown kind %s (%v)!\n", thisBranch.value.Kind(), *thisBranch.value))
			}
		}
	}
	c.flipBranch()
}

func (c *Converter) newSliceBranch() *converterBranch {
	newBranch := &converterBranch{}
	newBranch.self = reflect.ValueOf(make([]interface{}, 0))
	c.appendBranch(newBranch)
	return newBranch
}

func (c *Converter) convertSliceBranch(x []interface{}) {
	thisBranch := c.newSliceBranch()
	for k, v := range x {
		thisBranch.key = reflect.ValueOf(k)
		c.walkTree(v)
		if thisBranch.value != nil {
			fmt.Printf("convertSliceBranch: thisBranch=%#v\n", *thisBranch.value)
			if thisBranch.self.Kind() == reflect.Slice {
				thisBranch.self = reflect.Append(thisBranch.self, *thisBranch.value)
				fmt.Printf("convertSliceBranch: appended %q to slice (%#v)...\n", *thisBranch.value, thisBranch.self)
			} else {
				panic(fmt.Sprintf("convertSliceBranch: unknown kind %s (%v)!\n", thisBranch.value.Kind(), *thisBranch.value))
			}
		}
	}
	c.flipBranch()
}
