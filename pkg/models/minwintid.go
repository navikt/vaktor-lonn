package models

import "github.com/shopspring/decimal"

type MinWinTid struct {
	Timesheet map[string][]string
	Salary    decimal.Decimal
	Satser    map[string]decimal.Decimal
}
