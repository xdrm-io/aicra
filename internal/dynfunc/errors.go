package dynfunc

// Err allows you to create constant "const" error with type boxing.
type Err string

func (err Err) Error() string {
	return string(err)
}

const (
	// ErrNotAStruct - request/response is not a struct
	ErrNotAStruct = Err("must be a struct")

	// ErrUnexpectedFields - request/response fields are unexpected
	ErrUnexpectedFields = Err("unexpected struct fields")

	// ErrUnexportedField - struct field is unexported
	ErrUnexportedField = Err("unexported field name")

	// ErrMissingField - missing request/response field
	ErrMissingField = Err("missing struct field from the configuration")

	// ErrInvalidType - invalid struct field type
	ErrInvalidType = Err("invalid struct field type")
)
