package converter

import (
	"encoding/json"
	"fmt"

	"github.com/buger/jsonparser"
)

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
			// TODO should put more info in the error (perhaps use github.com/pkg/errors)
			return fmt.Errorf("callback on %q failed to modify the tree – %v", p, err)
		}
		delete(c.modifiers, p)
	}
	return nil
}

func (c *Converter) ifKeywordDoRegister(newBranch *BranchInfo, key string, errors chan error) {
	if makeModifier, ok := c.keywords[c.keywordEvalPhase][key]; ok {
		if err := makeModifier(c, newBranch); err != nil {
			errors <- fmt.Errorf("failed to register modifier for keyword %q – %v", key, err)
			return
		}
	}
}

func (c *Converter) doSet(branch *BranchInfo, value []byte) error {
	x, err := jsonparser.Set(c.data, value, branch.parent.path[1:]...)
	if err != nil {
		return err
	}
	c.loadStrict(x)
	return nil
}

func (c *Converter) Set(branch *BranchInfo, value interface{}) error {
	v, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal `%q:%#v` – %v", branch.PathToString(), value, err)
	}
	if err := c.doSet(branch, v); err != nil {
		return fmt.Errorf("failed to set `%q:%s` – %v", branch.PathToString(), value, err)
	}
	return nil
}

func (c *Converter) Delete(branch *BranchInfo) {
	c.data = jsonparser.Delete(c.data, branch.path[1:]...)
}
