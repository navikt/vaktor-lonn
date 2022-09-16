package models

import (
	"github.com/google/uuid"
	"time"
)

type Period struct {
	Begin time.Time `json:"start_timestamp"`
	End   time.Time `json:"end_timestamp"`
}

type Vaktplan struct {
	ID       uuid.UUID           `json:"id"`
	Ident    string              `json:"ident"`
	Schedule map[string][]Period `json:"schedule"`
}
