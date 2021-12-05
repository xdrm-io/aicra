package storage

import (
	"sync/atomic"

	"github.com/xdrm-io/aicra/api"
)

// User is the user model
type User struct {
	Username  string `json:"username"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

// CreateUser creates a new user and returns its id
func (db *DB) CreateUser(username, firstname, lastname string) (uint32, error) {
	if db.users == nil {
		db.users = map[uint32]*User{}
	}
	id := atomic.AddUint32(&db.userID, 1)
	db.users[id] = &User{
		Username:  username,
		Firstname: firstname,
		Lastname:  lastname,
	}
	return id, nil
}

// ListUsers returns all existing users
func (db *DB) ListUsers() ([]User, error) {
	users := []User{}
	if db.users == nil {
		return users, nil
	}

	for _, u := range db.users {
		users = append(users, *u)
	}
	return users, nil
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

// DeleteUser deletes an existing user
func (db *DB) DeleteUser(id uint32) error {
	if db.users == nil {
		return api.ErrNotFound
	}
	if _, ok := db.users[id]; !ok {
		return api.ErrNotFound
	}
	delete(db.users, id)
	return nil
}
