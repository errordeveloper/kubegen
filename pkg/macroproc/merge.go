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

func (t *Tree) overlayHere(key interface{}, value interface{}) error {
	ifObject := func(i string, x map[string]interface{}) error {
		t.rootLock()
		defer t.rootUnlock()

		x[i] = value
		return nil
	}
	ifArray := func(i int, x []interface{}) error {
		// no need for a lock, as setValue will lock it
		switch {
		case i == len(x):
			//log.Printf("appending to array – t=%v", t)
			t.setValue(append(x, value))
			//log.Printf("appended to array – t=%v", t)
		case i > len(x):
			//log.Println("extending array")
			extLen := i - len(x)
			ext := make([]interface{}, extLen)
			ext[extLen-1] = value
			t.parent.setValue(append(x, ext))
		default:
			return fmt.Errorf("cannot overlay %d to array of len %d", key, len(x))
		}
		return nil
	}
	if err := t.keyTypeSwitch(key, ifObject, ifArray); err != nil {
		return err
	}
	return nil
}

func (t *Tree) overlay(source *Tree) error {
	// for each key in source, attempt to overlay it onto the target
	iterateSource := func(k interface{}, v interface{}, vt ValueType) error {
		//log.Printf("<t:%s>.Get(<k:%v>)", t, k)
		targetSubtree, err := t.Get(k)
		if targetSubtree != nil && err == nil {
			//log.Println("key found in target!")
			switch vt {
			case Object:
				nextSource := NewTree(&v)
				//log.Printf("<targetSubtree:%s>.overlay(<nextSource:%v>)", targetSubtree, nextSource)
				return targetSubtree.overlay(nextSource)
			case Array:
				nextSource := NewTree(&v)
				//log.Printf("<targetSubtree:%s>.overlay(<nextSource:%v>)", targetSubtree, nextSource)
				return targetSubtree.overlay(nextSource)
			default:
				//log.Printf("leave k=%v of primitive type %s as is!", k, vt)
			}
		} else if isPathNotFound(err) {
			//log.Printf("<t:%s>.overlayHere(<k:%v>, <v:%v>)", t, k, v)
			return t.overlayHere(k, v)
		} else {
			return err
		}

		return nil
	}

	ifObject := func(k string, v interface{}, vt ValueType) error {
		//log.Printf("iterateSourceObject(<k:%v>, <v:%v>, <vt:%s>)", k, v, vt)
		return iterateSource(k, v, vt)
	}
	ifArray := func(k int, v interface{}, vt ValueType) error {
		//log.Printf("iterateSourceArray(<k:%v>, <v:%v>, <vt:%s>)", k, v, vt)
		return iterateSource(k, v, vt)
	}
	if err := source.Each(ifObject, ifArray); err != nil {
		return err
	}

	return nil
}
