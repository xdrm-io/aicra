package checker

import (
	"fmt"
	"io/ioutil"
	"log"
	"plugin"
	"strings"
)

// CreateRegistry creates an empty type registry
// - if loadDir is True if will load all available types
//   inside the local ./types folder
func CreateRegistry(loadDir ...string) *Registry {

	/* (1) Create registry */
	reg := &Registry{
		Types: make([]Type, 0),
	}

	/* (2) If no default to use -> empty registry */
	if len(loadDir) < 1 {
		return reg
	}

	/* (3) List types */
	plugins, err := ioutil.ReadDir(loadDir[0])
	if err != nil {
		log.Fatal(err)
	}

	/* (4) Else try to load each given default */
	for _, file := range plugins {

		// ignore non .so files
		if !strings.HasSuffix(file.Name(), ".so") {
			continue
		}

		err := reg.Add(file.Name())
		if err != nil {
			log.Fatalf("Cannot load plugin '%s'", file.Name())
		}

	}

	return reg
}

// Add adds a type to the registry; it must be a
// valid and existing plugin name with or without the .so extension
// it must be located in the relative directory ./types
func (reg *Registry) Add(pluginName string) error {

	/* (1) Check plugin name */
	if len(pluginName) < 1 {
		return fmt.Errorf("Plugin name must not be empty")
	}

	/* (2) Check if valid plugin name */
	if strings.ContainsAny(pluginName, "/") {
		return fmt.Errorf("'%s' can only be a name, not a path", pluginName)
	}

	/* (3) Check plugin extension */
	if !strings.HasSuffix(pluginName, ".so") {
		pluginName = fmt.Sprintf("%s.so", pluginName)
	}

	/* (4) Try to load the plugin */
	p, err := plugin.Open(fmt.Sprintf(".build/type/%s", pluginName))
	if err != nil {
		return err
	}

	/* (5) Export wanted properties */
	matcher, err := p.Lookup("Match")
	if err != nil {
		return fmt.Errorf("Missing method 'Match()'; %s", err)
	}

	checker, err := p.Lookup("Check")
	if err != nil {
		return fmt.Errorf("Missing method 'Check()'; %s", err)
	}

	/* (6) Cast Match+Check */
	matcherCast, ok := matcher.(func(string) bool)
	if !ok {
		return fmt.Errorf("Match() is malformed")
	}

	checkerCast, ok := checker.(func(interface{}) bool)
	if !ok {
		return fmt.Errorf("Check() is malformed")
	}

	/* (7) Add type to registry */
	reg.Types = append(reg.Types, Type{
		Match: matcherCast,
		Check: checkerCast,
	})

	return nil
}

// Run finds a type checker from the registry matching the type @typeName
// and uses this checker to check the @value. If no type checker matches
// the @typeName name, error is returned by default.
func (reg Registry) Run(typeName string, value interface{}) error {

	var T *Type

	/* (1) Iterate to find matching type (take first) */
	for _, t := range reg.Types {

		// stop if found
		if t.Match(typeName) {
			T = &t
			break
		}
	}

	/* (2) Abort if no matching type */
	if T == nil {
		return fmt.Errorf("No matching type")
	}

	/* (3) Check */
	if !T.Check(value) {
		return fmt.Errorf("Does not match")
	}

	return nil

}
