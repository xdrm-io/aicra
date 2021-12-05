package storage

// DB defines the in-memory dummy database
type DB struct {
	userID uint32
	users  map[uint32]*User
}
