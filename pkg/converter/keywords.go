package converter

import (
	"encoding/json"
	"fmt"
	"strings"

	yaml "gopkg.in/yaml.v2"
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

var (
	KeywordBooleanIf = &Keyword{
		ReturnType: Null,
		EvalPhase:  KeywordEvalPhaseA,
		VerbName:   "If",
	}

	KeywordBooleanLookup = &Keyword{
		ReturnType: Boolean,
		EvalPhase:  KeywordEvalPhaseB,
		VerbName:   "Lookup",
	}
	KeywordStringLookup = &Keyword{
		ReturnType: String,
		EvalPhase:  KeywordEvalPhaseB,
		VerbName:   "Lookup",
	}
	KeywordNumberLookup = &Keyword{
		ReturnType: Number,
		EvalPhase:  KeywordEvalPhaseB,
		VerbName:   "Lookup",
	}

	KeywordArrayLookup = &Keyword{
		ReturnType: Array,
		EvalPhase:  KeywordEvalPhaseC,
		VerbName:   "Lookup",
	}
	KeywordObjectLookup = &Keyword{
		ReturnType: Object,
		EvalPhase:  KeywordEvalPhaseC,
		VerbName:   "Lookup",
	}

	KeywordStringJoin = &Keyword{
		ReturnType: String,
		EvalPhase:  KeywordEvalPhaseC,
		VerbName:   "Join",
	}

	KeywordStringAsJSON = &Keyword{
		ReturnType: String,
		EvalPhase:  KeywordEvalPhaseD,
		VerbName:   "AsJSON",
	}
	KeywordStringAsYAML = &Keyword{
		ReturnType: String,
		EvalPhase:  KeywordEvalPhaseD,
		VerbName:   "AsYAML",
	}

	LoadObjectJSON = &Keyword{
		ReturnType: Object,
		EvalPhase:  KeywordEvalPhaseA,
		VerbName:   "LoadJSON",
	}

	LoadArrayJSON = &Keyword{
		ReturnType: Array,
		EvalPhase:  KeywordEvalPhaseA,
		VerbName:   "LoadJSON",
	}
)

func (kw *Keyword) String() string {
	// TODO maybe add args if given, so we get nicer error messages?
	return fmt.Sprintf("kubegen.%s.%s", kw.ReturnType.String(), kw.VerbName)
}

func MakeModifierStringJoin(c *Converter, branch *BranchLocator, _ *Keyword) (ModifierCallback, error) {
	cb := func(m *Modifier, c *Converter) error {
		x := []string{}
		branch.Value().ArrayEach(func(_ int, value interface{}, dataType ValueType, _ *Tree) error {
			x = append(x, fmt.Sprintf("%v", value))
			return nil
		})
		if err := c.Set(branch, strings.Join(x, "")); err != nil {
			return fmt.Errorf("could not join string – %v", err)
		}
		return nil
	}
	return c.TypeCheckModifier(branch, Array, cb)
}

func MakeModifierStringAsYAML(_ *Converter, _ *BranchLocator, _ *Keyword) (ModifierCallback, error) {
	cb := func(m *Modifier, c *Converter) error {
		o := new(interface{})
		if err := json.Unmarshal(m.Branch.Value().Bytes(), o); err != nil {
			return err
		}
		x, err := yaml.Marshal(o)
		if err != nil {
			return err
		}
		{
			if err := c.Set(m.Branch, string(x)); err != nil {
				return err
			}
			return nil
		}
	}
	return cb, nil
}

func MakeModifierStringAsJSON(_ *Converter, _ *BranchLocator, _ *Keyword) (ModifierCallback, error) {
	cb := func(m *Modifier, c *Converter) error {
		if err := c.Set(m.Branch, m.Branch.Value().String()); err != nil {
			return err
		}
		return nil
	}
	return cb, nil
}

func doLoadJSON(c *Converter, branch *BranchLocator, kw *Keyword, newData []byte) error {
	/*
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
			//if err := mergo.MergeWithOverwrite(&oldObj, newObj.(map[string]interface{})); err != nil {
			if err := mergo.Merge(&oldObj, newObj.(map[string]interface{})); err != nil {
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
	*/
	return nil
}

func addModifierLoadJSON(c *Converter, branch *BranchLocator, _ *Keyword, jsonData []byte) (ModifierCallback, error) {
	cb := func(m *Modifier, c *Converter) error {
		return doLoadJSON(c, m.Branch, m.Keyword, jsonData)
	}
	return c.TypeCheckModifier(branch, String, cb)
}

// TODO: generalise the way of passing contextual arugments - or is it better now?

func MakeArrayLoadJSON(c *Converter, branch *BranchLocator, jsonData []byte) (ModifierCallback, error) {
	return addModifierLoadJSON(c, branch, LoadArrayJSON, jsonData)
}

func MakeObjectLoadJSON(c *Converter, branch *BranchLocator, jsonData []byte) (ModifierCallback, error) {
	return addModifierLoadJSON(c, branch, LoadObjectJSON, jsonData)
}
