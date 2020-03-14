package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Parse builds a server configuration from a json reader and checks for most format errors.
func Parse(r io.Reader) (Services, error) {
	services := make(Services, 0)

	err := json.NewDecoder(r).Decode(&services)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrRead, err)
	}

	err = services.checkAndFormat()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrFormat, err)
	}

	if services.collide() {
		return nil, fmt.Errorf("%s: %w", ErrFormat, ErrPatternCollision)
	}

	return services, nil
}

// collide returns if there is collision between services
func (svc Services) collide() bool {
	// todo: implement pattern collision using types to check if braces can be equal to fixed uri parts
	return false
}

// Find a service matching an incoming HTTP request
func (svc Services) Find(r *http.Request) *Service {
	for _, service := range svc {
		if service.Match(r) {
			return service
		}
	}

	return nil
}

// checkAndFormat checks for errors and missing fields and sets default values for optional fields.
func (svc Services) checkAndFormat() error {
	for _, service := range svc {

		// check method
		err := service.checkMethod()
		if err != nil {
			return fmt.Errorf("%s '%s' [method]: %w", service.Method, service.Pattern, err)
		}

		// check pattern
		service.Pattern = strings.Trim(service.Pattern, " \t\r\n")
		err = service.checkPattern()
		if err != nil {
			return fmt.Errorf("%s '%s' [path]: %w", service.Method, service.Pattern, err)
		}

		// check description
		if len(strings.Trim(service.Description, " \t\r\n")) < 1 {
			return fmt.Errorf("%s '%s' [description]: %w", service.Method, service.Pattern, ErrMissingDescription)
		}

		// check input parameters
		err = service.checkAndFormatInput()
		if err != nil {
			return fmt.Errorf("%s '%s' [in]: %w", service.Method, service.Pattern, err)
		}

	}
	return nil
}
