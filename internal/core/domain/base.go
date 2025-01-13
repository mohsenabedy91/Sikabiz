package domain

import "time"

type Modifier struct {
	CreatedBy *uint64
	UpdatedBy uint64
	DeleteBy  uint64
}

type Base struct {
	ID uint64

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}
