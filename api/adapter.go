package api

import "net/http"

// Adapter to encapsulate incoming requests
type Adapter func(http.HandlerFunc) http.HandlerFunc

// AuthHandlerFunc is http.HandlerFunc with additional Authorization information
type AuthHandlerFunc func(Auth, http.ResponseWriter, *http.Request)

// AuthAdapter to encapsulate incoming request with access to api.Auth
// to manage permissions
type AuthAdapter func(AuthHandlerFunc) AuthHandlerFunc
