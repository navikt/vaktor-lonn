package models

import "github.com/shopspring/decimal"

// GuardDuty keeps track of minutes not worked in a given guard duty
type GuardDuty struct {
	Hvilende2006                 int
	Hvilende0620                 int
	Helgetillegg                 int
	Skifttillegg                 int
	WeekendOrHolidayCompensation bool
}

type Timesheet struct {
	Schedule        []Period
	MinutesWithDuty GuardDuty
	Work            []string
	MinutesWorked   GuardDuty
}

type Compensation struct {
	Total decimal.Decimal
}

type Overtime struct {
	Work    decimal.Decimal
	Weekend decimal.Decimal
	Total   decimal.Decimal
}

type Earnings struct {
	Overtime     Overtime
	Compensation Compensation
	Total        decimal.Decimal
}

type Report struct {
	Ident                           string
	Salary                          decimal.Decimal
	Satser                          map[string]float64
	Earnings                        Earnings
	TimesheetEachDay                map[string]Timesheet
	GuardDutyMinutes                GuardDuty
	GuardDutyHours                  GuardDuty
	OTS50                           decimal.Decimal
	OTS100                          decimal.Decimal
	MinutesNotWorkedinCoreWorkHours int
	TooMuchDutyMinutes              int
}
