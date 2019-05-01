package reqdata

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// parseParameter parses http GET/POST data
// - []string
//		- size = 1 : return json of first element
//		- size > 1 : return array of json elements
// - string : return json if valid, else return raw string
func parseParameter(data interface{}) interface{} {
	dtype := reflect.TypeOf(data)
	dvalue := reflect.ValueOf(data)

	switch dtype.Kind() {

	/* (1) []string -> recursive */
	case reflect.Slice:

		// 1. Return nothing if empty
		if dvalue.Len() == 0 {
			return nil
		}

		// 2. only return first element if alone
		if dvalue.Len() == 1 {

			element := dvalue.Index(0)
			if element.Kind() != reflect.String {
				return nil
			}
			return parseParameter(element.String())

		}

		// 3. Return all elements if more than 1
		result := make([]interface{}, dvalue.Len())

		for i, l := 0, dvalue.Len(); i < l; i++ {
			element := dvalue.Index(i)

			// ignore non-string
			if element.Kind() != reflect.String {
				continue
			}

			result[i] = parseParameter(element.String())
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
		if err == nil {

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

		// else return as string
		return dvalue.String()

	}

	/* (3) NIL if unknown type */
	return dvalue

}
