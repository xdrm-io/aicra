package generated

import (
	"bytes"
	"fmt"
	"context"
	"net/http"
	"github.com/xdrm-io/aicra"
	model "github.com/xdrm-io/aicra/examples/minimal/model"

	_ "embed"
)

type Server interface {
	// GetUsers fetches all users
	GetUsers(context.Context, GetUsersReq) (*GetUsersRes, error)
	// GetUser fetches all users
	GetUser(context.Context, GetUserReq) (*GetUserRes, error)
	// CreateUser creates a new user
	CreateUser(context.Context, CreateUserReq) (*CreateUserRes, error)
	// UpdateUser updates user information
	UpdateUser(context.Context, UpdateUserReq) (*UpdateUserRes, error)
	// DeleteUser deletes an existing user
	DeleteUser(context.Context, DeleteUserReq) (*DeleteUserRes, error)
}

type route struct {
	method, path string
	fn           http.HandlerFunc
}

func routes(impl mapper) []route {
	return []route{
		{"GET", "/user", impl.GetUsers},
		{"GET", "/user/{id}", impl.GetUser},
		{"POST", "/user", impl.CreateUser},
		{"PUT", "/user/{id}", impl.UpdateUser},
		{"DELETE", "/user/{id}", impl.DeleteUser},
	}
}

//go:embed api.json
var config []byte

func New(impl Server) (*aicra.Builder, error) {
	b := &aicra.Builder{}
	if err := b.Setup(bytes.NewReader(config)); err != nil {
		return nil, fmt.Errorf("cannot setup: %w", err)
	}

	mapped := mapper{impl: impl}
	for _, r := range routes(mapped) {
		if err := b.Bind(r.method, r.path, r.fn); err != nil {
			return nil, fmt.Errorf("cannot bind '%s %s': %w", r.method, r.path, err)
		}
	}

	return b, nil
}

type GetUsersReq struct{}
type GetUsersRes struct {
	Users []model.User
}
type GetUserReq struct {
	ID string
}
type GetUserRes struct {
	Firstname string
	Lastname  string
	Username  string
}
type CreateUserReq struct {
	Firstname string
	Lastname  string
	Username  string
}
type CreateUserRes struct {
	Firstname string
	ID        string
	Lastname  string
	Username  string
}
type UpdateUserReq struct {
	Firstname *string
	Lastname  *string
	Username  *string
	ID        string
}
type UpdateUserRes struct {
	Firstname string
	ID        string
	Lastname  string
	Username  string
}
type DeleteUserReq struct {
	ID string
}
type DeleteUserRes struct{}
