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
	Begin    time.Time
	End      time.Time
}

// GuardDuty keeps track of minutes not worked in a given guard duty
type GuardDuty struct {
	Hvilende2006                 int
	Hvilende0620                 int
	Helgetillegg                 int
	Skifttillegg                 int
	WeekendOrHolidayCompensation bool
}
