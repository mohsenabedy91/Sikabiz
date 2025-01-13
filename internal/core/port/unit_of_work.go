package port

import (
	"context"
)

type UnitOfWork interface {
	BeginTx(ctx context.Context) error
	Commit() error
	Rollback() error
}

type UserUnitOfWork interface {
	UnitOfWork

	UserRepository() UserRepository
	// Add other repositories as needed
}
