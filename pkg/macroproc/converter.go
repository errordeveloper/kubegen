package macroproc

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/errordeveloper/kubegen/pkg/util"
)

type BranchLocator struct {
	kind       ValueType
	self       branch
	value      *Tree
	parent     *BranchLocator
	path       branchPath
	stringPath stringBranchPath
}

type (
	ValueType int

	branch           = map[interface{}]*BranchLocator
	branchPath       = []interface{}
	stringBranchPath = []string

	MacrosEvalPhase   = int
	MacroArgumentType = int
)

const (
	Null = ValueType(iota)
	Boolean
	Number
	String
	Object
	Array
)

type Converter struct {
	locator BranchLocator
	tree    *Tree
	// TODO add module context, cause we want to be able to lookup parameters etc
	// when we encouter a macro, we construct a callback to do modifications
	// as it's unsafe to do it right on the spot (that may be just because of
	// how our parser works)
	// TODO we will have to use regex matchers here actually, we can keep keys
	// as string and use compiledRegexp.String()
	macros          [MacrosEvalPhases]map[string]*UnregisteredModifier
	macroMatcher    *macroMatcher
	macrosEvalPhase MacrosEvalPhase
	// modifiers are actual modifiers mapped by path
	modifiers map[string]*Modifier
}

func New() *Converter {
	return &Converter{
		macros: [MacrosEvalPhases]map[string]*UnregisteredModifier{
			MacrosEvalPhaseA: make(map[string]*UnregisteredModifier),
			MacrosEvalPhaseB: make(map[string]*UnregisteredModifier),
			MacrosEvalPhaseC: make(map[string]*UnregisteredModifier),
			MacrosEvalPhaseD: make(map[string]*UnregisteredModifier),
			MacrosEvalPhaseE: make(map[string]*UnregisteredModifier),
		},
		macroMatcher: newMacroMatcher(),
		modifiers:    make(map[string]*Modifier),
	}
}

func (c *Converter) load(data []byte) (err error) {
	c.tree, err = loadObject(data)
	return
}

func (c *Converter) loadStrict(data []byte) error {
	v, err := util.EnsureJSON(data)
	if err != nil {
		return err
	}
	return c.load(v)
}

func (c *Converter) LoadObject(data []byte, sourcePath string, instanceName string) error {
	obj := new(interface{})
	if err := util.LoadObj(obj, data, sourcePath, instanceName); err != nil {
		return err
	}
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("error while re-encoding %q (%q): %v", instanceName, sourcePath, err)
	}
	c.tree, err = loadObject(jsonData)
	return err
}

func (c *Converter) UnloadObject(obj interface{}, sourcePath string, instanceName string) error {
	jsonData, err := c.MarshalJSON()
	if err != nil {
		return err
	}
	if err := util.LoadObj(obj, jsonData, sourcePath, instanceName); err != nil {
		// log.Printf("c.tree=%s", c.tree)
		return err
	}
	return nil
}

func (c *Converter) MarshalJSON() ([]byte, error) { return json.Marshal(c.tree.self) }

func (c *Converter) doIterate(parentBranch *BranchLocator, key interface{}, value interface{}, dataType ValueType, errors chan error) {
	pathLen := len(parentBranch.path) + 1
	newBranch := BranchLocator{
		parent:     parentBranch,
		kind:       dataType,
		value:      NewTree(&value),
		self:       make(branch),
		path:       make(branchPath, pathLen),
		stringPath: make(stringBranchPath, pathLen),
	}

	// we have to copy the slice explicitly, as it turns append doesn't make a copy
	copy(newBranch.path, parentBranch.path)
	copy(newBranch.stringPath, parentBranch.stringPath)

	k := formatKey(key)
	if k == "" {
		errors <- fmt.Errorf("key %#v is not string or int", key)
	}

	newBranch.path[pathLen-1] = key
	newBranch.stringPath[pathLen-1] = k

	if _, ok := parentBranch.self[key]; ok {
		errors <- fmt.Errorf("key %q is already set in parent", key)
		return
	}
	parentBranch.self[key] = &newBranch

	c.ifMacroDoRegister(&newBranch, key, errors)

	switch dataType {
	case Object:
		handler := c.makeObjectIterator(&newBranch, errors)

		if err := newBranch.value.ObjectEach(handler); err != nil {
			errors <- (err)
			return
		}
	case Array:
		handler := c.makeArrayIterator(&newBranch, errors)

		if err := newBranch.value.ArrayEach(handler); err != nil {
			errors <- (err)
			return
		}
	}
}

