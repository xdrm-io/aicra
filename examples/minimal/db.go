package main

import (
	"fmt"
	"sync/atomic"

	"github.com/xdrm-io/aicra/api"
	"github.com/xdrm-io/aicra/examples/minimal/model"
)

// DB emulates a database for this example but does nothing
type DB struct {
	autoIncrementID uint64
	users           map[string]*model.User
}

// CreateUser creates a new user and returns its id
func (db *DB) CreateUser(username, firstname, lastname string) (string, error) {
	if db.users == nil {
		db.users = map[string]*model.User{}
	}
	id := fmt.Sprintf("%016x", atomic.AddUint64(&db.autoIncrementID, 1))
	db.users[id] = &model.User{
		Username:  username,
		Firstname: firstname,
		Lastname:  lastname,
	}
	return id, nil
}

// FetchAll returns all users
func (db *DB) FetchAll() (map[string]model.User, error) {
	if db.users == nil {
		return nil, api.ErrNotFound
	}
	res := make(map[string]model.User, len(db.users))
	for id, user := range db.users {
		res[id] = *user
	}
	return res, nil
}

// FetchUser returns an existing user
func (db *DB) FetchUser(id string) (model.User, error) {
	if db.users == nil {
		return model.User{}, api.ErrNotFound
	}
	user, ok := db.users[id]
	if !ok {
		return model.User{}, api.ErrNotFound
	}
	return *user, nil
}

// UpdateUsername updates an user's username
func (db *DB) UpdateUsername(id string, username string) error {
	if db.users == nil {
		return api.ErrNotFound
	}
	user, ok := db.users[id]
	if !ok {
		return api.ErrNotFound
	}
	user.Username = username
	return nil
}

// UpdateFirstname updates an existing user's firstname
func (db *DB) UpdateFirstname(id string, firstname string) error {
	if db.users == nil {
		return api.ErrNotFound
	}
	user, ok := db.users[id]
	if !ok {
		return api.ErrNotFound
	}
	user.Firstname = firstname
	return nil
}

// UpdateLastname updates an existing user's lastname
func (db *DB) UpdateLastname(id string, lastname string) error {
	if db.users == nil {
		return api.ErrNotFound
	}
	user, ok := db.users[id]
	if !ok {
		return api.ErrNotFound
	}
	user.Lastname = lastname
	return nil
}

// DeleteUser removes an existing user
func (db *DB) DeleteUser(id string) error {
	if db.users == nil {
		return api.ErrNotFound
	}
	if _, ok := db.users[id]; !ok {
		return api.ErrNotFound
	}
	delete(db.users, id)
	return nil
}
