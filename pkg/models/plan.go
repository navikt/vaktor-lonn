package models

import "time"

type Period struct {
	Begin time.Time `json:"start_timestamp"`
	End   time.Time `json:"end_timestamp"`
}

type Vaktplan struct {
	Ident    string              `json:"id"`
	Schedule map[string][]Period `json:"schedule"`
}
