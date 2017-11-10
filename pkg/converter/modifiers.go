package converter

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/buger/jsonparser"
)

func (c *Converter) DefineKeyword(kw *Keyword, fn keywordCallbackMaker) {
	c.keywords[kw.EvalPhase][kw.String()] = fn
	c.keywordMatcher.update(kw)
}

func (c *Converter) AddModifier(branch *BranchInfo, fn keywordCallback) {
	c.modifiers[branch.PathToString()] = fn
}

const validKeywordFmt = `^kubegen.(%s).(%s)(\((\S+)\))?$`

type keywordMatcher struct {
	validTypes []string
	validVerbs []string
	currentExp *regexp.Regexp
}

func newKeywordMatcher() *keywordMatcher {
	return &keywordMatcher{
		currentExp: regexp.MustCompile(fmt.Sprintf(validKeywordFmt, `\w`, `\w`)),
	}
}

func (m *keywordMatcher) update(kw *Keyword) {
	m.validTypes = append(m.validTypes, strings.Title(kw.ReturnType.String()))
	m.validVerbs = append(m.validVerbs, kw.VerbName)

	m.currentExp = regexp.MustCompile(
		fmt.Sprintf(validKeywordFmt,
			strings.Join(m.validTypes, "|"),
			strings.Join(m.validVerbs, "|"),
		))
}

func (m *keywordMatcher) isKeyword(key string) (string, string) {
	match := m.currentExp.MatchString(key)
	if !match {
		return key, ""
	}
	return key, ""
}

func (c *Converter) ifKeywordDoRegister(newBranch *BranchInfo, key string, errors chan error) {
	kw, _ := c.keywordMatcher.isKeyword(key)
	if makeModifier, ok := c.keywords[c.keywordEvalPhase][kw]; ok {
		if err := makeModifier(c, newBranch); err != nil {
			errors <- fmt.Errorf("failed to register modifier for keyword %q – %v", key, err)
			return
		}
	}
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
