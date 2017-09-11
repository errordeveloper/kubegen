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
	mapKey reflect.Value    // hash key to use for the current value (if it is in a hash)
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
	return reflect.ValueOf(nil)
}

func (c *Converter) appendSimpleValue(v reflect.Value) {
	if c.isRoot {
		c.values = append(c.values, v)
		return
	}
	c.values = append(c.values, v) // keep this for consistency
	c.thisBranch().value = &v
}

func (c *Converter) appendBranch(newBranch converterBranch) {
	if c.isRoot {
		c.isRoot = false
	} else {
		newBranch.parent = c.thisBranch()
	}
	c.values = append(c.values, newBranch.self)
	c.branches = append(c.branches, newBranch)
	i := len(c.branches) - 1
	c.branchIndex, newBranch.index = i, i
}

func (c *Converter) thisBranch() *converterBranch {
	return &c.branches[c.branchIndex]
}

func (c *Converter) flipBranch() {
	if c.thisBranch().parent != nil {
		// fmt.Printf("c.branchIndex = (%d => %d)\n", c.branchIndex, c.thisBranch().parent.index)
		c.branchIndex = c.thisBranch().parent.index
		c.thisBranch().value = nil // reset the value, so we don't inherit it unitentionally
	}
}

func (c *Converter) walkTree(v interface{}) {
	c.appendSimpleValue(reflect.ValueOf(v))
}

func (c *Converter) newMapBranch() converterBranch {
	newBranch := converterBranch{}
	newBranch.self = reflect.ValueOf(make(map[string]interface{}))
	c.appendBranch(newBranch)
	if newBranch.parent != nil {
		newBranch.parent.self.SetMapIndex(newBranch.parent.mapKey, newBranch.self)
	}
	return newBranch
}

func (c *Converter) convertMapBranch(x map[string]interface{}) {
	thisBranch := c.newMapBranch()
	for k, v := range x {
		thisBranch.mapKey = reflect.ValueOf(k)
		c.walkTree(v)
		if thisBranch.value != nil {
			thisBranch.self.SetMapIndex(thisBranch.mapKey, *thisBranch.value)
		}
	}
	c.flipBranch()
}

func (c *Converter) newSliceBranch() converterBranch {
	newBranch := converterBranch{}
	newBranch.self = reflect.ValueOf(make([]interface{}, 0))
	c.appendBranch(newBranch)
	if newBranch.parent != nil {
		newBranch.parent.self.SetMapIndex(newBranch.parent.mapKey, newBranch.self)
	}
	return newBranch
}

func (c *Converter) convertSliceBranch(x []interface{}) {
	thisBranch := c.newSliceBranch()
	for k, v := range x {
		thisBranch.mapKey = reflect.ValueOf(fmt.Sprintf("%d", k))
		c.walkTree(v)
		if thisBranch.value != nil {
			thisBranch.self.SetMapIndex(thisBranch.mapKey, *thisBranch.value)
		}
	}
	c.flipBranch()
}
