package models

import (
	"github.com/shopspring/decimal"
	"time"
)

type Clocking struct {
	In  time.Time
	Out time.Time
	OtG bool
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
	Ident        string
	ResourceID   string
	ApproverID   string
	ApproverName string
	Timesheet    map[string]TimeSheet
	Satser       Satser
}
