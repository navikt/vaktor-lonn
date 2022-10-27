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
	overtimeDayMinutes := 0.0
	overtimeEveningMinutes := 0.0
	overtimeMorningMinutes := 0.0

	for _, duty := range minutes {
		if duty.WeekendOrHolidayCompensation {
			overtimeWeekendOrHolidayMinutes += duty.Hvilende0620 + duty.Hvilende2000 + duty.Hvilende0006
		} else {
			overtimeDayMinutes += duty.Hvilende0620
			overtimeEveningMinutes += duty.Hvilende2000
			overtimeMorningMinutes += duty.Hvilende0006
		}
	}

	ots50 := salary.Div(decimal.NewFromInt(1850)).Mul(decimal.NewFromFloat(1.5))
	ots100 := salary.Div(decimal.NewFromInt(1850)).Mul(decimal.NewFromInt(2))

	minutesInHour := decimal.NewFromInt(60)
	fifthOfAnHour := decimal.NewFromInt(5)

	overtimeDayHours := decimal.NewFromFloat(overtimeDayMinutes).DivRound(minutesInHour, 0)
	overtimeDay := overtimeDayHours.Mul(ots50).Div(fifthOfAnHour).Round(2)
	payroll.TypeCodes[models.ArtskodeDag] = payroll.TypeCodes[models.ArtskodeDag].Add(overtimeDay)

	overtimeMorningHours := decimal.NewFromFloat(overtimeMorningMinutes).DivRound(minutesInHour, 0)
	overtimeMorning := overtimeMorningHours.Mul(ots100).Div(fifthOfAnHour).Round(2)
	payroll.TypeCodes[models.ArtskodeMorgen] = payroll.TypeCodes[models.ArtskodeMorgen].Add(overtimeMorning)

	overtimeEveningHours := decimal.NewFromFloat(overtimeEveningMinutes).DivRound(minutesInHour, 0)
	overtimeEvening := overtimeEveningHours.Mul(ots100).Div(fifthOfAnHour).Round(2)
	payroll.TypeCodes[models.ArtskodeKveld] = payroll.TypeCodes[models.ArtskodeKveld].Add(overtimeEvening)

	overtimeWeekendOrHolidayHours := decimal.NewFromFloat(overtimeWeekendOrHolidayMinutes).DivRound(minutesInHour, 0)
	overtimeWeekendOrHoliday := overtimeWeekendOrHolidayHours.Mul(ots100).Div(fifthOfAnHour).Round(2)
	payroll.TypeCodes[models.ArtskodeHelg] = payroll.TypeCodes[models.ArtskodeHelg].Add(overtimeWeekendOrHoliday)
	return nil
}
