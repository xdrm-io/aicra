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
const errMissingHandlerContextArgument = cerr("missing handler first argument of type context.Context")

// errMissingHandlerInputArgument - missing params arguments for handler
const errMissingHandlerInputArgument = cerr("missing handler argument: input struct")

// errUnexpectedInput - input argument is not expected
const errUnexpectedInput = cerr("unexpected input struct")

// errMissingHandlerOutputArgument - missing output for handler
const errMissingHandlerOutputArgument = cerr("missing handler first output argument: output struct")

// errMissingHandlerOutputError - missing error output for handler
const errMissingHandlerOutputError = cerr("missing handler last output argument of type api.Err")

// errMissingRequestArgument - missing request argument for handler
const errMissingRequestArgument = cerr("handler first argument must be of type api.Request")

// errMissingParamArgument - missing parameters argument for handler
const errMissingParamArgument = cerr("handler second argument must be a struct")

// errUnexportedName - argument is unexported in struct
const errUnexportedName = cerr("unexported name")

// errWrongOutputArgumentType - wrong type for output first argument
const errWrongOutputArgumentType = cerr("handler first output argument must be a *struct")

// errMissingConfigArgument - missing an input/output argument in handler struct
const errMissingConfigArgument = cerr("missing an argument from the configuration")

// errWrongParamTypeFromConfig - a configuration parameter type is invalid in the handler param struct
const errWrongParamTypeFromConfig = cerr("invalid struct field type")

// errMissingHandlerErrorArgument - missing handler output error
const errMissingHandlerErrorArgument = cerr("last output must be of type api.Err")
