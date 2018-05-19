package gfw

import (
	"fmt"
)

type Err struct {
	Code      int
	Reason    string
	Arguments []interface{}
}

var (
	/* Base */
	Success = &Err{0, "all right", nil}
	Failure = &Err{1, "it failed", nil}
	Unknown = &Err{-1, "", nil}

	NoMatchFound  = &Err{2, "no resource found", nil}
	AlreadyExists = &Err{3, "resource already exists", nil}

	Config = &Err{4, "configuration error", nil}

	/* I/O */
	UploadError            = &Err{100, "upload failed", nil}
	DownloadError          = &Err{101, "download failed", nil}
	MissingDownloadHeaders = &Err{102, "download headers are missing", nil}
	MissingDownloadBody    = &Err{103, "download body is missing", nil}

	/* Controllers */
	UnknownController    = &Err{200, "unknown controller", nil}
	UnknownMethod        = &Err{201, "unknown method", nil}
	UncallableController = &Err{202, "uncallable controller", nil}
	UncallableMethod     = &Err{203, "uncallable method", nil}

	/* Permissions */
	Permission = &Err{300, "permission error", nil}
	Token      = &Err{301, "token error", nil}

	/* Check */
	MissingParam        = &Err{400, "missing parameter", nil}
	InvalidParam        = &Err{401, "invalid parameter", nil}
	InvalidDefaultParam = &Err{402, "invalid default param", nil}
)

func (e Err) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Reason)
}
