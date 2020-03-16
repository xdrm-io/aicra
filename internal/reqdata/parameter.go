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
func (i *Parameter) Parse() error {

	/* (1) Stop if already parsed or nil*/
	if i.Parsed || i.Value == nil {
		return nil
	}

	/* (2) Try to parse value */
	parsed, err := parseParameter(i.Value)
	if err != nil {
		return err
	}

	i.Parsed = true
	i.Value = parsed

	return nil
}

// parseParameter parses http GET/POST data
// - []string
//		- size = 1 : return json of first element
//		- size > 1 : return array of json elements
// - string : return json if valid, else return raw string
func parseParameter(data interface{}) (interface{}, error) {
	dtype := reflect.TypeOf(data)
	dvalue := reflect.ValueOf(data)

	switch dtype.Kind() {

	/* (1) []string -> recursive */
	case reflect.Slice:

		// 1. ignore empty
		if dvalue.Len() == 0 {
			return data, nil
		}

		// 2. parse each element recursively
		result := make([]interface{}, dvalue.Len())

		for i, l := 0, dvalue.Len(); i < l; i++ {
			element := dvalue.Index(i)

			// ignore non-string
			if element.Kind() != reflect.String {
				result[i] = element.Interface()
				continue
			}

			parsed, err := parseParameter(element.String())
			if err != nil {
				return data, err
			}
			result[i] = parsed
		}
		return result, nil

	/* (2) string -> parse */
	case reflect.String:

		// build json wrapper
		wrapper := fmt.Sprintf("{\"wrapped\":%s}", dvalue.String())

		// try to parse as json
		var result interface{}
		err := json.Unmarshal([]byte(wrapper), &result)

		// return if success
		if err == nil {

			mapval, ok := result.(map[string]interface{})
			if !ok {
				return dvalue.String(), ErrInvalidRootType
			}

			wrapped, ok := mapval["wrapped"]
			if !ok {
				return dvalue.String(), ErrInvalidJSON
			}

			return wrapped, nil
		}

		// else return as string
		return dvalue.String(), nil

	}

	/* (3) NIL if unknown type */
	return dvalue.Interface(), nil

}
