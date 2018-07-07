package middleware

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"plugin"
	"strings"
)

// CreateRegistry creates an empty middleware registry
// - if loadDir is set -> load all available middlewares
//   inside the local ./middleware folder
func CreateRegistry(loadDir ...string) *MiddlewareRegistry {

	/* (1) Create registry */
	reg := &MiddlewareRegistry{
		Middlewares: make([]MiddleWare, 0),
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

// Add adds a middleware to the registry; it must be a
// valid and existing plugin name with or without the .so extension
// it must be located in the relative directory .build/middleware
func (tr *MiddlewareRegistry) Add(pluginName string) error {

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
	p, err := plugin.Open(fmt.Sprintf(".build/middleware/%s", pluginName))
	if err != nil {
		return err
	}

	/* (5) Export wanted properties */
	inspect, err := p.Lookup("Inspect")
	if err != nil {
		return fmt.Errorf("Missing method 'Inspect()'; %s", err)
	}

	/* (6) Cast Inspect */
	inspectCast, ok := inspect.(func(http.Request, Scope))
	if !ok {
		return fmt.Errorf("Inspect() is malformed")
	}

	/* (7) Add type to registry */
	tr.Middlewares = append(tr.Middlewares, MiddleWare{
		Inspect: inspectCast,
	})

	return nil
}

// Runs all middlewares (default browse order)
func (mr MiddlewareRegistry) Run(req http.Request) Scope {

	/* (1) Initialise scope */
	scope := Scope{}

	/* (2) Execute each middleware */
	for _, m := range mr.Middlewares {
		m.Inspect(req, scope)
	}

	return scope

}
