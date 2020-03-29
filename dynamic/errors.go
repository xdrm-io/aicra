package dynamic

// cerr allows you to create constant "const" error with type boxing.
type cerr string

// Error implements the error builtin interface.
func (err cerr) Error() string {
	return string(err)
}

// ErrHandlerNotFunc - handler is not a func
const ErrHandlerNotFunc = cerr("handler must be a func")

// ErrNoServiceForHandler - no service matching this handler
const ErrNoServiceForHandler = cerr("no service found for this handler")

// ErrMissingHandlerArgument - missing arguments for handler
const ErrMissingHandlerArgument = cerr("handler must have at least 1 arguments")

// ErrMissingHandlerArgumentParam - missing params arguments for handler
const ErrMissingHandlerArgumentParam = cerr("missing handler argument 2 : parameter struct")

// ErrMissingHandlerOutput - missing output for handler
const ErrMissingHandlerOutput = cerr("handler must have at least 1 output")

// ErrMissingHandlerOutputError - missing error output for handler
const ErrMissingHandlerOutputError = cerr("handler must have its last output of type api.Error")

// ErrMissingRequestArgument - missing request argument for handler
const ErrMissingRequestArgument = cerr("handler first argument must be of type api.Request")

// ErrMissingParamArgument - missing parameters argument for handler
const ErrMissingParamArgument = cerr("handler second argument must be a struct")

// ErrMissingParamOutput - missing output argument for handler
const ErrMissingParamOutput = cerr("handler first output must be a struct")

// ErrMissingParamFromConfig - missing a parameter in handler struct
const ErrMissingParamFromConfig = cerr("missing a parameter from configuration")

// ErrMissingOutputFromConfig - missing a parameter in handler struct
const ErrMissingOutputFromConfig = cerr("missing a parameter from configuration")

// ErrWrongParamTypeFromConfig - a configuration parameter type is invalid in the handler param struct
const ErrWrongParamTypeFromConfig = cerr("invalid parameter struct type from configuration")

// ErrWrongOutputTypeFromConfig - a configuration output type is invalid in the handler output struct
const ErrWrongOutputTypeFromConfig = cerr("invalid output struct type from configuration")

// ErrMissingHandlerErrorOutput - missing handler output error
const ErrMissingHandlerErrorOutput = cerr("last output must be of type api.Error")
