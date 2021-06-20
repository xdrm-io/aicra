package ctx

// Key defines a custom context key type
type Key int

const (
	// Request is the key for the current *http.Request
	Request Key = iota
	// Response is the key for the associated http.ResponseWriter
	Response
	// Auth is the key for the request's authentication information
	Auth
)
