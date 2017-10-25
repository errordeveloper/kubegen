package converter

// TODO try using jsonparser, assume that root is a map
// build a nested map without values to track the location
// try to write some tests first
// TODO also consider https://godoc.org/github.com/json-iterator/go
// it should have similar methods, it'd better cause it's what
// Kubernetes uses, and Tim has contributed a ton of tests etc :)

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/ghodss/yaml"

	"github.com/errordeveloper/kubegen/pkg/util"
)

type BranchInfo struct {
	kind   ValueType
	self   branch
	value  []byte
	parent *BranchInfo
	path   branchPath
}

type (
	ValueType = jsonparser.ValueType

	branch     = map[string]*BranchInfo
	branchPath = []string

	KeywordEvalPhase = int
)

const (
	Null    = jsonparser.Null
	Boolean = jsonparser.Boolean
	Number  = jsonparser.Number
	String  = jsonparser.String
	Object  = jsonparser.Object
	Array   = jsonparser.Array
)

const (
	KeywordEvalPhaseA = iota
	KeywordEvalPhaseB
	KeywordEvalPhaseC
	KeywordEvalPhaseD
	KeywordEvalPhases
)

var keywordEvalPhases = [KeywordEvalPhases]KeywordEvalPhase{
	KeywordEvalPhaseA,
	KeywordEvalPhaseB,
	KeywordEvalPhaseC,
	KeywordEvalPhaseD,
}

type Converter struct {
	tree BranchInfo
	data []byte
	// TODO add module context, cause we want to be able to lookup parameters etc
	// when we encouter a keyword, we construct a callback to do modifications
	// as it's unsafe to do it right on the spot (that may be just because of
	// how our parser works)
	// TODO we will have to use regex matchers here actually, we can keep keys
	// as string and use compiledRegexp.String()
	keywords         [KeywordEvalPhases]map[string]keywordCallbackMaker
	keywordEvalPhase KeywordEvalPhase
	// modifiers are actual modifiers mapped by dot-joined path
	// TODO we probably want to do something better here, as dot-joined path
	// doesn't guarantee uniqueness (TBD, also cosider escaping literal dots)
	modifiers map[string]keywordCallback
}

type keywordCallbackMaker func(*Converter, *BranchInfo) error
type keywordCallback func(*Converter) error

type Keyword struct {
	ReturnType ValueType
	EvalPhase  KeywordEvalPhase
	FuncName   string
}

func (kw *Keyword) String() string {
	return fmt.Sprintf("kubegen.%s.%s", strings.Title(kw.ReturnType.String()), kw.FuncName)
}

func New() *Converter {
	return &Converter{
		keywords: [KeywordEvalPhases]map[string]keywordCallbackMaker{
			KeywordEvalPhaseA: make(map[string]keywordCallbackMaker),
			KeywordEvalPhaseB: make(map[string]keywordCallbackMaker),
			KeywordEvalPhaseC: make(map[string]keywordCallbackMaker),
			KeywordEvalPhaseD: make(map[string]keywordCallbackMaker),
		},
		modifiers: make(map[string]keywordCallback),
	}
}

func (c *Converter) load(data []byte) { c.data = data }

func (c *Converter) loadStrict(data []byte) error {
	obj := new(interface{})
	if err := json.Unmarshal(data, obj); err != nil {
		return fmt.Errorf("error while re-decoding – %v", err)
	}
	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("error while re-encoding – %v", err)
	}
	c.load(data)
	return nil
}

// TODO util.LoadObj and Unmarshall have a lot in common
// TODO error strings are kind of quite messy

func (c *Converter) LoadObj(data []byte, sourcePath string, instanceName string) error {
	obj := new(interface{})
	if err := util.LoadObj(obj, data, sourcePath, instanceName); err != nil {
		return err
	}
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("error while re-encoding %q (%q): %v", instanceName, sourcePath, err)
	}
	c.load(jsonData)
	return nil
}

