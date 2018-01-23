package macroproc

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/errordeveloper/kubegen/pkg/util"
)

// Tree holds arbitrary JSON object and allows
// for walking and manipulating it (as needed
// for this package)
type Tree struct {
	mutex    sync.Mutex
	self     interface{}
	key      interface{}
	root     *Tree
	next     *Tree
	parent   *Tree
	setValue func(newValue interface{})
	delete   func()
}

func loadObject(data []byte) (*Tree, error) {
	x := new(interface{})
	if err := util.LoadObj(x, data, "_.json", "_"); err != nil {
		return nil, err
	}
	t := NewTree(x)
	t.root = t
	return t, nil
}

// NewTree creates a tree root
func NewTree(x *interface{}) *Tree {
	tree := &Tree{
		self:   *x,
		parent: nil,
		next:   nil,
	}
	tree.setValue = func(newValue interface{}) {
		tree.self = newValue
	}
	return tree
}

func (t *Tree) newTree() *Tree {
	//log.Printf("root@%p=%s", t.root, t.root)
	return &Tree{parent: t, root: t.root}
}

func (t *Tree) rootLock() { t.root.mutex.Lock() }

func (t *Tree) rootUnlock() { t.root.mutex.Unlock() }

// Len returns lenght of Object or Array of
// a tree branch
func (t *Tree) Len() int {
	return len(t.self.(map[string]interface{}))
}

// Bytes returns JSON-encoded representation
// it is mostly used for testing, so it will
// panic if the object cannot be encoded
func (t *Tree) Bytes() []byte {
	js, err := json.Marshal(t.self)
	if err != nil {
		panic(err) // this mostly for testing, so we are okay
	}
	return js
}

// String returns JSON-encoded representation
// it is mostly used for testing, so it will
// panic if the object cannot be encoded
func (t *Tree) String() string { return string(t.Bytes()) }

func (t *Tree) makeNext(k, v interface{}) (*ValueType, error) {
	var vt ValueType
	switch v.(type) {
	case map[string]interface{}:
		vt = Object
	case []interface{}:
		vt = Array
	case string:
		vt = String
	case float32:
		vt = Number
	case float64:
		vt = Number
	case int:
		vt = Number
	case int16:
		vt = Number
	case int32:
		vt = Number
	case int64:
		vt = Number
	case bool:
		vt = Boolean
	default:
		return nil, fmt.Errorf("unexpected value type %s [%v=%v]",
			reflect.ValueOf(v).Kind(), k, v)
	}

	if t.next == nil {
		t.next = t.newTree()
		t.next.self = v
		t.next.key = k
	}

	return &vt, nil
}

type treeObjectIterator func(key string, value interface{}, valueType ValueType) error

// ObjectEach is for walking Object branches
func (t *Tree) ObjectEach(fn treeObjectIterator) error {
	x, ok := t.self.(map[string]interface{})
	if !ok {
		return fmt.Errorf("%s is not map[string]interface{} (ObjectEach)", reflect.ValueOf(t.self).Kind())
	}
	for k := range x {
		vt, err := t.makeNext(k, x[k])
		if err != nil {
			return err
		}
		if err := fn(k, x[k], *vt); err != nil {
			return err
		}
	}
	return nil
}

type treeArrayIterator func(index int, value interface{}, valueType ValueType) error

// ArrayEach is for walking Array branches
func (t *Tree) ArrayEach(fn treeArrayIterator) error {
	x, ok := t.self.([]interface{})
	if !ok {
		return fmt.Errorf("%s is not []interface{} (ArrayEach)", reflect.ValueOf(t.self).Kind())
	}
	for k := range x {
		vt, err := t.makeNext(k, x[k])
		if err != nil {
			return err
		}
		if err := fn(k, x[k], *vt); err != nil {
			return err
		}
	}
	return nil
}

// Each is for walking branches
func (t *Tree) Each(ifObject treeObjectIterator, ifArray treeArrayIterator) error {
	switch t.self.(type) {
	case map[string]interface{}:
		if err := t.ObjectEach(ifObject); err != nil {
			return err
		}
	case []interface{}:
		if err := t.ArrayEach(ifArray); err != nil {
			return err
		}
	default:
		return fmt.Errorf("cannot iterate over a primitive type %s", reflect.ValueOf(t.self).Kind())
	}
	return nil
}

type objectValueCallback func(i string, x map[string]interface{}) error
type arrayValueCallback func(i int, x *[]interface{}) error

func (t *Tree) keyTypeSwitch(key interface{}, ifObject objectValueCallback, ifArray arrayValueCallback) error {
	switch key.(type) {
	case string:
		x, ok := t.self.(map[string]interface{})
		if !ok {
			return fmt.Errorf("%s is not map[string]interface{} (keyTypeSwitch)", reflect.ValueOf(t.self).Kind())
		}
		return ifObject(key.(string), x)
	case int:
		x, ok := t.self.([]interface{})
		if !ok {
			return fmt.Errorf("%s is not []interface{} (keyTypeSwitch)", reflect.ValueOf(t.self).Kind())
		}
		return ifArray(key.(int), &x)
	default:
		return fmt.Errorf("key %#v is not string or number – path not found", key)
	}
}
func (t *Tree) moveTo(index int, keys []interface{}) (*Tree, error) {
	k := keys[index]
	var next *Tree
	if t.next == nil {
		next = t.newTree()
		t.next = next
	} else {
		next = t.next
	}

	ifObject := func(k string, x map[string]interface{}) error {
		var ok bool
		next.self, ok = x[k]
		if !ok {
			return fmt.Errorf("key %s in poisition %d not found in object – %v path not found", k, index, keys)
		}
		next.key = k
		next.setValue = func(newValue interface{}) {
			t.rootLock()
			x[k] = newValue
			t.rootUnlock()
		}
		next.delete = func() {
			t.rootLock()
			delete(x, k)
			t.rootUnlock()
		}
		return nil
	}

	ifArray := func(k int, x *[]interface{}) error {
		if len(*x)-1 < k {
			return fmt.Errorf("key %d in poisition %d is out of range for array – %v path not found", k, index, keys)
		}
		next.self = (*x)[k]
		next.key = k
		next.setValue = func(newValue interface{}) {
			t.rootLock()
			//log.Printf("x=%v k=%v newValue=%v", x, k, newValue)
			(*x)[k] = newValue
			//log.Printf("x=%v", x)
			t.rootUnlock()
		}
		next.delete = func() {
			next.parent.setValue(append((*x)[:k], (*x)[k+1:]...))
		}
		return nil
	}

	if err := next.parent.keyTypeSwitch(k, ifObject, ifArray); err != nil {
		return nil, err
	}
	return next, nil
}

