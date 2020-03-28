package aicra

import (
	"fmt"
	"io"
	"os"

	"git.xdrm.io/go/aicra/api"
	"git.xdrm.io/go/aicra/datatype"
	"git.xdrm.io/go/aicra/internal/config"
)

// Server represents an AICRA instance featuring: type checkers, services
type Server struct {
	config   *config.Server
	handlers []*api.Handler
}

// New creates a framework instance from a configuration file
func New(configPath string, dtypes ...datatype.T) (*Server, error) {
	var (
		err        error
		configFile io.ReadCloser
	)

	// 1. init instance
	var i = &Server{
		config:   nil,
		handlers: make([]*api.Handler, 0),
	}

	// 2. open config file
	configFile, err = os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	// 3. load configuration
	i.config, err = config.Parse(configFile, dtypes...)
	if err != nil {
		return nil, err
	}

	return i, nil

}

// HandleFunc sets a new handler for an HTTP method to a path
func (s *Server) Handle(httpMethod, path string, fn api.HandlerFn) {
	handler, err := api.NewHandler(httpMethod, path, fn)
	if err != nil {
		panic(err)
	}
	s.handlers = append(s.handlers, handler)
}

// ToHTTPServer converts the server to a http server
func (s Server) ToHTTPServer() (*httpServer, error) {

	// check if handlers are missing
	for _, service := range s.config.Services {
		found := false
		for _, handler := range s.handlers {
			if handler.GetMethod() == service.Method && handler.GetPath() == service.Pattern {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("missing handler for %s '%s'", service.Method, service.Pattern)
		}
	}

	// 2. cast to http server
	httpServer := httpServer(s)
	return &httpServer, nil
}
