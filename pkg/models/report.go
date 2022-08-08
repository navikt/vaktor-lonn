package models

// GuardDuty keeps track of minutes not worked in a given guard duty
type GuardDuty struct {
	Hvilende2006                 int
	Hvilende0620                 int
	Helgetillegg                 int
	Skifttillegg                 int
	WeekendOrHolidayCompensation bool
}

type Timesheet struct {
	GuardDuty       Period
	MinutesWithDuty GuardDuty
	Work            []string
	MinutesWorked   GuardDuty
}

type Compensation struct {
	Total float64
}

type Overtime struct {
	Work    float64
	Weekend float64
	Total   float64
}

type Earnings struct {
	Overtime     Overtime
	Compensation Compensation
	Total        float64
}

type Report struct {
	Ident                           string
	Salary                          float64
	Satser                          map[string]float64
	Earnings                        Earnings
	TimesheetEachDay                map[string]Timesheet
	GuardDutyMinutes                GuardDuty
	GuardDutyHours                  GuardDuty
	OTS50                           float64
	OTS100                          float64
	MinutesNotWorkedinCoreWorkHours int
	TooMuchDutyMinutes              int
}
