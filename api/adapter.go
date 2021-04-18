package api

import "net/http"

// Adapter to encapsulate incoming requests
type Adapter func(http.HandlerFunc) http.HandlerFunc
