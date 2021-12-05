package endpoints

import (
	"context"

	"github.com/xdrm-io/aicra/api"
	"github.com/xdrm-io/aicra/examples/user-crud/storage"
)

type createRequest struct {
	Username  string
	Firstname string
	Lastname  string
}
type createResponse struct {
	ID   uint
	User storage.User
}

type updateRequest struct {
	ID        uint
	Username  *string
	Firstname *string
	Lastname  *string
}
type idRequest struct {
	ID uint
}

type userResponse struct {
	User storage.User
}
type usersResponse struct {
	Users []storage.User
}

// listUsers is the endpoint to list available users
func (e *Endpoints) listUsers(ctx context.Context) (*usersResponse, error) {
	users, err := e.db.ListUsers()
	if err != nil {
		return nil, api.ErrNotFound
	}
	return &usersResponse{Users: users}, nil
}

// getUser is the endpoint to get a specific user's information
func (e *Endpoints) getUser(ctx context.Context, req idRequest) (*userResponse, error) {
	user, err := e.db.FetchUser(uint32(req.ID))
	if err != nil {
		return nil, api.ErrNotFound
	}
	return &userResponse{User: user}, nil
}

// createUser is the endpoint to create a new user
func (e *Endpoints) createUser(ctx context.Context, req createRequest) (*createResponse, error) {
	id, err := e.db.CreateUser(req.Username, req.Firstname, req.Lastname)
	if err != nil {
		return nil, api.ErrCreate
	}
	user, err := e.db.FetchUser(id)
	if err != nil {
		return nil, api.ErrFailure
	}
	return &createResponse{ID: uint(id), User: user}, nil
}

// updateUser is the endpoint to update an existing user
func (e *Endpoints) updateUser(ctx context.Context, req updateRequest) (*userResponse, error) {
	// only update fields that are provided
	if req.Username != nil {
		err := e.db.UpdateUsername(uint32(req.ID), *req.Username)
		if err != nil {
			return nil, api.ErrUpdate
		}
	}
	if req.Firstname != nil {
		err := e.db.UpdateFirstname(uint32(req.ID), *req.Firstname)
		if err != nil {
			return nil, api.ErrUpdate
		}
	}
	if req.Lastname != nil {
		err := e.db.UpdateLastname(uint32(req.ID), *req.Lastname)
		if err != nil {
			return nil, api.ErrUpdate
		}
	}

	user, err := e.db.FetchUser(uint32(req.ID))
	if err != nil {
		return nil, api.ErrFailure
	}
	return &userResponse{User: user}, nil
}

// deleteUser is the endpoint to delete an existing user
func (e *Endpoints) deleteUser(ctx context.Context, req idRequest) error {
	err := e.db.DeleteUser(uint32(req.ID))
	if err != nil {
		return api.ErrDelete
	}
	return nil
}
