package gfw

import (
	"encoding/json"
	"fmt"
)

type Err struct {
	Code      int
	Reason    string
	Arguments []interface{}
}

var (
	/* Base */
	ErrSuccess = &Err{0, "all right", nil}
	ErrFailure = &Err{1, "it failed", nil}
	ErrUnknown = &Err{-1, "", nil}

	ErrNoMatchFound  = &Err{2, "no resource found", nil}
	ErrAlreadyExists = &Err{3, "resource already exists", nil}

	ErrConfig = &Err{4, "configuration error", nil}

	/* I/O */
	ErrUpload                 = &Err{100, "upload failed", nil}
	ErrDownload               = &Err{101, "download failed", nil}
	ErrMissingDownloadHeaders = &Err{102, "download headers are missing", nil}
	ErrMissingDownloadBody    = &Err{103, "download body is missing", nil}

	/* Controllers */
	ErrUnknownController    = &Err{200, "unknown controller", nil}
	ErrUnknownMethod        = &Err{201, "unknown method", nil}
	ErrUncallableController = &Err{202, "uncallable controller", nil}
	ErrUncallableMethod     = &Err{203, "uncallable method", nil}

	/* Permissions */
	ErrPermission = &Err{300, "permission error", nil}
	ErrToken      = &Err{301, "token error", nil}

	/* Check */
	ErrMissingParam        = &Err{400, "missing parameter", nil}
	ErrInvalidParam        = &Err{401, "invalid parameter", nil}
	ErrInvalidDefaultParam = &Err{402, "invalid default param", nil}
)

// BindArgument adds an argument to the error
// to be displayed back to API caller
func (e *Err) BindArgument(arg interface{}) {

	/* (1) Make slice if not */
	if e.Arguments == nil {
		e.Arguments = make([]interface{}, 0)
	}

	/* (2) Append argument */
	e.Arguments = append(e.Arguments, arg)

}

// Implements 'error'
func (e Err) Error() string {

	return fmt.Sprintf("[%d] %s", e.Code, e.Reason)

}

// Implements json.Marshaler
func (e Err) MarshalJSON() ([]byte, error) {

	var json_arguments string

	/* (1) Marshal 'Arguments' if set */
	if e.Arguments != nil && len(e.Arguments) > 0 {
		arg_representation, err := json.Marshal(e.Arguments)
		if err == nil {
			json_arguments = fmt.Sprintf(",\"arguments\":%s", arg_representation)
		}

	}

	/* (2) Render JSON manually */
	return []byte(fmt.Sprintf("{\"error\":%d,\"reason\":\"%s\"%s}", e.Code, e.Reason, json_arguments)), nil

}
