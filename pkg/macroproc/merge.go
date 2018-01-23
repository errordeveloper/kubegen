package macroproc

import (
	"fmt"
)

// getPathCutoffIndex is unused, but we have a test for it and it may come handy
func (t *Tree) getPathCutoffIndex(keys ...interface{}) int {
	x := -1
	iterator := t
	for index := range keys {
		i, err := iterator.moveTo(index, keys)
		if err != nil {
			return x
		}
		iterator = i
		x = index
	}
	return x
}

// Overlay source tree onto
func (t *Tree) Overlay(source *Tree, keys ...interface{}) error {
	targetSubtree, err := t.Get(keys...)
	if err != nil {
		return fmt.Errorf("cannot overlay at path %v – %v", keys, err)
	}

	if source.self == nil {
		return fmt.Errorf("source.self cannot be nil")
	}

	return targetSubtree.overlay(source)
}

func (t *Tree) overlayHere(newKey interface{}, newValue interface{}) error {
	ifObject := func(k string, v map[string]interface{}) error {
		// log.Printf("=> <t:%v>.overlayHereObject(<newKey:%v>, <newValue:%v>)", t, newKey, newValue)
		t.rootLock()
		defer t.rootUnlock()

		v[k] = newValue
		// log.Printf("<= <t:%v>.overlayHereObject(<newKey:%v>, <newValue:%v>)", t, newKey, newValue)
		return nil
	}
	ifArray := func(k int, v *[]interface{}) error {
		// log.Printf("=> <t:%v>.overlayHereArray(<newKey:%v>, <newValue:%v>)", t, newKey, newValue)
		// no need for t.rootLock as t.setValue takes the lock

		switch {
		case k == len(*v):
			// log.Printf("appending to array – t=%v", t)
			t.setValue(append(*v, newValue))
		// arrays are always iterated in the same order, so we shouldn't need this
		/*
			case k > len(*v):
					// log.Printf("extending array – t=%v", t)
					extLen := k - len(*v)
					ext := make([]interface{}, extLen)
					ext[extLen-1] = newValue
					t.setValue(append(*v, ext...))
		*/
		default:
			return fmt.Errorf("cannot overlay array of len %d", len(*v))
		}
		// log.Printf("<= <t:%v>.overlayHereArray(<newKey:%v>, <newValue:%v>)", t, newKey, newValue)
		return nil
	}
	if err := t.keyTypeSwitch(newKey, ifObject, ifArray); err != nil {
		return err
	}
	return nil
}

func (t *Tree) replaceEmptiedObjectWithArray(source *Tree) error {
	// log.Printf("=> <t:%s>.replaceEmptiedObjectWithArray(<source:%v>)", t, source)

	if v, ok := t.self.(map[string]interface{}); ok {
		switch {
		case len(v) == 0:
			// log.Println("new array with one element")
			t.setValue(source.self.([]interface{}))
		default:
			return fmt.Errorf("cannot replace non-empty object with an array")
		}
	}
	// log.Printf("<= <t:%s>.replaceEmptiedObjectWithArray(<source:%v>)", t, source)
	// we must return a nil here, otherwise we iterate over this again an get an error
	return nil
	// TODO perhaps we are still have some stale state, but this generally works
	// return fmt.Errorf("cannot replace %s with an array", reflect.ValueOf(t.self).Kind())
}

func (t *Tree) overlay(source *Tree) error {
	// for each key in source, attempt to overlay it onto the target
	iterateSource := func(key interface{}, value interface{}, vt ValueType) error {
		// log.Printf("<t:%s>.Get(<k:%v>)", t, key)
		targetSubtree, err := t.Get(key)
		if targetSubtree != nil && err == nil {
			// iterate over non-primitive types, otherwise leave source as is
			if vt == Object || vt == Array {
				nextSource := NewTree(&value)
				// log.Printf("<targetSubtree:%s>.overlay(<nextSource:%v>)", targetSubtree, nextSource)
				return targetSubtree.overlay(nextSource)
			}
		} else if isPathNotFound(err) {
			// when key is missing, add it
			return t.overlayHere(key, value)
		} else if isObjectNotArray(err) {
			// very common case of { "kubegen.Array.Lookup": "someParam" } => [ <someParamValue> ]
			return t.replaceEmptiedObjectWithArray(source)
		} else {
			return err
		}
		//fmt.Printf("return <t:%s>.overlay(<source:%s>)", t, source)
		return nil
	}

	ifObject := func(key string, value interface{}, vt ValueType) error {
		//log.Printf("iterateSourceObject(<k:%v>, <v:%v>, <vt:%s>)", key, value, vt)
		return iterateSource(key, value, vt)
	}
	ifArray := func(key int, value interface{}, vt ValueType) error {
		//log.Printf("iterateSourceArray(<k:%v>, <v:%v>, <vt:%s>)", key, value, vt)
		return iterateSource(key, value, vt)
	}
	if err := source.Each(ifObject, ifArray); err != nil {
		return err
	}

	return nil
}
