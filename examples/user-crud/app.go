package main

import (
	"io"
	"log"
	"net/http"

	"github.com/xdrm-io/aicra"
	"github.com/xdrm-io/aicra/examples/user-crud/api/endpoints"
	"github.com/xdrm-io/aicra/examples/user-crud/api/middlewares"
	"github.com/xdrm-io/aicra/examples/user-crud/api/validators"
	"github.com/xdrm-io/aicra/examples/user-crud/storage"
	"github.com/xdrm-io/aicra/validator"
)

// App wraps the application
type App struct {
	endpoints   *endpoints.Endpoints
	httpHandler http.Handler
	db          *storage.DB
}

// NewApp builds a new application
func NewApp(config io.Reader, db *storage.DB) (*App, error) {
	builder := &aicra.Builder{}

	// list available type validators
	builder.Validate(validator.BoolType{})
	builder.Validate(validator.UintType{})
	builder.Validate(validator.StringType{})
	builder.Validate(validators.User{})
	builder.Validate(validators.Users{})

	// load the api definition
	if err := builder.Setup(config); err != nil {
		log.Fatalf("invalid config: %s", err)
	}

	// middlewares
	builder.With(middlewares.Logger)

	// endpoints
	endpoints, err := endpoints.New(builder, db)
	if err != nil {
		log.Fatalf("cannot build endpoints: %s", err)
	}

	// build your services
	handler, err := builder.Build()
	if err != nil {
		log.Fatalf("cannot build handler: %s", err)
	}

	return &App{
		httpHandler: handler,
		db:          db,
		endpoints:   endpoints,
	}, nil
}

// Listen and serve
func (a *App) Listen(addr string) error {
	return http.ListenAndServe(addr, a.httpHandler)
}