func isPathNotFound(err error) bool {
	return strings.HasSuffix(fmt.Sprintf("%v", err), "path not found")
}
func isObjectNotArray(err error) bool {
	return strings.HasSuffix(fmt.Sprintf("%v", err), "map is not []interface{} (keyTypeSwitch)")
}

// Get fetches sub-tree at a given path
func (t *Tree) Get(keys ...interface{}) (*Tree, error) {
	iterator := t
	for index := range keys {
		i, err := iterator.moveTo(index, keys)
		if err != nil {
			return nil, err
		}
		iterator = i
	}
	return iterator, nil
}

// GetValue fetches value at a given path
func (t *Tree) GetValue(keys ...interface{}) (interface{}, error) {
	iterator, err := t.Get(keys...)
	if err != nil {
		return nil, fmt.Errorf("cannot get value at path %v – %v", keys, err)
	}
	return iterator.self, nil
}

// GetArray fetches Array at a given path
func (t *Tree) GetArray(keys ...interface{}) ([]interface{}, error) {
	v, err := t.GetValue(keys...)
	if err != nil {
		return nil, err
	}
	x, ok := v.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot cast %s to []interface{}", reflect.ValueOf(v).Kind())
	}
	return x, nil
}

// GetObject fetches Object at a given path
func (t *Tree) GetObject(keys ...interface{}) (map[string]interface{}, error) {
	v, err := t.GetValue(keys...)
	if err != nil {
		return nil, err
	}
	x, ok := v.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot cast %s to map[string]interface{}", reflect.ValueOf(v).Kind())
	}
	return x, nil
}

// GetString fetches String at a given path
func (t *Tree) GetString(keys ...interface{}) (string, error) {
	v, err := t.GetValue(keys...)
	if err != nil {
		return "", err
	}
	x, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("cannot cast %s to string", reflect.ValueOf(v).Kind())
	}
	return x, nil
}

// GetFloat fetches Float at a given path
func (t *Tree) GetFloat(keys ...interface{}) (float64, error) {
	v, err := t.GetValue(keys...)
	if err != nil {
		return 0, err
	}
	x, ok := v.(float64)
	if !ok {
		return 0, fmt.Errorf("cannot cast %s to float64", reflect.ValueOf(v).Kind())
	}
	return x, nil
}

// GetInt fetches Int at a given path
func (t *Tree) GetInt(keys ...interface{}) (int, error) {
	v, err := t.GetValue(keys...)
	if err != nil {
		return 0, err
	}
	x, ok := v.(int)
	if !ok {
		return 0, fmt.Errorf("cannot cast %s to int", reflect.ValueOf(v).Kind())
	}
	return x, nil
}

// GetBoolean fetches Boolean at a given path
func (t *Tree) GetBoolean(keys ...interface{}) (bool, error) {
	v, err := t.GetValue(keys...)
	if err != nil {
		return true, err
	}
	x, ok := v.(bool)
	if !ok {
		return true, fmt.Errorf("cannot cast %s to bool", reflect.ValueOf(v).Kind())
	}
	return x, nil
}

// Set value at a given path
func (t *Tree) Set(value interface{}, keys ...interface{}) error {
	iterator, err := t.Get(keys...)
	if err != nil {
		return fmt.Errorf("cannot set value at path %v – %v", keys, err)
	}
	iterator.setValue(value)
	return nil
}

// SetArray at a given path
func (t *Tree) SetArray(v []interface{}, keys ...interface{}) error {
	return t.Set(v, keys...)
}

// SetObject at a given path
func (t *Tree) SetObject(v map[string]interface{}, keys ...interface{}) error {
	return t.Set(v, keys...)
}

// SetString at a given path
func (t *Tree) SetString(v string, keys ...interface{}) error {
	return t.Set(v, keys...)
}

// SetFloat at a given path
func (t *Tree) SetFloat(v float64, keys ...interface{}) error {
	return t.Set(v, keys...)
}

// SetInt at a given path
func (t *Tree) SetInt(v int64, keys ...interface{}) error {
	return t.Set(v, keys...)
}

// SetBoolean at a given path
func (t *Tree) SetBoolean(v bool, keys ...interface{}) error {
	return t.Set(v, keys...)
}

// Delete a given path
func (t *Tree) Delete(keys ...interface{}) error {
	iterator, err := t.Get(keys...)
	if err != nil {
		return fmt.Errorf("cannot delete path %v – %v", keys, err)
	}
	if iterator.delete == nil {
		return fmt.Errorf("cannot delete tree root")
	}
	iterator.delete()
	return nil
}
