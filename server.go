package aicra

import (
	"io"
	"log"
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

	// 4. log configuration services
	log.Printf("🔧   Reading configuration '%s'\n", configPath)
	for _, service := range i.config.Services {
		log.Printf("    ->\t%s\t'%s'\n", service.Method, service.Pattern)
	}

	return i, nil

}

// HandleFunc sets a new handler for an HTTP method to a path
func (s *Server) HandleFunc(httpMethod, path string, handlerFunc api.HandlerFunc) {
	handler := api.NewHandler(httpMethod, path, handlerFunc)
	s.handlers = append(s.handlers, handler)
}

// Handle sets a new handler
func (s *Server) Handle(handler *api.Handler) {
	s.handlers = append(s.handlers, handler)
}

// HTTP converts the server to a http server
func (s Server) HTTP() httpServer {

	// 1. log available handlers
	log.Printf("🔗	 Mapping handlers\n")
	for i := 0; i < len(s.handlers); i++ {
		log.Printf("    ->\t%s\t'%s'\n", s.handlers[i].GetMethod(), s.handlers[i].GetPath())
	}

	// 2. cast to http server
	return httpServer(s)
}
