package main

import (
	"sync/atomic"

	"github.com/xdrm-io/aicra/api"
)

// User is the model of an user
type User struct {
	Username  string
	Firstname string
	Lastname  string
}

// DB emulates a database for this example but does nothing
type DB struct {
	autoIncrementID uint32
	users           map[uint32]*User
}

// CreateUser creates a new user and returns its id
func (db *DB) CreateUser(username, firstname, lastname string) (uint32, error) {
	if db.users == nil {
		db.users = map[uint32]*User{}
	}
	id := atomic.AddUint32(&db.autoIncrementID, 1)
	db.users[id] = &User{
		Username:  username,
		Firstname: firstname,
		Lastname:  lastname,
	}
	return id, nil
}

// FetchUser returns an existing user
func (db *DB) FetchUser(id uint32) (User, error) {
	if db.users == nil {
		return User{}, api.ErrNotFound
	}
	user, ok := db.users[id]
	if !ok {
		return User{}, api.ErrNotFound
	}
	return *user, nil
}

// UpdateUsername updates an user's username
func (db *DB) UpdateUsername(id uint32, username string) error {
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
func (db *DB) UpdateFirstname(id uint32, firstname string) error {
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
func (db *DB) UpdateLastname(id uint32, lastname string) error {
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
