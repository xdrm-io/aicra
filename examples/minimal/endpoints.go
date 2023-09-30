package main

import (
	"context"

	"github.com/xdrm-io/aicra/examples/minimal/generated"
	"github.com/xdrm-io/aicra/examples/minimal/model"
)

// Endpoints implements generated.Server
type Endpoints struct {
	db *DB
}

// NewEndpoints creates a new Endpoints
func NewEndpoints(db *DB) *Endpoints {
	return &Endpoints{db: db}
}

// GetUsers implements generated.Server
func (e *Endpoints) GetUsers(ctx context.Context, req generated.GetUsersReq) (*generated.GetUsersRes, error) {
	users, err := e.db.FetchAll()
	if err != nil {
		return nil, err
	}
	res := make(model.Users, 0, len(users))
	for id, user := range users {
		user.ID = id
		res = append(res, user)
	}

	return &generated.GetUsersRes{
		Users: res,
	}, nil
}

// GetUser implements generated.Server
func (e *Endpoints) GetUser(ctx context.Context, req generated.GetUserReq) (*generated.GetUserRes, error) {
	users, err := e.db.FetchUser(req.ID)
	if err != nil {
		return nil, err
	}
	return &generated.GetUserRes{
		Username:  users.Username,
		Firstname: users.Firstname,
		Lastname:  users.Lastname,
	}, nil
}

// CreateUser implements generated.Server
func (e *Endpoints) CreateUser(ctx context.Context, req generated.CreateUserReq) (*generated.CreateUserRes, error) {
	id, err := e.db.CreateUser(req.Username, req.Firstname, req.Lastname)
	if err != nil {
		return nil, err
	}
	return &generated.CreateUserRes{
		ID:        id,
		Username:  req.Username,
		Firstname: req.Firstname,
		Lastname:  req.Lastname,
	}, nil
}

// UpdateUser implements generated.Server
func (e *Endpoints) UpdateUser(ctx context.Context, req generated.UpdateUserReq) (*generated.UpdateUserRes, error) {
	_, err := e.db.FetchUser(req.ID)
	if err != nil {
		return nil, err
	}

	if req.Username != nil {
		if err := e.db.UpdateUsername(req.ID, *req.Username); err != nil {
			return nil, err
		}
	}
	if req.Firstname != nil {
		if err := e.db.UpdateFirstname(req.ID, *req.Firstname); err != nil {
			return nil, err
		}
	}

	if req.Lastname != nil {
		if err := e.db.UpdateLastname(req.ID, *req.Lastname); err != nil {
			return nil, err
		}
	}

	user, err := e.db.FetchUser(req.ID)
	if err != nil {
		return nil, err
	}
	user.ID = req.ID

	return &generated.UpdateUserRes{
		ID:        req.ID,
		Username:  user.Username,
		Firstname: user.Firstname,
		Lastname:  user.Lastname,
	}, nil
}

// DeleteUser implements generated.Server
func (e *Endpoints) DeleteUser(ctx context.Context, req generated.DeleteUserReq) (*generated.DeleteUserRes, error) {
	if err := e.db.DeleteUser(req.ID); err != nil {
		return nil, err
	}
	return &generated.DeleteUserRes{}, nil
}