func (c *Converter) Unmarshal(obj interface{}, sourcePath string, instanceName string) error {
	var errorFmt string

	if instanceName != "" {
		errorFmt = fmt.Sprintf("error while re-decoding module %q source", instanceName)
	} else {
		errorFmt = "error while re-decoding manifest file"
	}

	ext := path.Ext(sourcePath)
	switch {
	case ext == ".json" || ext == ".yaml" || ext == ".yml":
		if err := json.Unmarshal(c.data, obj); err != nil {
			return fmt.Errorf("%s %q (%q): %v", errorFmt, instanceName, sourcePath, err)
		}
	case ext == ".kg" || ext == ".hcl":
		if err := util.NewFromHCL(obj, c.data); err != nil {
			return fmt.Errorf("%s %q (%q): %v", errorFmt, instanceName, sourcePath, err)
		}
	default:
		return fmt.Errorf("%s %q – unknown file extension", errorFmt, sourcePath)
	}

	return nil
}

func (c *Converter) doIterate(parentBranch *BranchInfo, key string, value []byte, dataType ValueType, errors chan error) {
	pathLen := len(parentBranch.path) + 1
	newBranch := BranchInfo{
		parent: parentBranch,
		kind:   dataType,
		value:  value,
		self:   make(branch),
		path:   make(branchPath, pathLen),
	}

	// we have to copy the slice explicitly, as it turns append doesn't make a copy
	copy(newBranch.path, parentBranch.path)
	newBranch.path[pathLen-1] = key

	if _, ok := parentBranch.self[key]; ok {
		errors <- fmt.Errorf("key %q is already set in parent", key)
		return
	}
	parentBranch.self[key] = &newBranch
	if makeModifier, ok := c.keywords[c.keywordEvalPhase][key]; ok {
		if err := makeModifier(c, &newBranch); err != nil {
			errors <- fmt.Errorf("failed to register modifier for keyword %q – %v", key, err)
			return
		}
	}

	switch dataType {
	case jsonparser.Object:
		handler := c.makeObjectIterator(&newBranch, errors)

		if err := jsonparser.ObjectEach(value, handler); err != nil {
			errors <- (err)
			return
		}
	case jsonparser.Array:
		handler := c.makeArrayIterator(&newBranch, errors)

		jsonparser.ArrayEach(value, handler)
	}
}

type objectIterator func(key []byte, value []byte, dataType ValueType, offset int) error

func (c *Converter) makeObjectIterator(parentBranch *BranchInfo, errors chan error) objectIterator {
	callback := func(key []byte, value []byte, dataType ValueType, offset int) error {
		c.doIterate(parentBranch, string(key), value, dataType, errors)
		keys := []string{}
		paths := [][]string{}
		for i := range parentBranch.self {
			keys = append(keys, i)
			paths = append(paths, parentBranch.self[i].path)
		}
		return nil
	}
	return callback
}

type arrayIterator func(value []byte, dataType ValueType, offset int, err error)

func (c *Converter) makeArrayIterator(parentBranch *BranchInfo, errors chan error) arrayIterator {
	index := 0
	callback := func(value []byte, dataType ValueType, offset int, err error) {
		c.doIterate(parentBranch, fmt.Sprintf("[%d]", index), value, dataType, errors)
		index = index + 1
	}
	return callback
}

func (c *Converter) interate(errors chan error) {
	if err := jsonparser.ObjectEach(c.data, c.makeObjectIterator(&c.tree, errors)); err != nil {
		errors <- (err)
		return
	}
	errors <- nil
}

