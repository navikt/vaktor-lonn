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
	Ident    string              `json:"user_id"`
	Schedule map[string][]Period `json:"schedule"`
	Begin    time.Time           `json:"start_timestamp"`
	End      time.Time           `json:"end_timestamp"`
}

// GuardDuty keeps track of minutes not worked in a given guard duty
type GuardDuty struct {
	Hvilende2000                 float64
	Hvilende0006                 float64
	Hvilende0620                 float64
	Helgetillegg                 float64
	Skifttillegg                 float64
	WeekendOrHolidayCompensation bool
}
