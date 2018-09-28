package middleware

import (
	"git.xdrm.io/go/aicra/driver"
	"io/ioutil"
	"log"
	"net/http"
	"path"
)

// CreateRegistry creates an empty middleware registry
// - if loadDir is set -> load all available middlewares
//   inside the local ./middleware folder
func CreateRegistry(_driver driver.Driver, loadDir ...string) *Registry {

	/* (1) Create registry */
	reg := &Registry{
		Middlewares: make([]*Wrapper, 0),
	}

	/* (2) If no default to use -> empty registry */
	if len(loadDir) < 1 {
		return reg
	}

	/* (3) List middleware files */
	files, err := ioutil.ReadDir(loadDir[0])
	if err != nil {
		log.Fatal(err)
	}

	/* (4) Else try to load each given default */
	for _, file := range files {

		mwFunc, err := _driver.LoadMiddleware(path.Join(loadDir[0], file.Name()))
		if err != nil {
			log.Printf("Cannot load middleware '%s' | %s", file.Name(), err)
		}
		reg.Middlewares = append(reg.Middlewares, &Wrapper{Inspect: mwFunc})

	}

	return reg
}

// Run executes all middlewares (default browse order)
func (reg Registry) Run(req http.Request) []string {

	/* (1) Initialise scope */
	scope := make([]string, 0)

	/* (2) Execute each middleware */
	for _, m := range reg.Middlewares {
		m.Inspect(req, &scope)
	}

	return scope

}
