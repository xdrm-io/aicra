package gfw

import (
	"fmt"
	"net/http"
)

// Launch listens and binds the server to the given port
func (s *Server) Launch(port uint16) error {

	/* (1) Bind router */
	http.HandleFunc("/", route)

	/* (2) Bind listener */
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

}
