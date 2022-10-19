package models

import (
	"github.com/shopspring/decimal"
	"time"
)

const (
	KursSeminar        = 740
	AnnetFravarMedLonn = 940
)

type Clocking struct {
	In  time.Time
	Out time.Time
}

type TimeSheet struct {
	Date         time.Time
	WorkingHours float64
	WorkingDay   string
	FormName     string
	Salary       decimal.Decimal
	Formal       string
	Koststed     string
	Aktivitet    string
	Clockings    []Clocking
}

type MinWinTid struct {
	Ident      string
	ResourceID string
	Approver   string
	Timesheet  map[string]TimeSheet
	Satser     map[string]decimal.Decimal
}
