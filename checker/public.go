package checker

import (
	"fmt"
	"plugin"
	"strings"
)

// CreateRegistry creates an empty type registry
// if TRUE is given, it will use default Types
// see ./default_types.go
func CreateRegistry(useDefaultTypes bool) *TypeRegistry {
	return &TypeRegistry{
		Types: make([]Type, 0),
	}
}

// Add adds a type to the registry; it must be a
// valid and existing plugin name without the .so extension
func (tr *TypeRegistry) Add(pluginName string) error {

	/* (1) Check plugin name */
	if len(pluginName) < 1 {
		return fmt.Errorf("Plugin name must not be empty")
	}

	/* (2) Check plugin extension */
	if strings.HasSuffix(pluginName, ".so") {
		return fmt.Errorf("Plugin name must be provided without extension: '%s'", pluginName[0:len(pluginName)-3])
	}

	/* (3) Try to load the plugin */
	p, err := plugin.Open(fmt.Sprintf("%s.so", pluginName))
	if err != nil {
		return err
	}

	/* (4) Export wanted properties */
	matcher, err := p.Lookup("Match")
	if err != nil {
		return fmt.Errorf("Missing method 'Match()'; %s", err)
	}

	checker, err := p.Lookup("Check")
	if err != nil {
		return fmt.Errorf("Missing method 'Check()'; %s", err)
	}

	/* (5) Cast Match+Check */
	matcherCast, ok := matcher.(func(string) bool)
	if !ok {
		return fmt.Errorf("Match() is malformed")
	}

	checkerCast, ok := checker.(func(interface{}) bool)
	if !ok {
		return fmt.Errorf("Check() is malformed")
	}

	/* (6) Add type to registry */
	tr.Types = append(tr.Types, Type{
		Match: matcherCast,
		Check: checkerCast,
	})

	return nil
}

// Checks the 'value' which must be of type 'name'
func (tr TypeRegistry) Run(name string, value interface{}) bool {

	var T *Type = nil

	/* (1) Iterate to find matching type (take first) */
	for _, t := range tr.Types {

		// stop if found
		if t.Match(name) {
			T = &t
			break
		}

		// else log
		fmt.Printf("does not match\n")
	}

	/* (2) Abort if no matching type */
	if T == nil {
		return false
	}

	/* (3) Check */
	fmt.Printf("Check is %t\n", T.Check(value))
	return T.Check(value)

}
