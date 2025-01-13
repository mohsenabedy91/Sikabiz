package domain

import (
	"github.com/google/uuid"
	"time"
)

type Modifier struct {
	CreatedBy *uint64
	UpdatedBy uint64
	DeleteBy  uint64
}

type Base struct {
	ID   uint64
	UUID uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}
