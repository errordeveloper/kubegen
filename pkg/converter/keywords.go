package converter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/ghodss/yaml"
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
