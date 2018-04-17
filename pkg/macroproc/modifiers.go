package macroproc

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type MakeModifier func(*Converter, *BranchLocator, *Macro) (ModifierCallback, error)
type ModifierCallback func(*Modifier, *Converter) error

type Macro struct {
	ReturnType   ValueType
	EvalPhase    MacrosEvalPhase
	VerbName     string
	ArgumentType MacroArgumentType
}

type UnregisteredModifier struct {
	Macro        *Macro
	makeModifier MakeModifier
}

type Modifier struct {
	Macro            *Macro
	Branch           *BranchLocator
	modifierCallback ModifierCallback
}

func (m *UnregisteredModifier) Register(c *Converter, branch *BranchLocator) (*Modifier, error) {
	cb, err := m.makeModifier(c, branch, m.Macro)
	if err != nil {
		return nil, err
	}

	return &Modifier{
		Macro:            m.Macro,
		Branch:           branch,
		modifierCallback: cb,
	}, nil
}

func (m *Modifier) Do(c *Converter) error {
	// As pointers are used for all sub-trees, a refresh of
	// pointer saves from having to call the actual modifier.
	// Besides that, it's actually essential to refresh here
	// to avoid stale state.
	if err := m.Branch.Refresh(c); err != nil {
		// TODO consider adding some kind of safety net here, right now we just
		// assume we did the right thing, but we might want to check
		// log.Printf("skipping stale modifier on %s", m.Branch.PathToString())
		return nil
	}

	if err := m.modifierCallback(m, c); err != nil {
		return err
	}
	failFmt := "failed to type-check new value after macro evaluation"
	vt, err := c.tree.Check(m.Branch.path[1 : len(m.Branch.path)-1]...)
	if err != nil {
		return fmt.Errorf("%s – %v", failFmt, err)
	}
	if *vt != m.Macro.ReturnType {
		return fmt.Errorf("%s – result is a %s, not a %s", failFmt, *vt, m.Macro.ReturnType)
	}
	return nil
}
func (c *Converter) DefineMacro(m *Macro, fn MakeModifier) {
	c.macros[m.EvalPhase][m.String()] = &UnregisteredModifier{m, fn}
	c.macroMatcher.update(m)
}

func (c *Converter) DefineMacroWithCallbackt(m *Macro, cb func() MakeModifier) {
	fn := cb()
	c.macros[m.EvalPhase][m.String()] = &UnregisteredModifier{m, fn}
	c.macroMatcher.update(m)
}

func (c *Converter) TypeCheckModifier(branch *BranchLocator, kind ValueType, cb ModifierCallback) (ModifierCallback, error) {
	if branch.Kind() != kind {
		return cb, fmt.Errorf("in %q value is a %s, but must be a %s", branch.PathToString(), branch.Kind(), kind)
	}
	return cb, nil
}

const validMacroFmt = `^kubegen.(%s).(%s)(\((\S+)\))?$`

type macroMatcher struct {
	validTypes []string
	validVerbs []string
	currentExp *regexp.Regexp
}

func newMacroMatcher() *macroMatcher {
	return &macroMatcher{
		currentExp: regexp.MustCompile(fmt.Sprintf(validMacroFmt, `\w`, `\w`)),
	}
}

func (m *macroMatcher) update(macro *Macro) {
	m.validTypes = append(m.validTypes, strings.Title(macro.ReturnType.String()))
	m.validVerbs = append(m.validVerbs, macro.VerbName)

	m.currentExp = regexp.MustCompile(
		fmt.Sprintf(validMacroFmt,
			strings.Join(m.validTypes, "|"),
			strings.Join(m.validVerbs, "|"),
		))
}

func (m *macroMatcher) isMacro(key interface{}) (string, bool) {
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

func (c *Converter) ifMacroDoRegister(newBranch *BranchLocator, key interface{}, errors chan error) {
	m, _ := c.macroMatcher.isMacro(key)
	// TODO using second return value makes our tests fail
	// we still have work todo here to use regexp matcher properly
	//if !ok {
	//	return
	//}
	if modifier, ok := c.macros[c.macrosEvalPhase][m]; ok {
		registered, err := modifier.Register(c, newBranch)
		if err != nil {
			errors <- fmt.Errorf("failed to register modifier for macro %v – %v", key, err)
			return
		}
		c.modifiers[newBranch.PathToString()] = registered
	}
}

func (c *Converter) callModifiersOnce() error {
	// Our keys need to be sorted by depth
	keys := []*BranchLocator{}
	for _, modifier := range c.modifiers {
		keys = append(keys, modifier.Branch)
	}

	sort.Slice(keys, func(i, j int) bool {
		return len(keys[i].path) > len(keys[j].path)
	})

	for x := range keys {
		p := keys[x].PathToString()
		// log.Printf("calling %s", p)
		modifier := c.modifiers[p]
		if err := modifier.Do(c); err != nil {
			// TODO should put more info in the error (perhaps use github.com/pkg/errors)
			// as this is probably the most verbose kind of error user may see
			return fmt.Errorf("callback on %s failed to modify the tree – %v", p, err)
		}
		delete(c.modifiers, p)
	}
	return nil
}

func (c *Converter) Set(branch *BranchLocator, value interface{}) error {
	if err := c.tree.Set(value, branch.parent.path[1:]...); err != nil {
		return fmt.Errorf("failed to set %v – %v", value, err)
	}
	return nil
}

func (c *Converter) Overlay(branch *BranchLocator, value interface{}) error {
	if err := c.Delete(branch); err != nil {
		return err
	}
	if err := c.tree.Overlay(NewTree(&value), branch.parent.path[1:]...); err != nil {
		return fmt.Errorf("failed to overlay %v onto %v – %v", value, branch.parent.value, err)
	}
	return nil
}
func (c *Converter) Delete(branch *BranchLocator) error {
	if err := c.tree.Delete(branch.path[1:]...); err != nil {
		return fmt.Errorf("failed to delete %v – %v", branch.value, err)
	}
	return nil
}
