package driver

import (
	"encoding/json"
	"fmt"
	"git.xdrm.io/go/aicra/err"
	"git.xdrm.io/go/aicra/response"
	"os/exec"
	"strings"
)

// Load implements the Driver interface
func (d *Generic) Load(_path []string, _method string) (func(response.Arguments) response.Response, err.Error) {

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
		stdin, err2 := json.Marshal(d)
		if err2 != nil {
			res.Err = err.UncallableController
			return *res
		}

		/* (2) Try to load command with <stdin> -> stdout */
		cmd := exec.Command(path, string(stdin))

		stdout, err2 := cmd.Output()
		if err2 != nil {
			res.Err = err.UncallableController
			return *res
		}

		/* (3) Get output json */
		output := make(response.Arguments)
		err2 = json.Unmarshal(stdout, output)

		res.Err = err.Success

		// extract error (success by default or on error)
		if outErr, ok := output["error"]; ok {
			tmpErr, ok := outErr.(err.Error)
			if ok {
				res.Err = tmpErr
			}

			delete(output, "error")
		}

		/* (4) fill response */
		for k, v := range output {
			res.Set(k, v)
		}

		return *res

	}, err.Success
}
