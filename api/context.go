package api

import (
	"net/http"
)

// Ctx contains additional information for handlers
//
// usually input/output arguments built by aicra are sufficient
// but the Ctx lets you manage your request from scratch if required
//
// If required, set api.Ctx as the first argument of your handler; if you
// don't need it, only use standard input arguments and it will be ignored
type Ctx struct {
	w http.ResponseWriter
	r *http.Request
}
