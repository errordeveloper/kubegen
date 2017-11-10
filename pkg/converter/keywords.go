package converter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/errordeveloper/kubegen/pkg/util"
	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"
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

type keywordCallbackMaker func(*Converter, *BranchInfo) error
type keywordCallback func(*Converter) error

type Keyword struct {
	ReturnType ValueType
	EvalPhase  KeywordEvalPhase
	FuncName   string
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

	LoadObjectJSON = &Keyword{
		ReturnType: Object,
		EvalPhase:  KeywordEvalPhaseA,
		FuncName:   "LoadJSON",
	}

	LoadArrayJSON = &Keyword{
		ReturnType: Array,
		EvalPhase:  KeywordEvalPhaseA,
		FuncName:   "LoadJSON",
	}
)

func (kw *Keyword) String() string {
	return fmt.Sprintf("kubegen.%s.%s", strings.Title(kw.ReturnType.String()), kw.FuncName)
}

func StringJoin(c *Converter, branch *BranchInfo) error {
	if branch.Kind() != Array {
		return fmt.Errorf("must be an array")
	}
	c.AddModifier(branch, func(c *Converter) error {
		x := []string{}
		jsonparser.ArrayEach(branch.Value(), func(value []byte, dataType ValueType, offset int, err error) {
			x = append(x, string(value))
		})
		if err := c.Set(branch, strings.Join(x, "")); err != nil {
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
			if err = c.Set(branch, string(x)); err != nil {
				return err
			}
			return nil
		}
	})
	return nil
}

func StringAsJSON(c *Converter, branch *BranchInfo) error {
	c.AddModifier(branch, func(c *Converter) error {
		if err := c.Set(branch, string(branch.Value())); err != nil {
			return err
		}
		return nil
	})
	return nil
}

func doLoadJSON(c *Converter, branch *BranchInfo, kw *Keyword, newData []byte) error {
	var (
		err         error
		oldData     []byte
		oldDataTemp []byte
		oldDataType jsonparser.ValueType
		oldObj      map[string]interface{}
		newObj      interface{}
	)

	isRoot := (len(branch.path[1:]) == 1)
	switch kw.ReturnType {
	case Object:
		if isRoot {
			_, oldDataType, _, err = jsonparser.Get(c.data)
			oldData = make([]byte, len(c.data))
			copy(oldData, c.data)
		} else {
			oldDataTemp, oldDataType, _, err = jsonparser.Get(c.data, branch.parent.path[1:]...)
			oldData = make([]byte, len(oldDataTemp))
			copy(oldData, oldDataTemp)
		}
	case Array:
		if isRoot {
			return fmt.Errorf("cannot insert array in place of root object")
		}
		oldDataTemp, oldDataType, _, err = jsonparser.Get(c.data, branch.parent.path[1:]...)
		oldData = make([]byte, len(oldDataTemp))
		copy(oldData, oldDataTemp)
	}

	switch {
	case err != nil:
		return fmt.Errorf("cannot get old data – %v", err)
	case len(oldData) == 0:
		return fmt.Errorf("old data is empty")
	case oldDataType != jsonparser.Object:
		return fmt.Errorf("old data type is %s, but must be an object", oldDataType)
	}

	oldData = jsonparser.Delete(oldData, branch.path[len(branch.path)-1])

	if err := json.Unmarshal(oldData, &oldObj); err != nil {
		return fmt.Errorf("cannot unmarshal old data – %v", err)
	}

	if err := json.Unmarshal(newData, &newObj); err != nil {
		return fmt.Errorf("cannot unmarshal new data – %v", err)
	}

	switch kw.ReturnType {
	case Object:
		if err := mergo.Map(&oldObj, newObj.(map[string]interface{})); err != nil {
			return fmt.Errorf("cannot merge – %v", err)
		}
		if newData, err = json.Marshal(oldObj); err != nil {
			return fmt.Errorf("cannot marshal new object – %v", err)
		}
	case Array:
		if len(oldObj) > 1 {
			return fmt.Errorf("old data object contains non expected keys, cannot replace with an array")
		}
		if newData, err = json.Marshal(newObj); err != nil {
			return fmt.Errorf("cannot marshal new object – %v", err)
		}
	}

	c.Delete(branch)

	switch kw.ReturnType {
	case Object:
		if isRoot {
			if c.data, err = util.EnsureJSON(newData); err != nil {
				return err
			}
			return nil
		}
	}

	if c.data, err = jsonparser.Set(c.data, newData, branch.parent.path[1:]...); err != nil {
		return fmt.Errorf("could not set %s value of %s – %v", kw.ReturnType.String(), branch.parent.PathToString(), err)
	}
	if c.data, err = util.EnsureJSON(c.data); err != nil {
		return err
	}

	return nil

}

func addModifierLoadJSON(c *Converter, branch *BranchInfo, kw *Keyword, jsonData []byte) error {
	c.AddModifier(branch, func(c *Converter) error {
		switch branch.kind {
		case jsonparser.String:
			return doLoadJSON(c, branch, kw, jsonData)
		default:
			return fmt.Errorf("in %q value is a %s, but must be a string", branch.PathToString(), branch.kind)
		}
	})
	return nil
}

// TODO: generalise the way of passing contextual arugments

func MakeArrayLoadJSON(c *Converter, branch *BranchInfo, jsonData []byte) error {
	return addModifierLoadJSON(c, branch, LoadArrayJSON, jsonData)
}

func MakeObjectLoadJSON(c *Converter, branch *BranchInfo, jsonData []byte) error {
	return addModifierLoadJSON(c, branch, LoadObjectJSON, jsonData)
}
