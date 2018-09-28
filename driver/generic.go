package driver

import (
	"encoding/json"
	"fmt"
	e "git.xdrm.io/go/aicra/err"
	"git.xdrm.io/go/aicra/response"
	"os/exec"
	"strings"
)

// Load implements the Driver interface
func (d *Generic) Load(_path []string, _method string) (func(response.Arguments) response.Response, e.Error) {

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
				res.Err = e.Error{int(errCode), "unknown reason", nil}
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
