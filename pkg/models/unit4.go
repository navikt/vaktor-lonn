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
	ID         uuid.UUID
	ResourceID string
	Approver   string
	TypeCodes  map[string]decimal.Decimal
	CommitSHA  string
}
