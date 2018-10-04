package response

import (
	"errors"
)

var ErrUnknownKey = errors.New("key does not exist")
var ErrInvalidType = errors.New("invalid type")

// Has checks whether a key exists in the arguments
func (i Arguments) Has(key string) bool {
	_, exists := i[key]
	return exists
}

// Get extracts a parameter as an interface{} value
func (i Arguments) Get(key string) (interface{}, error) {
	val, ok := i[key]
	if !ok {
		return 0, ErrUnknownKey
	}

	return val, nil
}

// GetFloat extracts a parameter as a float value
func (i Arguments) GetFloat(key string) (float64, error) {
	val, err := i.Get(key)
	if err != nil {
		return 0, err
	}

	floatval, ok := val.(float64)
	if !ok {
		return 0, ErrInvalidType
	}

	return floatval, nil
}

// GetInt extracts a parameter as an int value
func (i Arguments) GetInt(key string) (int, error) {
	floatval, err := i.GetFloat(key)
	if err != nil {
		return 0, err
	}

	return int(floatval), nil
}

// GetUint extracts a parameter as an uint value
func (i Arguments) GetUint(key string) (uint, error) {
	floatval, err := i.GetFloat(key)
	if err != nil {
		return 0, err
	}

	return uint(floatval), nil
}

// GetString extracts a parameter as a string value
func (i Arguments) GetString(key string) (string, error) {
	val, ok := i[key]
	if !ok {
		return "", ErrUnknownKey
	}

	stringval, ok := val.(string)
	if !ok {
		return "", ErrInvalidType
	}

	return stringval, nil
}

// GetBool extracts a parameter as a bool value
func (i Arguments) GetBool(key string) (bool, error) {
	val, ok := i[key]
	if !ok {
		return false, ErrUnknownKey
	}

	boolval, ok := val.(bool)
	if !ok {
		return false, ErrInvalidType
	}

	return boolval, nil
}
