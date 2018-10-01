package driver

import (
	"encoding/json"
	"fmt"
	e "git.xdrm.io/go/aicra/err"
	"git.xdrm.io/go/aicra/response"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
)

// Name returns the driver name
func (d *Generic) Name() string { return "generic" }

// Path returns the universal path from the source path
func (d Generic) Path(_root, _folder, _src string) string {
	return _src
}

// Source returns the source path from the universal path
func (d Generic) Source(_root, _folder, _path string) string {
	return filepath.Join(_root, _folder, _path)

}

// Build returns the build path from the universal path
func (d Generic) Build(_root, _folder, _path string) string {
	return filepath.Join(_root, _folder, _path)
}

// Compiled returns whether the driver has to be build
func (d Generic) Compiled() bool { return false }

// RunController implements the Driver interface
func (d *Generic) RunController(_path []string, _method string) (func(response.Arguments) response.Response, e.Error) {

	/* (1) Build controller path */
	path := strings.Join(_path, "-")
	if len(path) == 0 {
		path = fmt.Sprintf("./controller/ROOT")
	} else {
		path = fmt.Sprintf("./controller/%s", path)
	}

	/* (2) Format method */
	method := strings.ToUpper(_method)

	return func(d response.Arguments) response.Response {

		res := response.New()

		/* (1) Prepare stdin data */
		d["_HTTP_METHOD_"] = method
		stdin, err := json.Marshal(d)
		if err != nil {
			res.Err = e.UncallableController
			return *res
		}

		/* (2) Try to load command with <stdin> -> stdout */
		cmd := exec.Command(path, string(stdin))

		stdout, err := cmd.Output()
		if err != nil {
			res.Err = e.UncallableController
			return *res
		}

		/* (3) Get output json */
		var outputI interface{}
		err = json.Unmarshal(stdout, &outputI)
		if err != nil {
			res.Err = e.UncallableController
			return *res
		}

		output, ok := outputI.(map[string]interface{})
		if !ok {
			res.Err = e.UncallableController
			return *res
		}

		res.Err = e.Success

		// extract error (success by default or on error)
		if outErr, ok := output["error"]; ok {
			errCode, ok := outErr.(float64)
			if ok {
				res.Err = e.Error{Code: int(errCode), Reason: "unknown reason", Arguments: nil}
			}

			delete(output, "error")
		}

		/* (4) fill response */
		for k, v := range output {
			res.Set(k, v)
		}

		return *res

	}, e.Success
}

// LoadMiddleware returns a new middleware function; it must be a
// valid and existing folder/filename file
func (d *Generic) LoadMiddleware(_path string) (func(http.Request, *[]string), error) {

	/* (1) Check plugin name */
	if len(_path) < 1 {
		return nil, fmt.Errorf("Middleware name must not be empty")
	}

	/* (2) Create method + error  */
	return func(_req http.Request, _scope *[]string) {

		/* (1) Prepare stdin data */
		stdin, err := json.Marshal(_scope)
		if err != nil {
			return
		}

		/* (2) Try to load command with <stdin> -> stdout */
		cmd := exec.Command(_path, string(stdin))

		stdout, err := cmd.Output()
		if err != nil {
			return
		}

		/* (3) Get output json */
		var outputI interface{}
		err = json.Unmarshal(stdout, &outputI)
		if err != nil {
			return
		}

		/* (4) Get as []string */
		scope, ok := outputI.([]interface{})
		if !ok {
			return
		}

		/* (5) Try to add each value to the scope */
		for _, v := range scope {
			stringScope, ok := v.(string)
			if !ok {
				continue
			}
			*_scope = append(*_scope, stringScope)
		}

	}, nil

}
