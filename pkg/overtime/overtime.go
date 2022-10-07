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

func Calculate(minutes map[string]models.GuardDuty, timesheet map[string]models.TimeSheet) (decimal.Decimal, error) {
	salary, err := getSalary(timesheet)
	if err != nil {
		return decimal.Decimal{}, err
	}

	overtimeWeekendMinutes := 0.0
	overtimeWorkDayMinutes := 0.0
	overtimeWorkNightMinutes := 0.0

	for _, duty := range minutes {
		if duty.WeekendOrHolidayCompensation {
			overtimeWeekendMinutes += float64(duty.Hvilende0620 + duty.Hvilende2000 + duty.Hvilende0006)
		} else {
			overtimeWorkDayMinutes += float64(duty.Hvilende0620)
			overtimeWorkNightMinutes += float64(duty.Hvilende2000 + duty.Hvilende0006)
		}
	}

	ots50 := salary.Div(decimal.NewFromInt(1850)).Mul(decimal.NewFromFloat(1.5))
	ots100 := salary.Div(decimal.NewFromInt(1850)).Mul(decimal.NewFromInt(2))

	overTimeWorkDay := decimal.NewFromFloat(overtimeWorkDayMinutes).Div(decimal.NewFromInt(60)).Mul(ots50)
	overTimeWorkNight := decimal.NewFromFloat(overtimeWorkNightMinutes).Div(decimal.NewFromInt(60)).Mul(ots100)
	overtimeWork := overTimeWorkDay.Add(overTimeWorkNight).Div(decimal.NewFromInt(5))
	overtimeWeekend := decimal.NewFromFloat(overtimeWeekendMinutes).Div(decimal.NewFromInt(60)).Mul(ots100).Div(decimal.NewFromInt(5))
	return overtimeWeekend.Add(overtimeWork).Round(2), nil
}
