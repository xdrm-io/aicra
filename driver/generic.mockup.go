package driver

import (
	"encoding/json"
	e "git.xdrm.io/go/aicra/err"
	"git.xdrm.io/go/aicra/response"
	"net/http"
	"os/exec"
	"strings"
)

// genericController is the mockup for returning a controller with as a string the path
type genericController string

func (path genericController) Get(d response.Arguments) response.Response {

	res := response.New()

	/* (1) Prepare stdin data */
	stdin, err := json.Marshal(d)
	if err != nil {
		res.Err = e.UncallableController
		return *res
	}

	// extract HTTP method
	rawMethod, ok := d["_HTTP_METHOD_"]
	if !ok {
		res.Err = e.UncallableController
		return *res
	}
	method, ok := rawMethod.(string)
	if !ok {
		res.Err = e.UncallableController
		return *res
	}

	/* (2) Try to load command with <stdin> -> stdout */
	cmd := exec.Command(string(path), method, string(stdin))

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

}

func (path genericController) Post(d response.Arguments) response.Response {
	return path.Get(d)
}
func (path genericController) Put(d response.Arguments) response.Response {
	return path.Get(d)
}
func (path genericController) Delete(d response.Arguments) response.Response {
	return path.Get(d)
}

// genericMiddleware is the mockup for returning a middleware as a string (its path)
type genericMiddleware string

func (path genericMiddleware) Inspect(_req http.Request, _scope *[]string) {

	/* (1) Prepare stdin data */
	stdin, err := json.Marshal(_scope)
	if err != nil {
		return
	}

	/* (2) Try to load command with <stdin> -> stdout */
	cmd := exec.Command(string(path), string(stdin))

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

}

// genericChecker is the mockup for returning a checker as a string (its path)
type genericChecker string

func (path genericChecker) Match(_type string) bool {

	/* (1) Try to load command with <stdin> -> stdout */
	cmd := exec.Command(string(path), "MATCH", _type)

	stdout, err := cmd.Output()
	if err != nil {
		return false
	}

	/* (2) Parse output */
	output := strings.ToLower(strings.Trim(string(stdout), " \t\r\n"))

	return output == "true" || output == "1"

}
func (path genericChecker) Check(_value interface{}) bool {

	/* (1) Prepare stdin data */
	indata := make(map[string]interface{})
	indata["value"] = _value

	stdin, err := json.Marshal(indata)
	if err != nil {
		return false
	}

	/* (2) Try to load command with <stdin> -> stdout */
	cmd := exec.Command(string(path), "CHECK", string(stdin))

	stdout, err := cmd.Output()
	if err != nil {
		return false
	}

	/* (2) Parse output */
	output := strings.ToLower(strings.Trim(string(stdout), " \t\r\n"))

	return output == "true" || output == "1"

}
