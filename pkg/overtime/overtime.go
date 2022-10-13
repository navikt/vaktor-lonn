package overtime

import (
	"fmt"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/shopspring/decimal"
)

func getSalary(timesheet map[string]models.TimeSheet) (decimal.Decimal, error) {
	var salary decimal.Decimal
	for _, period := range timesheet {
		if salary.IsZero() {
			salary = period.Salary
			continue
		}
		if !salary.Equal(period.Salary) {
			return decimal.Decimal{}, fmt.Errorf("salary has changed")
		}
	}

	return salary, nil
}

func Calculate(minutes map[string]models.GuardDuty, timesheet map[string]models.TimeSheet, payroll *models.Payroll) error {
	salary, err := getSalary(timesheet)
	if err != nil {
		return err
	}

	overtimeWeekendOrHolidayMinutes := 0.0
	overtimeWorkDayMinutes := 0.0
	overtimeWorkEveningMinutes := 0.0
	overtimeWorkMorningMinutes := 0.0

	for _, duty := range minutes {
		if duty.WeekendOrHolidayCompensation {
			overtimeWeekendOrHolidayMinutes += duty.Hvilende0620 + duty.Hvilende2000 + duty.Hvilende0006
		} else {
			overtimeWorkDayMinutes += duty.Hvilende0620
			overtimeWorkEveningMinutes += duty.Hvilende2000
			overtimeWorkMorningMinutes += duty.Hvilende0006
		}
	}

	ots50 := salary.Div(decimal.NewFromInt(1850)).Mul(decimal.NewFromFloat(1.5))
	ots100 := salary.Div(decimal.NewFromInt(1850)).Mul(decimal.NewFromInt(2))

	payroll.TypeCodes[models.ArtskodeDag] = payroll.TypeCodes[models.ArtskodeDag].Add(decimal.NewFromFloat(overtimeWorkDayMinutes).Div(decimal.NewFromInt(60)).Mul(ots50).Div(decimal.NewFromInt(5)).Round(2))
	payroll.TypeCodes[models.ArtskodeMorgen] = payroll.TypeCodes[models.ArtskodeMorgen].Add(decimal.NewFromFloat(overtimeWorkMorningMinutes).Div(decimal.NewFromInt(60)).Mul(ots100).Div(decimal.NewFromInt(5)).Round(2))
	payroll.TypeCodes[models.ArtskodeKveld] = payroll.TypeCodes[models.ArtskodeKveld].Add(decimal.NewFromFloat(overtimeWorkEveningMinutes).Div(decimal.NewFromInt(60)).Mul(ots100).Div(decimal.NewFromInt(5)).Round(2))
	payroll.TypeCodes[models.ArtskodeHelg] = payroll.TypeCodes[models.ArtskodeHelg].Add(decimal.NewFromFloat(overtimeWeekendOrHolidayMinutes).Div(decimal.NewFromInt(60)).Mul(ots100).Div(decimal.NewFromInt(5)).Round(2))
	return nil
}
