package generated

import (
	"context"
	"fmt"
	"net/http"

	"github.com/xdrm-io/aicra"
	model "github.com/xdrm-io/aicra/examples/minimal/model"
)

type Server interface {
	// GetUsers fetches all users
	GetUsers(context.Context, GetUsersReq) (*GetUsersRes, error)
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
		{"POST", "/user", impl.CreateUser},
		{"PUT", "/user/{id}", impl.UpdateUser},
		{"DELETE", "/user/{id}", impl.DeleteUser},
	}
}

func Bind(b *aicra.Builder, impl Server) error {
	mapped := mapper{impl: impl}
	for _, r := range routes(mapped) {
		if err := b.Bind(r.method, r.path, r.fn); err != nil {
			return fmt.Errorf("cannot bind '%s %s': %s", r.method, r.path, err)
		}
	}
	return nil
}

type GetUsersReq struct {
}

type GetUsersRes struct {
	Users []model.User
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

type DeleteUserRes struct {
}