func (c *Converter) makeObjectIterator(parentBranch *BranchLocator, errors chan error) treeObjectIterator {
	callback := func(key string, value interface{}, dataType ValueType) error {
		c.doIterate(parentBranch, key, value, dataType, errors)
		return nil
	}
	return callback
}

func (c *Converter) makeArrayIterator(parentBranch *BranchLocator, errors chan error) treeArrayIterator {
	callback := func(index int, value interface{}, dataType ValueType) error {
		c.doIterate(parentBranch, index, value, dataType, errors)
		return nil
	}
	return callback
}

func (c *Converter) iterate(errors chan error) {
	if err := c.tree.ObjectEach(c.makeObjectIterator(&c.locator, errors)); err != nil {
		errors <- (err)
		return
	}
	errors <- nil
}

func (c *Converter) checkKind() (err error) {
	kindUpper, errUpper := c.tree.GetString("Kind")
	kindLower, errLower := c.tree.GetString("kind")
	if errUpper != nil && errLower != nil {
		err = fmt.Errorf("unknown kind of object (kind unspecified)")
	}

	if kindUpper == "" && kindLower == "" {
		err = fmt.Errorf("unknown kind of object (empty string)")
	}
	return
}

func (c *Converter) run(phase MacrosEvalPhase) error {
	// it is okay to check this here, so we fail early,
	// however we should abstract deep checks more specifically
	if err := c.checkKind(); err != nil {
		return err
	}

	c.locator = BranchLocator{
		parent:     nil,
		kind:       Object,
		value:      c.tree,
		self:       make(branch),
		path:       []interface{}{nil},
		stringPath: []string{""},
	}
	c.macrosEvalPhase = phase

	{
		errors := make(chan error)

		go c.iterate(errors)

		err := <-errors
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Converter) Run() error {
	eval := func(phase MacrosEvalPhase) error {
		for {
			if err := c.run(phase); err != nil {
				return err
			}
			if len(c.modifiers) == 0 {
				return nil
			}
			// log.Println(c.modifiers)
			if err := c.callModifiersOnce(); err != nil {
				return err
			}
		}
	}
	for phase := range macrosEvalPhases {
		// log.Printf("len(c.modifiers)=%d eval(<phase:%d>)", len(c.modifiers), phase)
		if err := eval(phase); err != nil {
			return err
		}
	}

	return nil
}

func (b *BranchLocator) get(key interface{}) *BranchLocator {
	if next := b.self[key]; next != nil {
		return next
	}
	return nil
}

func (c *Converter) get(path ...interface{}) *BranchLocator {
	branch := BranchLocator{self: c.locator.self}
	for x := range path {
		if next := branch.get(path[x]); next != nil {
			branch = *next
		} else {
			return nil
		}
	}
	return &branch
}

// Refresh get latest value from t and recurses into parents
func (b *BranchLocator) Refresh(c *Converter) error {
	v, err := c.tree.Get(b.path[1:]...)
	if err != nil {
		return fmt.Errorf("cannot refresh value at %s (b.value=%s c.tree=%s) – %v", b.PathToString(), b.value, c.tree, err)
	}
	b.value = v
	if b.parent != nil {
		return b.parent.Refresh(c)
	}
	return nil
}

func (b *BranchLocator) Kind() ValueType { return b.kind }

func (b *BranchLocator) Value() *Tree { return b.value }

func (b *BranchLocator) StringValue() *string {
	v, ok := b.value.self.(string)
	if !ok {
		return nil
	}
	return &v
}

func (b *BranchLocator) PathToString() string {
	return strings.Join(b.stringPath, "")
}

func formatKey(k interface{}) string {
	switch k.(type) {
	case string:
		return fmt.Sprintf("[%q]", k)
	case int:
		return fmt.Sprintf("[%d]", k)
	default:
		return ""
	}
}

func (vt ValueType) String() string {
	switch vt {
	case String:
		return "String"
	case Number:
		return "Number"
	case Object:
		return "Object"
	case Array:
		return "Array"
	case Boolean:
		return "Boolean"
	case Null:
		return "Null"
	default:
		return "Unknown"
	}
}
