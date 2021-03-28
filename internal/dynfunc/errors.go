package dynfunc

// cerr allows you to create constant "const" error with type boxing.
type cerr string

func (err cerr) Error() string {
	return string(err)
}

// errHandlerNotFunc - handler is not a func
const errHandlerNotFunc = cerr("handler must be a func")

// errNoServiceForHandler - no service matching this handler
const errNoServiceForHandler = cerr("no service found for this handler")

// errMissingHandlerArgumentParam - missing params arguments for handler
const errMissingHandlerArgumentParam = cerr("missing handler argument : parameter struct")

// errUnexpectedInput - input argument is not expected
const errUnexpectedInput = cerr("unexpected input struct")

// errMissingHandlerOutput - missing output for handler
const errMissingHandlerOutput = cerr("handler must have at least 1 output")

// errMissingHandlerOutputError - missing error output for handler
const errMissingHandlerOutputError = cerr("handler must have its last output of type api.Err")

// errMissingRequestArgument - missing request argument for handler
const errMissingRequestArgument = cerr("handler first argument must be of type api.Request")

// errMissingParamArgument - missing parameters argument for handler
const errMissingParamArgument = cerr("handler second argument must be a struct")

// errUnexportedName - argument is unexported in struct
const errUnexportedName = cerr("unexported name")

// errMissingParamOutput - missing output argument for handler
const errMissingParamOutput = cerr("handler first output must be a *struct")

// errMissingParamFromConfig - missing a parameter in handler struct
const errMissingParamFromConfig = cerr("missing a parameter from configuration")

// errMissingOutputFromConfig - missing a parameter in handler struct
const errMissingOutputFromConfig = cerr("missing a parameter from configuration")

// errWrongParamTypeFromConfig - a configuration parameter type is invalid in the handler param struct
const errWrongParamTypeFromConfig = cerr("invalid struct field type")

// errMissingHandlerErrorOutput - missing handler output error
const errMissingHandlerErrorOutput = cerr("last output must be of type api.Err")
