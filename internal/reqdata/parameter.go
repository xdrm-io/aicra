package reqdata

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Parameter represents an http request parameter
// that can be of type URL, GET, or FORM (multipart, json, urlencoded)
type Parameter struct {
	// whether the value has been json-parsed
	// for optimisation purpose, parameters are only parsed
	// if they are required by the current service
	Parsed bool

	// whether the value is a file
	File bool

	// the actual parameter value
	Value interface{}
}

// Parse parameter (json-like) if not already done
func (i *Parameter) Parse() {

	/* ignore already parsed or nil*/
	if i.Parsed || i.Value == nil {
		return
	}

	/* parse value */
	i.Parsed = true
	i.Value = parseParameter(i.Value)
}

// parseParameter parses http URI/GET/POST data
// - []string : return array of json elements
// - string   : return json if valid, else return raw string
func parseParameter(data interface{}) interface{} {
	dtype := reflect.TypeOf(data)
	dvalue := reflect.ValueOf(data)

	switch dtype.Kind() {

	/* (1) []string -> recursive */
	case reflect.Slice:

		// 1. ignore empty
		if dvalue.Len() == 0 {
			return data
		}

		// 2. parse each element recursively
		result := make([]interface{}, dvalue.Len())

		for i, l := 0, dvalue.Len(); i < l; i++ {
			element := dvalue.Index(i)
			result[i] = parseParameter(element.Interface())
		}
		return result

	/* (2) string -> parse */
	case reflect.String:

		// build json wrapper
		wrapper := fmt.Sprintf("{\"wrapped\":%s}", dvalue.String())

		// try to parse as json
		var result interface{}
		err := json.Unmarshal([]byte(wrapper), &result)

		// return if success
		if err != nil {
			return dvalue.String()
		}

		mapval, ok := result.(map[string]interface{})
		if !ok {
			return dvalue.String()
		}

		wrapped, ok := mapval["wrapped"]
		if !ok {
			return dvalue.String()
		}

		return wrapped

	}

	/* (3) NIL if unknown type */
	return dvalue.Interface()

}
