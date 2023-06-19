package dynfunc

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/xdrm-io/aicra/internal/config"
)

// Signature represents input and output arguments for a specific service of the
// aicra configuration
type Signature struct {
	In  map[string]reflect.Type
	Out map[string]reflect.Type
}

// NewSignature builds the handler signature type from a service's configuration
func NewSignature(service *config.Endpoint) *Signature {
	s := &Signature{
		In:  make(map[string]reflect.Type, len(service.Input)),
		Out: make(map[string]reflect.Type, len(service.Output)),
	}

	for _, param := range service.Input {
		if len(param.Rename) < 1 {
			continue
		}
		// make a pointer if optional
		if param.Optional {
			s.In[param.Rename] = reflect.PtrTo(param.GoType)
			continue
		}
		s.In[param.Rename] = param.GoType
	}

	for _, param := range service.Output {
		if len(param.Rename) < 1 {
			continue
		}
		s.Out[param.Rename] = param.GoType
	}
	return s
}

// ValidateRequest type of a handler against the service signature
func (s *Signature) ValidateRequest(treq reflect.Type) error {
	if treq.Kind() != reflect.Struct {
		return ErrNotAStruct
	}

	// no input required
	if len(s.In) == 0 {
		// fail on unexpected fields
		if treq.NumField() > 0 {
			return ErrUnexpectedFields
		}
		return nil
	}

	// check for invalid param
	for name, tparam := range s.In {
		if name[0] == strings.ToLower(name)[0] {
			return fmt.Errorf("%s: %w", name, ErrUnexportedField)
		}

		field, exists := treq.FieldByName(name)
		if !exists {
			return fmt.Errorf("%s: %w", name, ErrMissingField)
		}

		if !tparam.AssignableTo(field.Type) {
			return fmt.Errorf("%s: %w (%s instead of %s)", name, ErrInvalidType, field.Type, tparam)
		}
	}
	return nil
}

// ValidateResponse type of a handler against the service signature
func (s Signature) ValidateResponse(tres reflect.Type) error {
	if tres.Kind() != reflect.Struct {
		return ErrNotAStruct
	}

	// no output required
	if len(s.Out) == 0 {
		// fail when output struct is still provided
		if tres.NumField() > 0 {
			return ErrUnexpectedFields
		}
		return nil
	}

	// fail on invalid param
	for name, tparam := range s.Out {
		if name[0] == strings.ToLower(name)[0] {
			return fmt.Errorf("%s: %w", name, ErrUnexportedField)
		}

		field, exists := tres.FieldByName(name)
		if !exists {
			return fmt.Errorf("%s: %w", name, ErrMissingField)
		}

		if !field.Type.ConvertibleTo(tparam) {
			return fmt.Errorf("%s: %w (%s instead of %s)", name, ErrInvalidType, field.Type, tparam)
		}
	}
	return nil
}
