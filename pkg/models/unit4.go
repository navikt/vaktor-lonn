package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	ArtskodeMorgen = "2600B"
	ArtskodeKveld  = "2603B"
	ArtskodeDag    = "2604B"
	ArtskodeHelg   = "2606B"
)

type Payroll struct {
	ID         uuid.UUID
	ResourceID string
	Approver   string
	TypeCodes  map[string]decimal.Decimal
	CommitSHA  string
}
