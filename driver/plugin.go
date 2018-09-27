package driver

import (
	"fmt"
	"git.xdrm.io/go/aicra/err"
	"git.xdrm.io/go/aicra/response"
	"plugin"
	"strings"
)

// Load implements the Driver interface
func (d *Plugin) Load(_path []string, _method string) (func(response.Arguments, *response.Response) response.Response, err.Error) {

	/* (1) Build controller path */
	path := strings.Join(_path, "-")
	if len(path) == 0 {
		path = fmt.Sprintf(".build/controller/ROOT.so")
	} else {
		path = fmt.Sprintf(".build/controller/%s.so", path)
	}

	/* (2) Format url */
	tmp := []byte(strings.ToLower(_method))
	tmp[0] = tmp[0] - ('a' - 'A')
	method := string(tmp)

	/* (2) Try to load plugin */
	p, err2 := plugin.Open(path)
	if err2 != nil {
		return nil, err.UncallableController
	}

	/* (3) Try to extract method */
	m, err2 := p.Lookup(method)
	if err2 != nil {
		return nil, err.UncallableMethod
	}

	/* (4) Check signature */
	callable, validSignature := m.(func(response.Arguments, *response.Response) response.Response)
	if !validSignature {
		return nil, err.UncallableMethod
	}

	return callable, err.Success
}
