package dynfunc

// Err allows you to create constant "const" error with type boxing.
type Err string

func (err Err) Error() string {
	return string(err)
}

const (
	// ErrHandlerNotFunc - handler is not a func
	ErrHandlerNotFunc = Err("handler must be a func")

	// ErrNoServiceForHandler - no service matching this handler
	ErrNoServiceForHandler = Err("no service found for this handler")

	// errMissingHandlerArgumentParam - missing params arguments for handler
	ErrMissingHandlerContextArgument = Err("missing handler first argument of type context.Context")

	// ErrInvalidHandlerContextArgument - missing handler output error
	ErrInvalidHandlerContextArgument = Err("first input argument should be of type context.Context")

	// ErrMissingHandlerInputArgument - missing params arguments for handler
	ErrMissingHandlerInputArgument = Err("missing handler argument: input struct")

	// ErrUnexpectedInput - input argument is not expected
	ErrUnexpectedInput = Err("unexpected input struct")

	// ErrMissingHandlerOutputArgument - missing output for handler
	ErrMissingHandlerOutputArgument = Err("missing handler first output argument: output struct")

	// ErrMissingHandlerErrorArgument - missing error output for handler
	ErrMissingHandlerErrorArgument = Err("missing handler last output argument of type api.Err")

	// ErrInvalidHandlerErrorArgument - missing handler output error
	ErrInvalidHandlerErrorArgument = Err("last output must be of type api.Err")

	// ErrMissingParamArgument - missing parameters argument for handler
	ErrMissingParamArgument = Err("handler second argument must be a struct")

	// ErrUnexportedName - argument is unexported in struct
	ErrUnexportedName = Err("unexported name")

	// ErrWrongOutputArgumentType - wrong type for output first argument
	ErrWrongOutputArgumentType = Err("handler first output argument must be a *struct")

	// ErrMissingConfigArgument - missing an input/output argument in handler struct
	ErrMissingConfigArgument = Err("missing an argument from the configuration")

	// ErrWrongParamTypeFromConfig - a configuration parameter type is invalid in the handler param struct
	ErrWrongParamTypeFromConfig = Err("invalid struct field type")
)
