package api

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

// GetString returns a string and an error if not found or string
func (rp RequestParam) GetString(key string) (string, error) {
	rawValue, found := rp[key]
	if !found {
		return "", ErrReqParamNotFound
	}

	convertedValue, canConvert := rawValue.(string)
	if !canConvert {
		return "", ErrReqParamNotType
	}

	return convertedValue, nil
}
