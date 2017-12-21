package converter

import (
	"fmt"
	"reflect"
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

func (t *Tree) Submerge(source *Tree, keys ...interface{}) error {
	target, err := t.Get(keys...)
	if err != nil {
		return fmt.Errorf("cannot submerge at path %v – %v", keys, err)
	}

	if source.self == nil {
		return fmt.Errorf("source.self cannot be nil")
	}

	return target.merger(source)
}

func (t *Tree) mergeHere(key interface{}, value interface{}) error {
	objectMerger := func(i string, x map[string]interface{}) error {
		t.rootLock()
		defer t.rootUnlock()

		x[i] = value
		return nil
	}
	arrayMerger := func(i int, x []interface{}) error {
		t.rootLock()
		defer t.rootUnlock()

		switch {
		case i == len(x):
			t.parent.setValue(append(x, value))
		case i > len(x):
			extLen := i - len(x)
			ext := make([]interface{}, extLen)
			ext[extLen-1] = value
			t.parent.setValue(append(x, ext))
		default:
			return fmt.Errorf("cannot add %d to array of len %d", key, len(x))
		}
		return nil
	}
	if err := t.keyTypeSwitch(key, objectMerger, arrayMerger); err != nil {
		return err
	}
	return nil
}

func (t *Tree) merger(source *Tree) error {
	targetKind := reflect.ValueOf(t.self).Kind()
	sourceKind := reflect.ValueOf(source.self).Kind()
	if targetKind != sourceKind {
		return fmt.Errorf("%s(%v) and %s(%v) are not the same type", targetKind, t.self, sourceKind, source.self)
	}

	iterateSource := func(k interface{}, v interface{}, vt ValueType, nextSource *Tree) error {
		if nextSource.self == nil {
			return fmt.Errorf("nextSource.self cannot be nil")
		}

		targetSubtree, err := t.Get(k)
		if targetSubtree != nil && err == nil {
			if vt == Object || vt == Array {
				return targetSubtree.merger(nextSource)
				// return t.merger(nextSource) behaves different, but both options are flaky for some reason
			}
			return nil // leave any existing key as is
		}

		if err != nil {
			if isPathNotFound(err) {
				return t.mergeHere(k, v)
			}
			return err
		}

		return fmt.Errorf("cannot submerge – unexpected error")
	}

	ifObject := func(k string, v interface{}, vt ValueType, n *Tree) error {
		return iterateSource(k, v, vt, n)
	}
	ifArray := func(k int, v interface{}, vt ValueType, n *Tree) error {
		return iterateSource(k, v, vt, n)
	}
	if err := source.Each(ifObject, ifArray); err != nil {
		return err
	}

	return nil
}
