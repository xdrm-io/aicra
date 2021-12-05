package endpoints

import (
	"fmt"

	"github.com/xdrm-io/aicra"
	"github.com/xdrm-io/aicra/examples/user-crud/storage"
)

// Endpoints wraps all endpoints with dependency injection
type Endpoints struct {
	db *storage.DB
}

// New builds endpoints into the aicra server
func New(b *aicra.Builder, db *storage.DB) (*Endpoints, error) {
	e := &Endpoints{
		db: db,
	}

	if err := e.wire(b); err != nil {
		return nil, fmt.Errorf("wire: %w", err)
	}

	return e, nil
}
