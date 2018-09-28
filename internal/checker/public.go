package checker

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"plugin"
	"strings"
)

var ErrNoMatchingType = errors.New("no matching type")
var ErrDoesNotMatch = errors.New("does not match")
var ErrEmptyTypeName = errors.New("type name must not be empty")

// CreateRegistry creates an empty type registry
func CreateRegistry(_folder string) *Registry {

	/* (1) Create registry */
	reg := &Registry{
		Types: make([]Type, 0),
	}

	/* (2) List types */
	files, err := ioutil.ReadDir(_folder)
	if err != nil {
		log.Fatal(err)
	}

	/* (3) Else try to load each given default */
	for _, file := range files {

		// ignore non .so files
		if !strings.HasSuffix(file.Name(), ".so") {
			continue
		}

		err := reg.add(file.Name())
		if err != nil {
			log.Printf("cannot load plugin '%s'", file.Name())
		}

	}

	return reg
}

// add adds a type to the registry; it must be a
// valid and existing plugin name with or without the .so extension
// it must be located in the relative directory ./types
func (reg *Registry) add(pluginName string) error {

	/* (1) Check plugin name */
	if len(pluginName) < 1 {
		return ErrEmptyTypeName
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
		return ErrNoMatchingType
	}

	/* (3) Check */
	if !T.Check(value) {
		return ErrDoesNotMatch
	}

	return nil

}