func (c *Converter) checkKind() (err error) {
	kindUpper, errUpper := jsonparser.GetString(c.data, "Kind")
	kindLower, errLower := jsonparser.GetString(c.data, "kind")
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

	c.tree = BranchInfo{
		parent: nil,
		kind:   jsonparser.Object,
		value:  c.data,
		self:   make(branch),
		path:   []string{""},
	}
	c.keywordEvalPhase = phase

	{
		errors := make(chan error)

		go c.interate(errors)

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

var (
	KeywordBooleanIf = &Keyword{
		ReturnType: Null,
		EvalPhase:  KeywordEvalPhaseA,
		FuncName:   "If",
	}

	KeywordBooleanLookup = &Keyword{
		ReturnType: Boolean,
		EvalPhase:  KeywordEvalPhaseB,
		FuncName:   "Lookup",
	}
	KeywordStringLookup = &Keyword{
		ReturnType: String,
		EvalPhase:  KeywordEvalPhaseB,
		FuncName:   "Lookup",
	}
	KeywordNumberLookup = &Keyword{
		ReturnType: Number,
		EvalPhase:  KeywordEvalPhaseB,
		FuncName:   "Lookup",
	}
	KeywordArrayLookup = &Keyword{
		ReturnType: Array,
		EvalPhase:  KeywordEvalPhaseB,
		FuncName:   "Lookup",
	}
	KeywordObjectLookup = &Keyword{
		ReturnType: Object,
		EvalPhase:  KeywordEvalPhaseB,
		FuncName:   "Lookup",
	}

	KeywordStringJoin = &Keyword{
		ReturnType: String,
		EvalPhase:  KeywordEvalPhaseC,
		FuncName:   "Join",
	}

	KeywordStringAsJSON = &Keyword{
		ReturnType: String,
		EvalPhase:  KeywordEvalPhaseD,
		FuncName:   "AsJSON",
	}
	KeywordStringAsYAML = &Keyword{
		ReturnType: String,
		EvalPhase:  KeywordEvalPhaseD,
		FuncName:   "AsYAML",
	}
)

func StringJoin(c *Converter, branch *BranchInfo) error {
	if branch.Kind() != Array {
		return fmt.Errorf("must be an array")
	}
	c.AddModifier(branch, func(c *Converter) error {
		x := []string{}
		jsonparser.ArrayEach(branch.Value(), func(value []byte, dataType ValueType, offset int, err error) {
			x = append(x, string(value))
		})
		if err := c.Replace(branch, []byte(fmt.Sprintf("%q", strings.Join(x, "")))); err != nil {
			return fmt.Errorf("could not join string – %v", err)
		}
		return nil
	})
	return nil
}

func StringAsYAML(c *Converter, branch *BranchInfo) error {
	c.AddModifier(branch, func(c *Converter) error {
		o := new(interface{})
		if err := json.Unmarshal(branch.Value(), o); err != nil {
			return err
		}
		x, err := yaml.Marshal(o)
		if err != nil {
			return err
		}
		{
			x, err := json.Marshal(string(x))
			if err != nil {
				return err
			}
			if err = c.Replace(branch, x); err != nil {
				return err
			}
			return nil
		}
	})
	return nil
}

func StringAsJSON(c *Converter, branch *BranchInfo) error {
	c.AddModifier(branch, func(c *Converter) error {
		x, err := json.Marshal(string(branch.Value()))
		if err != nil {
			return err
		}
		if err = c.Replace(branch, x); err != nil {
			return err
		}
		return nil
	})
	return nil
}

func (c *Converter) DefineKeyword(kw *Keyword, fn keywordCallbackMaker) {
	// TODO compile regex here and use compiledRegexp.String() to get the key back?
	c.keywords[kw.EvalPhase][kw.String()] = fn
}

func (c *Converter) AddModifier(branch *BranchInfo, fn keywordCallback) {
	c.modifiers[branch.PathToString()] = fn
}

func (c *Converter) callModifiersOnce() error {
	for p, fn := range c.modifiers {
		if err := fn(c); err != nil {
			return fmt.Errorf("callback on %q failed to modify the tree – %v", p, err)
		}
		delete(c.modifiers, p)
	}
	return nil
}

func (b *BranchInfo) get(key string) *BranchInfo {
	if next := b.self[key]; next != nil {
		return next
	}
	return nil
}

func (c *Converter) get(path ...string) *BranchInfo {
	branch := BranchInfo{self: c.tree.self}
	for x := range path {
		if next := branch.get(path[x]); next != nil {
			branch = *next
		} else {
			return nil
		}
	}
	return &branch
}

func (b *BranchInfo) Kind() ValueType { return b.kind }

func (b *BranchInfo) Value() []byte { return b.value }

func (b *BranchInfo) PathToString() string {
	// TODO escape literal dots with something or find another join character
	// TODO look into what JSONPath does about this, also check jq too
	return strings.Join(b.path, ".")
}

func (c *Converter) Replace(branch *BranchInfo, value []byte) error {
	x, err := jsonparser.Set(c.data, value, branch.parent.path[1:]...)
	if err != nil {
		return err
	}
	c.load(x)
	return nil
}
