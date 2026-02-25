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

type Artskode struct {
	Sum   decimal.Decimal `json:"sum"`
	Hours int64           `json:"hours"`
}

type Artskoder struct {
	Morgen    Artskode `json:"2680"`
	Kveld     Artskode `json:"2681"`
	Dag       Artskode `json:"2682"`
	Helg      Artskode `json:"2683"`
	Skift     Artskode `json:"2684"`
	Utrykning Artskode `json:"2685"`
}

type Payroll struct {
	ID            uuid.UUID
	ApproverID    string    `json:"approver_id"`
	ApproverName  string    `json:"approver_name"`
	Artskoder     Artskoder `json:"artskoder"`
	CommitSHA     string    `json:"commit_sha"`
	Stillingskode string    `json:"stillingskode"`
}
