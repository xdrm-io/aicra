package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"git.xdrm.io/go/aicra/internal/cerr"
)

// ErrRead - a problem ocurred when trying to read the configuration file
const ErrRead = cerr.Error("cannot read config")

// ErrFormat - a invalid format has been detected
const ErrFormat = cerr.Error("invalid config format")

// Parse builds a service from a json reader and checks for most format errors.
func Parse(r io.Reader) (*Service, error) {
	receiver := &Service{}

	err := json.NewDecoder(r).Decode(receiver)
	if err != nil {
		return nil, ErrRead.Wrap(err)
	}

	err = receiver.checkAndFormat("/")
	if err != nil {
		return nil, ErrFormat.Wrap(err)
	}

	return receiver, nil
}

// Method returns the actual method from the http method.
func (svc *Service) Method(httpMethod string) *Method {
	httpMethod = strings.ToUpper(httpMethod)

	switch httpMethod {
	case http.MethodGet:
		return svc.GET
	case http.MethodPost:
		return svc.POST
	case http.MethodPut:
		return svc.PUT
	case http.MethodDelete:
		return svc.DELETE
	}

	return nil
}

// Browse the service childtree and returns the deepest matching child. The `path` is a formatted URL split by '/'
func (svc *Service) Browse(path []string) (*Service, int) {
	currentService := svc
	var depth int

	// for each URI depth
	for depth = 0; depth < len(path); depth++ {
		currentPath := path[depth]

		child, exists := currentService.Children[currentPath]
		if !exists {
			break
		}
		currentService = child

	}

	return currentService, depth
}

// checkAndFormat checks for errors and missing fields and sets default values for optional fields.
func (svc *Service) checkAndFormat(servicePath string) error {

	// 1. check and format every method
	for _, httpMethod := range availableHTTPMethods {
		methodDef := svc.Method(httpMethod)
		if methodDef == nil {
			continue
		}

		err := methodDef.checkAndFormat(servicePath, httpMethod)
		if err != nil {
			return err
		}
	}

	// 1. stop if no child */
	if svc.Children == nil || len(svc.Children) < 1 {
		return nil
	}

	// 2. for each service */
	for childService, ctl := range svc.Children {

		// 3. invalid name */
		if strings.ContainsAny(childService, "/-") {
			return fmt.Errorf("service '%s' must not contain any slash '/' nor '-' symbols", childService)
		}

		// 4. check recursively */
		err := ctl.checkAndFormat(childService)
		if err != nil {
			return err
		}

	}

	return nil

}
