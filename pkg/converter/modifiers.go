package converter

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/buger/jsonparser"
)

type MakeModifier func(*Converter, *BranchInfo, *Keyword) (ModifierCallback, error)
type ModifierCallback func(*Modifier, *Converter) error

type Keyword struct {
	ReturnType ValueType
	EvalPhase  KeywordEvalPhase
	VerbName   string
	Argument   bool
}

type UnregisteredModifier struct {
	Keyword      *Keyword
	makeModifier MakeModifier
}

type Modifier struct {
	Keyword          *Keyword
	Branch           *BranchInfo
	modifierCallback ModifierCallback
}

func (m *UnregisteredModifier) Register(c *Converter, branch *BranchInfo) (*Modifier, error) {
	cb, err := m.makeModifier(c, branch, m.Keyword)

	if err != nil {
		return nil, err
	}

	return &Modifier{
		Keyword:          m.Keyword,
		Branch:           branch,
		modifierCallback: cb,
	}, nil
}

func (m *Modifier) Do(c *Converter) error { return m.modifierCallback(m, c) }

func (c *Converter) DefineKeyword(kw *Keyword, fn MakeModifier) {
	c.keywords[kw.EvalPhase][kw.String()] = &UnregisteredModifier{kw, fn}
	c.keywordMatcher.update(kw)
}

func (c *Converter) DefineKeywordWithCallbackt(kw *Keyword, cb func() MakeModifier) {
	fn := cb()
	c.keywords[kw.EvalPhase][kw.String()] = &UnregisteredModifier{kw, fn}
	c.keywordMatcher.update(kw)
}

func (c *Converter) TypeCheckModifier(branch *BranchInfo, kind ValueType, cb ModifierCallback) (ModifierCallback, error) {
	if branch.Kind() != kind {
		return cb, fmt.Errorf("in %q value is a %s, but must be a %s", branch.PathToString(), branch.Kind(), kind)
	}
	return cb, nil
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
	if modifier, ok := c.keywords[c.keywordEvalPhase][kw]; ok {
		registered, err := modifier.Register(c, newBranch)
		if err != nil {
			errors <- fmt.Errorf("failed to register modifier for keyword %q – %v", key, err)
			return
		}
		c.modifiers[newBranch.PathToString()] = registered
	}
}

func (c *Converter) callModifiersOnce() error {
	for p, modifier := range c.modifiers {
		if err := modifier.Do(c); err != nil {
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
