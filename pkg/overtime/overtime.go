package overtime

import (
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/shopspring/decimal"
)

func Calculate(minutes map[string]models.GuardDuty, salary decimal.Decimal) decimal.Decimal {
	overtimeWeekendMinutes := 0.0
	overtimeWorkDayMinutes := 0.0
	overtimeWorkNightMinutes := 0.0

	for _, duty := range minutes {
		if duty.WeekendOrHolidayCompensation {
			overtimeWeekendMinutes += float64(duty.Hvilende0620 + duty.Hvilende2006)
		} else {
			overtimeWorkDayMinutes += float64(duty.Hvilende0620)
			overtimeWorkNightMinutes += float64(duty.Hvilende2006)
		}
	}

	ots50 := salary.Div(decimal.NewFromInt(1850)).Mul(decimal.NewFromFloat(1.5))
	ots100 := salary.Div(decimal.NewFromInt(1850)).Mul(decimal.NewFromInt(2))

	overTimeWorkDay := decimal.NewFromFloat(overtimeWorkDayMinutes).Div(decimal.NewFromInt(60)).Mul(ots50)
	overTimeWorkNight := decimal.NewFromFloat(overtimeWorkNightMinutes).Div(decimal.NewFromInt(60)).Mul(ots100)
	overtimeWork := overTimeWorkDay.Add(overTimeWorkNight).Div(decimal.NewFromInt(5))
	overtimeWeekend := decimal.NewFromFloat(overtimeWeekendMinutes).Div(decimal.NewFromInt(60)).Mul(ots100).Div(decimal.NewFromInt(5))
	return overtimeWeekend.Add(overtimeWork).Round(2)
}
