package converter

import (
	"encoding/json"
	"fmt"
	"reflect"

	_ "github.com/buger/jsonparser"
)

type converterBranch struct {
	self   *reflect.Value
	parent *converterBranch // pointer to the parent, so we know where to go back once done
	key    reflect.Value    // hash key to use for the current value (if it is in a hash)
	value  *reflect.Value   // current value we are handling
	index  int              // our index in the tree
}

type Converter struct {
	branches    []*converterBranch // list of tree branches – hashes and arrays
	branchIndex int                // maps to the current position in the tree
	values      []*reflect.Value   // used to store all values we find
	isRoot      bool               // idiates we are at the root of the tree
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
		if c.values[0] == nil {
			return reflect.ValueOf(nil)
		}
		return *c.values[0]
	}
	panic("Value: uninitialised converter cannot be used to obtain a value!")
}

func (c *Converter) Dump() {
	for k, v := range c.values {
		if v != nil {
			fmt.Printf("value %d: %#v\n", k, *v)
		}
	}
	for k, v := range c.branches {
		fmt.Printf("branch %d: %#v\n", k, v)
		if v != nil {
			if v.self != nil {
				fmt.Printf("branch %d: (itself) %#v\n", k, *v.self)
			}
			if v.parent != nil {
				if v.parent.self != nil {
					fmt.Printf("branch %d: (parent) %#v\n", k, *v.parent.self)
				}
			}
		}
	}
}

func (c *Converter) walkTree(v interface{}) {
	fmt.Printf("walkTree: v=%#v\n", v)
	switch vv := v.(type) {
	case map[string]interface{}:
		//fmt.Printf("walkTree: mapBranch=%#v\n", vv)
		c.convertMapBranch(vv)
	case []interface{}:
		//fmt.Printf("walkTree: sliceBranch=%#v\n", vv)
		c.convertSliceBranch(vv)
	default:
		x := reflect.ValueOf(v)
		//fmt.Printf("walkTree: x.Kind=%s\n", x.Kind())
		c.appendSimpleValue(&x)
	}
}

func (c *Converter) appendSimpleValue(v *reflect.Value) {
	if c.isRoot {
		c.values = append(c.values, v)
		return
	}
	c.values = append(c.values, v) // keep this for consistency
	c.curentBranch().value = v
}

func (c *Converter) curentBranch() *converterBranch {
	return c.branches[c.branchIndex]
}

func (c *Converter) flipBranch() {
	if c.curentBranch().parent != nil {
		fmt.Printf("c.branchIndex = (%d => %d)\n", c.branchIndex, c.curentBranch().parent.index)
		c.branchIndex = c.curentBranch().parent.index
		c.curentBranch().value = nil // reset the value, so we don't inherit it unitentionally
	}
}

func (b *converterBranch) setValueInParent(v reflect.Value) {
	fmt.Printf("setValueInParent: parent index %d\n", b.parent.index)
	fmt.Printf("setValueInParent: v=%#v\n", v)
	var n reflect.Value
	switch b.parent.self.Kind() {
	case reflect.Map:
		b.parent.self.SetMapIndex(b.parent.key, v)
		n = b.parent.self.MapIndex(b.parent.key)
		b.self = &n
		fmt.Printf("setValueInParent: set %q to %#v in %#v [branch index %d]...\n", b.parent.key, *b.self, *b.parent.self, b.parent.index)
	case reflect.Interface:
		switch b.parent.self.Elem().Kind() {
		case reflect.Slice:
			n = b.parent.self.Elem()
			b.parent.self = &n
		default:
			fmt.Sprintf("setValueInParent: unknown interface kind of parent %q (%#v)!\n", b.parent.self.Elem().Kind(), *b.parent.self)
		}
		fallthrough
	case reflect.Slice:
		switch v.Kind() {
		case reflect.Interface:
			fmt.Println("OMG!\n")
			n = reflect.Append(*b.parent.self, v)
		default:
			n = reflect.Append(*b.parent.self, v)
		}
		b.parent.self = &n
		x := n.Index(n.Len() - 1)
		b.self = &x
		fmt.Printf("setValueInParent: appended %q to parent slice (%#v) [length %d] [branch index %d]...\n", *b.self, *b.parent.self, n.Len(), b.parent.index)
	default:
		panic(fmt.Sprintf("setValueInParent: unknown kind of parent %q (%#v)!\n", b.parent.self.Kind(), *b.parent.self))
	}
}

