package endpoints

import (
	"fmt"
	"net/http"

	"github.com/xdrm-io/aicra"
)

type route struct {
	path    string
	method  string
	handler interface{}
}

func (e *Endpoints) wire(b *aicra.Builder) error {
	routes := []route{
		{"/user", http.MethodGet, e.listUsers},
		{"/user/{id}", http.MethodGet, e.getUser},
		{"/user", http.MethodPost, e.createUser},
		{"/user/{id}", http.MethodPut, e.updateUser},
		{"/user/{id}", http.MethodDelete, e.deleteUser},
	}

	for _, r := range routes {
		if err := b.Bind(r.method, r.path, r.handler); err != nil {
			return fmt.Errorf("'%s %s': %w", r.method, r.path, err)
		}
	}
	return nil
}
