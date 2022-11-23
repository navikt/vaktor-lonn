package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Satser struct {
	Dag     decimal.Decimal `json:"0620"`
	Natt    decimal.Decimal `json:"2006"`
	Helg    decimal.Decimal `json:"helg"`
	Utvidet decimal.Decimal `json:"skift"`
}

type Artskoder struct {
	Morgen decimal.Decimal `json:"2680"`
	Kveld  decimal.Decimal `json:"2681"`
	Dag    decimal.Decimal `json:"2682"`
	Helg   decimal.Decimal `json:"2683"`
}

type Payroll struct {
	ID           uuid.UUID
	ApproverID   string    `json:"approver_id"`
	ApproverName string    `json:"approver_name"`
	Artskoder    Artskoder `json:"artskoder"`
	CommitSHA    string    `json:"commit_sha"`
	Formal       string    `json:"formal"`
	Koststed     string    `json:"koststed"`
	Aktivitet    string    `json:"aktivitet"`
}
