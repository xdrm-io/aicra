package main

import (
	"context"

	"github.com/xdrm-io/aicra/api"
)

// Endpoints implements endpoints defined in the configuration
// It also wraps the database shared among endpoints
type Endpoints struct {
	db *DB
}

// matches the endpoint definition `in`
type updateUserReq struct {
	ID uint
	// optional parameters are pointer, nil when not provided
	DryRun    *bool
	Username  *string
	Firstname *string
	Lastname  *string
}

// matches the endpoint definition `out`
type updateUserRes struct {
	Username  string
	Firstname string
	Lastname  string
}

// updateUser endpoint definition
func (e *Endpoints) updateUser(ctx context.Context, req updateUserReq) (*updateUserRes, error) {
	// if permissions are not met, there is an automatic response with api.ErrForbidden
	// this function is only called when permissions and the request are valid
	dryRun := (req.DryRun != nil && *req.DryRun)

	// unknown id
	user, err := e.db.FetchUser(uint32(req.ID))
	if err != nil {
		return nil, api.ErrNotFound
	}

	if dryRun {
		if req.Username != nil {
			user.Username = *req.Username
		}
		if req.Firstname != nil {
			user.Firstname = *req.Firstname
		}
		if req.Lastname != nil {
			user.Lastname = *req.Lastname
		}
		return &updateUserRes{
			Username:  user.Username,
			Firstname: user.Firstname,
			Lastname:  user.Lastname,
		}, nil
	}

	if req.Username != nil {
		if err := e.db.UpdateUsername(uint32(req.ID), *req.Username); err != nil {
			return nil, api.ErrUpdate
		}
	}
	if req.Firstname != nil {
		if err := e.db.UpdateFirstname(uint32(req.ID), *req.Firstname); err != nil {
			return nil, api.ErrUpdate
		}
	}
	if req.Lastname != nil {
		if err := e.db.UpdateLastname(uint32(req.ID), *req.Lastname); err != nil {
			return nil, api.ErrUpdate
		}
	}

	// fetch updated user info
	user, err = e.db.FetchUser(uint32(req.ID))
	if err != nil {
		return nil, api.ErrFailure
	}
	return &updateUserRes{
		Username:  user.Username,
		Firstname: user.Firstname,
		Lastname:  user.Lastname,
	}, nil

}
