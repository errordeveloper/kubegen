package converter

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

	KeywordEvalPhase = int
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
	// when we encouter a keyword, we construct a callback to do modifications
	// as it's unsafe to do it right on the spot (that may be just because of
	// how our parser works)
	// TODO we will have to use regex matchers here actually, we can keep keys
	// as string and use compiledRegexp.String()
	keywords         [KeywordEvalPhases]map[string]*UnregisteredModifier
	keywordMatcher   *keywordMatcher
	keywordEvalPhase KeywordEvalPhase
	// modifiers are actual modifiers mapped by dot-joined path
	// TODO we probably want to do something better here, as dot-joined path
	// doesn't guarantee uniqueness (TBD, also cosider escaping literal dots)
	modifiers map[string]*Modifier
}

func New() *Converter {
	return &Converter{
		keywords: [KeywordEvalPhases]map[string]*UnregisteredModifier{
			KeywordEvalPhaseA: make(map[string]*UnregisteredModifier),
			KeywordEvalPhaseB: make(map[string]*UnregisteredModifier),
			KeywordEvalPhaseC: make(map[string]*UnregisteredModifier),
			KeywordEvalPhaseD: make(map[string]*UnregisteredModifier),
		},
		keywordMatcher: newKeywordMatcher(),
		modifiers:      make(map[string]*Modifier),
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
	jsonData, err := json.Marshal(c.tree.self)
	if err != nil {
		return err
	}
	return util.LoadObj(obj, jsonData, sourcePath, instanceName)
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

	c.ifKeywordDoRegister(&newBranch, key, errors)

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
	callback := func(key string, value interface{}, dataType ValueType, _ *Tree) error {
		c.doIterate(parentBranch, key, value, dataType, errors)
		/*
			paths := [][]interface{}{}
			for i := range parentBranch.self {
				paths = append(paths, parentBranch.self[i].path)
			}
		*/
		return nil
	}
	return callback
}

func (c *Converter) makeArrayIterator(parentBranch *BranchLocator, errors chan error) treeArrayIterator {
	callback := func(index int, value interface{}, dataType ValueType, _ *Tree) error {
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

func (c *Converter) run(phase KeywordEvalPhase) error {
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
	c.keywordEvalPhase = phase

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
	eval := func(phase KeywordEvalPhase) error {
		for {
			if err := c.run(phase); err != nil {
				return err
			}
			if len(c.modifiers) == 0 {
				return nil
			}
			if err := c.callModifiersOnce(); err != nil {
				return err
			}
		}
	}

	for phase := range keywordEvalPhases {
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
