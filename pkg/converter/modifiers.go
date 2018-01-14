package converter

import (
	"fmt"
	"regexp"
	"strings"
)

type MakeModifier func(*Converter, *BranchLocator, *Keyword) (ModifierCallback, error)
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
	Branch           *BranchLocator
	modifierCallback ModifierCallback
}

func (m *UnregisteredModifier) Register(c *Converter, branch *BranchLocator) (*Modifier, error) {
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

func (m *Modifier) Do(c *Converter) error {
	return m.modifierCallback(m, c)
}

func (c *Converter) DefineKeyword(kw *Keyword, fn MakeModifier) {
	c.keywords[kw.EvalPhase][kw.String()] = &UnregisteredModifier{kw, fn}
	c.keywordMatcher.update(kw)
}

func (c *Converter) DefineKeywordWithCallbackt(kw *Keyword, cb func() MakeModifier) {
	fn := cb()
	c.keywords[kw.EvalPhase][kw.String()] = &UnregisteredModifier{kw, fn}
	c.keywordMatcher.update(kw)
}

func (c *Converter) TypeCheckModifier(branch *BranchLocator, kind ValueType, cb ModifierCallback) (ModifierCallback, error) {
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

func (m *keywordMatcher) isKeyword(key interface{}) (string, bool) {
	k, ok := key.(string)
	if !ok {
		return k, false
	}
	match := m.currentExp.MatchString(k)
	if !match {
		return k, false
	}
	return k, true
}

func (c *Converter) ifKeywordDoRegister(newBranch *BranchLocator, key interface{}, errors chan error) {
	kw, _ := c.keywordMatcher.isKeyword(key)
	// TODO using second return value makes our tests fail
	// we still have work todo here to use regexp matcher properly
	//if !ok {
	//	return
	//}
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
			return fmt.Errorf("callback on %s failed to modify the tree – %v", p, err)
		}
		delete(c.modifiers, p)
	}
	return nil
}

func (c *Converter) Set(branch *BranchLocator, value interface{}) error {
	if err := c.tree.Set(value, branch.parent.path[1:]...); err != nil {
		return fmt.Errorf("failed to set `%q:%s` [%v] – %v", branch.PathToString(), branch.path, value, err)
	}
	return nil
}

func (c *Converter) Delete(branch *BranchLocator) error {
	if err := c.tree.Delete(branch.path[1:]...); err != nil {
		return fmt.Errorf("failed to delete `%q` [%v] – %v", branch.PathToString(), branch.path, err)
	}
	return nil
}
