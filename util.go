package aicra

import (
	"log"
	"net/http"

	"git.xdrm.io/go/aicra/api"
)

var handledMethods = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}

// Prints an error as HTTP response
func logError(res *api.Response) {
	log.Printf("[http.fail] %v\n", res)
}
