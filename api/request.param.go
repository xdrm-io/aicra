package api

import (
	"fmt"
)

// ConstError is a wrapper to set constant errors
type ConstError string

// Error implements error
func (err ConstError) Error() string {
	return string(err)
}

// ErrReqParamNotFound is thrown when a request parameter is not found
const ErrReqParamNotFound = ConstError("request parameter not found")

// ErrReqParamNotType is thrown when a request parameter is not asked with the right type
const ErrReqParamNotType = ConstError("request parameter does not fulfills type")

// RequestParam defines input parameters of an api request
type RequestParam map[string]interface{}

// Get returns the raw value (not typed) and an error if not found
func (rp RequestParam) Get(key string) (interface{}, error) {
	rawValue, found := rp[key]
	if !found {
		return "", ErrReqParamNotFound
	}
	return rawValue, nil
}

// GetString returns a string and an error if not found or invalid type
func (rp RequestParam) GetString(key string) (string, error) {
	rawValue, err := rp.Get(key)
	if err != nil {
		return "", err
	}

	switch cast := rawValue.(type) {
	case fmt.Stringer:
		return cast.String(), nil
	case []byte:
		return string(cast), nil
	case string:
		return cast, nil
	default:
		return "", ErrReqParamNotType
	}
}

// GetFloat returns a float64 and an error if not found or invalid type
func (rp RequestParam) GetFloat(key string) (float64, error) {
	rawValue, err := rp.Get(key)
	if err != nil {
		return 0, err
	}

	switch cast := rawValue.(type) {
	case float32:
		return float64(cast), nil
	case float64:
		return cast, nil
	case int, int8, int16, int32, int64:
		intVal, ok := cast.(int)
		if !ok || intVal != int(float64(intVal)) {
			return 0, ErrReqParamNotType
		}
		return float64(intVal), nil
	case uint, uint8, uint16, uint32, uint64:
		uintVal, ok := cast.(uint)
		if !ok || uintVal != uint(float64(uintVal)) {
			return 0, ErrReqParamNotType
		}
		return float64(uintVal), nil
	default:
		return 0, ErrReqParamNotType
	}
}

// GetInt returns an int and an error if not found or invalid type
func (rp RequestParam) GetInt(key string) (int, error) {
	rawValue, err := rp.Get(key)
	if err != nil {
		return 0, err
	}

	switch cast := rawValue.(type) {
	case float32, float64:
		floatVal, ok := cast.(float64)
		if !ok || floatVal < 0 || floatVal != float64(int(floatVal)) {
			return 0, ErrReqParamNotType
		}
		return int(floatVal), nil
	case int, int8, int16, int32, int64:
		intVal, ok := cast.(int)
		if !ok || intVal != int(int(intVal)) {
			return 0, ErrReqParamNotType
		}
		return int(intVal), nil
	default:
		return 0, ErrReqParamNotType
	}
}

// GetUint returns an uint and an error if not found or invalid type
func (rp RequestParam) GetUint(key string) (uint, error) {
	rawValue, err := rp.Get(key)
	if err != nil {
		return 0, err
	}

	switch cast := rawValue.(type) {
	case float32, float64:
		floatVal, ok := cast.(float64)
		if !ok || floatVal < 0 || floatVal != float64(uint(floatVal)) {
			return 0, ErrReqParamNotType
		}
		return uint(floatVal), nil
	case int, int8, int16, int32, int64:
		intVal, ok := cast.(int)
		if !ok || intVal != int(uint(intVal)) {
			return 0, ErrReqParamNotType
		}
		return uint(intVal), nil
	case uint, uint8, uint16, uint32, uint64:
		uintVal, ok := cast.(uint)
		if !ok {
			return 0, ErrReqParamNotType
		}
		return uintVal, nil
	default:
		return 0, ErrReqParamNotType
	}
}

// GetStrings returns an []slice and an error if not found or invalid type
func (rp RequestParam) GetStrings(key string) ([]string, error) {
	rawValue, err := rp.Get(key)
	if err != nil {
		return nil, err
	}

	switch cast := rawValue.(type) {
	case []fmt.Stringer:
		strings := make([]string, len(cast))
		for i, stringer := range cast {
			strings[i] = stringer.String()
		}
		return strings, nil
	case [][]byte:
		strings := make([]string, len(cast))
		for i, bytes := range cast {
			strings[i] = string(bytes)
		}
		return strings, nil
	case []string:
		return cast, nil
	default:
		return nil, ErrReqParamNotType
	}
}
