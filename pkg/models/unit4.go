package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	ArtskodeMorgen = "2680"
	ArtskodeKveld  = "2681"
	ArtskodeDag    = "2682"
	ArtskodeHelg   = "2683"
)

type Payroll struct {
	ID           uuid.UUID
	ApproverID   string                     `json:"approver_id"`
	ApproverName string                     `json:"approver_name"`
	TypeCodes    map[string]decimal.Decimal `json:"artskoder"`
	CommitSHA    string                     `json:"commit_sha"`
	Formal       string                     `json:"formal"`
	Koststed     string                     `json:"koststed"`
	Aktivitet    string                     `json:"aktivitet"`
}
