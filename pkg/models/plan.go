package models

import (
	"time"

	"github.com/google/uuid"
)

type Period struct {
	Begin time.Time `json:"start_timestamp"`
	End   time.Time `json:"end_timestamp"`
}

type Vaktplan struct {
	ID       uuid.UUID           `json:"id"`
	Ident    string              `json:"user_id"`
	Schedule map[string][]Period `json:"schedule"`
}

// GuardDuty keeps track of minutes not worked in a given guard duty
type GuardDuty struct {
	Hvilende2000  float64
	Hvilende0006  float64
	Hvilende0620  float64
	Helligdag0620 float64
	Helgetillegg  float64
	Skifttillegg  float64
	IsWeekend     bool
}
