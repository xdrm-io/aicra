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

// ErrMissingHandlerArgumentParam - missing params arguments for handler
const ErrMissingHandlerArgumentParam = cerr("missing handler argument : parameter struct")

// ErrUnexpectedInput - input argument is not expected
const ErrUnexpectedInput = cerr("unexpected input struct")

// ErrMissingHandlerOutput - missing output for handler
const ErrMissingHandlerOutput = cerr("handler must have at least 1 output")

// ErrMissingHandlerOutputError - missing error output for handler
const ErrMissingHandlerOutputError = cerr("handler must have its last output of type api.Error")

// ErrMissingRequestArgument - missing request argument for handler
const ErrMissingRequestArgument = cerr("handler first argument must be of type api.Request")

// ErrMissingParamArgument - missing parameters argument for handler
const ErrMissingParamArgument = cerr("handler second argument must be a struct")

// ErrUnexportedName - argument is unexported in struct
const ErrUnexportedName = cerr("unexported name")

// ErrMissingParamOutput - missing output argument for handler
const ErrMissingParamOutput = cerr("handler first output must be a *struct")

// ErrMissingParamFromConfig - missing a parameter in handler struct
const ErrMissingParamFromConfig = cerr("missing a parameter from configuration")

// ErrMissingOutputFromConfig - missing a parameter in handler struct
const ErrMissingOutputFromConfig = cerr("missing a parameter from configuration")

// ErrWrongParamTypeFromConfig - a configuration parameter type is invalid in the handler param struct
const ErrWrongParamTypeFromConfig = cerr("invalid struct field type")

// ErrMissingHandlerErrorOutput - missing handler output error
const ErrMissingHandlerErrorOutput = cerr("last output must be of type api.Error")
