package modules

import (
	"fmt"
	"reflect"
	//"strings"

	"github.com/mitchellh/reflectwalk"
)

// TODO: add validation for all known "kubegen.*" keys

type interpolationWalker struct {
	FindKey  string
	Callback interpolationWalkerFunc

	//Replace bool

	// ContextF is an advanced version of F that also receives the
	// location of where it is in the structure. This lets you do
	// context-aware validation.
	ContextF interpolationWalkerContextFunc

	key        []string
	lastValue  reflect.Value
	loc        reflectwalk.Location
	cs         []reflect.Value
	csKey      []reflect.Value
	csData     interface{}
	sliceIndex []int
}

type interpolationWalkerFunc func(map[string]interface{}) (map[string]interface{}, error)

// interpolationWalkerContextFunc is called by interpolationWalk if
// ContextF is set. This receives both the interpolation and the location
// where the interpolation is.
//
// This callback can be used to validate the location of the interpolation
// within the configuration.
type interpolationWalkerContextFunc func(reflectwalk.Location, interface{})

func (w *interpolationWalker) Enter(loc reflectwalk.Location) error {
	w.loc = loc
	return nil
}

func (w *interpolationWalker) Exit(loc reflectwalk.Location) error {
	w.loc = reflectwalk.None

	switch loc {
	case reflectwalk.Map:
		w.cs = w.cs[:len(w.cs)-1]
	case reflectwalk.MapValue:
		w.key = w.key[:len(w.key)-1]
		w.csKey = w.csKey[:len(w.csKey)-1]
	case reflectwalk.Slice:
		// Split any values that need to be split
		if err := w.splitSlice(); err != nil {
			return err
		}
		w.cs = w.cs[:len(w.cs)-1]
	case reflectwalk.SliceElem:
		w.csKey = w.csKey[:len(w.csKey)-1]
		w.sliceIndex = w.sliceIndex[:len(w.sliceIndex)-1]
	}

	return nil
}

func (w *interpolationWalker) Map(m reflect.Value) error {
	fmt.Printf("Map(original): %#v\n", m)
	i := m.Interface().(map[string]interface{})
	if _, ok := i[w.FindKey]; ok {
		v, err := w.Callback(i)
		if err != nil {
			return err
		}
		m = reflect.ValueOf(v)
		fmt.Printf("Map(modified): %#v\n", m)
	}
	w.cs = append(w.cs, m)
	return nil
}

func (w *interpolationWalker) MapElem(m, k, v reflect.Value) error {
	w.csData = k
	w.csKey = append(w.csKey, k)

	if l := len(w.sliceIndex); l > 0 {
		w.key = append(w.key, fmt.Sprintf("%d.%s", w.sliceIndex[l-1], k.String()))
	} else {
		w.key = append(w.key, k.String())
	}

	w.lastValue = v
	return nil
}

func (w *interpolationWalker) Slice(s reflect.Value) error {
	w.cs = append(w.cs, s)
	return nil
}

func (w *interpolationWalker) SliceElem(i int, elem reflect.Value) error {
	w.csKey = append(w.csKey, reflect.ValueOf(i))
	w.sliceIndex = append(w.sliceIndex, i)
	return nil
}

func (w *interpolationWalker) replaceCurrent(v reflect.Value) {
	// if we don't have at least 2 values, we're not going to find a map, but
	// we could panic.
	if len(w.cs) < 2 {
		return
	}

	c := w.cs[len(w.cs)-2]
	switch c.Kind() {
	case reflect.Map:
		// Get the key and delete it
		k := w.csKey[len(w.csKey)-1]
		c.SetMapIndex(k, v)
	}
}

func (w *interpolationWalker) splitSlice() error {
	raw := w.cs[len(w.cs)-1]

	if raw.Kind() != reflect.Interface {
		return fmt.Errorf("encoutered non-interface type: %#v", raw)
	}

	var s []interface{}
	switch v := raw.Interface().(type) {
	case []interface{}:
		s = v
	case []map[string]interface{}:
		return nil
	}

	split := false
	for _, val := range s {
		if _, ok := val.([]interface{}); ok {
			split = true
		}
	}

	if !split {
		return nil
	}

	result := make([]interface{}, 0)
	for _, v := range s {
		switch val := v.(type) {
		case []interface{}:
			for _, element := range val {
				result = append(result, element)
			}
		default:
			result = append(result, v)
		}
	}

	w.replaceCurrent(reflect.ValueOf(result))

	return nil
}
