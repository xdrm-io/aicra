package err

var (
	/* Base */
	Success = Error{0, "all right", nil}
	Failure = Error{1, "it failed", nil}
	Unknown = Error{-1, "", nil}

	NoMatchFound  = Error{2, "no resource found", nil}
	AlreadyExists = Error{3, "resource already exists", nil}

	Config = Error{4, "configuration error", nil}

	/* I/O */
	Upload                 = Error{100, "upload failed", nil}
	Download               = Error{101, "download failed", nil}
	MissingDownloadHeaders = Error{102, "download headers are missing", nil}
	MissingDownloadBody    = Error{103, "download body is missing", nil}

	/* Controllers */
	UnknownController    = Error{200, "unknown controller", nil}
	UnknownMethod        = Error{201, "unknown method", nil}
	UncallableController = Error{202, "uncallable controller", nil}
	UncallableMethod     = Error{203, "uncallable method", nil}

	/* Permissions */
	Permission = Error{300, "permission error", nil}
	Token      = Error{301, "token error", nil}

	/* Check */
	MissingParam        = Error{400, "missing parameter", nil}
	InvalidParam        = Error{401, "invalid parameter", nil}
	InvalidDefaultParam = Error{402, "invalid default param", nil}
)
