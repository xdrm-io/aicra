package aicra

import (
	"fmt"
	"net/http"
	"os"

	"git.xdrm.io/go/aicra/datatype"
	"git.xdrm.io/go/aicra/dynfunc"
	"git.xdrm.io/go/aicra/internal/config"
)

// Server represents an AICRA instance featuring: type checkers, services
type Server struct {
	config   *config.Server
	handlers []*apiHandler
}

type apiHandler struct {
	Method     string
	Path       string
	dynHandler *dynfunc.Handler
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
		handlers: make([]*handler, 0),
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

// Handle sets a new handler for an HTTP method to a path
func (s *Server) Handle(method, path string, fn interface{}) error {
	// find associated service
	var service *config.Service
	for _, s := range s.config.Services {
		if method == s.Method && path == s.Pattern {
			service = s
			break
		}
	}

	if service == nil {
		return fmt.Errorf("%s '%s': %w", method, path, ErrUnknownService)
	}

	dynHandler, err := dynfunc.Build(fn, *service)
	if err != nil {
		return fmt.Errorf("%s '%s' handler: %w", method, path, err)
	}

	s.handlers = append(s.handlers, &apiHandler{
		Path:       path,
		Method:     method,
		dynHandler: dynHandler,
	})

	return nil
}

// ToHTTPServer converts the server to a http.Handler
func (s Server) ToHTTPServer() (http.Handler, error) {
	for _, service := range s.config.Services {
		var hasAssociatedHandler bool
		for _, handler := range s.handlers {
			if handler.Method == service.Method && handler.Path == service.Pattern {
				hasAssociatedHandler = true
				break
			}
		}
		if !hasAssociatedHandler {
			return nil, fmt.Errorf("%s '%s': %w", service.Method, service.Pattern, ErrMissingHandler)
		}
	}

	return httpHandler(s), nil
}
