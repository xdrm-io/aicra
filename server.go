package aicra

import (
	"io"
	"log"
	"os"

	"git.xdrm.io/go/aicra/api"

	"git.xdrm.io/go/aicra/internal/config"
	checker "git.xdrm.io/go/aicra/typecheck"
)

// Server represents an AICRA instance featuring: type checkers, services
type Server struct {
	services *config.Service
	Checkers *checker.Set
	handlers []*api.Handler
}

// New creates a framework instance from a configuration file
func New(configPath string) (*Server, error) {
	var (
		err        error
		configFile io.ReadCloser
	)

	// 1. init instance
	var i = &Server{
		services: nil,
		Checkers: checker.New(),
		handlers: make([]*api.Handler, 0),
	}

	// 2. open config file
	configFile, err = os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	// 3. load configuration
	i.services, err = config.Parse(configFile)
	if err != nil {
		return nil, err
	}

	// 4. log configuration services
	log.Printf("=== Aicra configuration ===\n")
	logService(*i.services, "")

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
	log.Printf("=== Mapped handlers ===\n")
	for i := 0; i < len(s.handlers); i++ {
		log.Printf("* [rest] %s\t'%s'\n", s.handlers[i].GetMethod(), s.handlers[i].GetPath())
	}

	// 2. cast to http server
	return httpServer(s)
}