func (c *Converter) appendNewBranch(v reflect.Value) *converterBranch {
	newBranch := converterBranch{parent: nil}
	if c.isRoot {
		c.isRoot = false
		fmt.Println("isRoot!")
	} else {
		newBranch.parent = c.curentBranch()
		fmt.Println("thisBranch!")
	}
	if newBranch.parent != nil {
		fmt.Printf("appendNewBranch: parent=%#v\n", newBranch.parent)
		if newBranch.parent.self != nil {
			newBranch.setValueInParent(v)
		}
	} else {
		newBranch.self = &v
		newBranch.parent = &newBranch
	}
	c.values = append(c.values, newBranch.self)
	c.branches = append(c.branches, &newBranch)
	i := len(c.branches) - 1
	fmt.Printf("appendNewBranch: new index %d\n", i)
	c.branchIndex, newBranch.index = i, i
	return &newBranch
}

func (c *Converter) convertMapBranch(x map[string]interface{}) {
	thisBranch := c.appendNewBranch(reflect.ValueOf(make(map[string]interface{})))
	for k, v := range x {
		thisBranch.key = reflect.ValueOf(k)
		c.walkTree(v)
		if thisBranch.value != nil && thisBranch.parent != nil && thisBranch.parent.self != nil {
			fmt.Printf("convertMapBranch: thisBranch.value=%#v\n", *thisBranch.value)
			thisBranch.setValueInParent(*thisBranch.value)
			//} else if thisBranch.self != nil && thisBranch.value != nil {
			//	fmt.Println("NEVER?")
			//	fmt.Printf("convertMapBranch: thisBranch.value=%#v\n", *thisBranch.value)
			//	switch thisBranch.self.Kind() {
			//	case reflect.Map:
			//		thisBranch.self.SetMapIndex(thisBranch.key, *thisBranch.value)
			//		fmt.Printf("convertMapBranch: set %q to %#v in parent (%#v)...\n", thisBranch.key, *thisBranch.value, *thisBranch.self)
			//	default:
			//		panic(fmt.Sprintf("convertMapBranch: unknown kind of parent %q (%#v)!\n", thisBranch.self.Kind(), *thisBranch.self))
			//	}
		} else {
			fmt.Printf("convertMapBranch: unknown code path thisBranch=%#v\n", thisBranch)
			fmt.Printf("convertMapBranch: unknown code path thisBranch.key=%#v\n", thisBranch.key)
			if thisBranch.value != nil {
				fmt.Printf("convertMapBranch: unknown code path thisBranch.value=%#v\n", *thisBranch.value)
			}
			if thisBranch.parent == nil {
				panic("PARENT IS NIL")
			}
		}
	}
	c.flipBranch()
}

func (c *Converter) convertSliceBranch(x []interface{}) {
	thisBranch := c.appendNewBranch(reflect.ValueOf(make([]interface{}, 0)))
	for k, v := range x {
		thisBranch.key = reflect.ValueOf(k)
		c.walkTree(v)
		if thisBranch.value != nil && thisBranch.self != nil && thisBranch.parent != nil && thisBranch.parent.self != nil {
			fmt.Printf("convertSliceBranch: thisBranch=%#v\n", *thisBranch.value)
			thisBranch.setValueInParent(*thisBranch.value)
		} else {
			fmt.Printf("convertSliceBranch: self=%#v value=%#v\n", thisBranch.self, thisBranch.value)
		}
	}
	c.flipBranch()
}
