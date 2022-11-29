package overtime

import (
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/shopspring/decimal"
)

func Calculate(minutes map[string]models.GuardDuty, salary decimal.Decimal, payroll *models.Payroll) {
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
	payroll.Artskoder.Dag.Hours = overtimeDayHours.IntPart()
	overtimeDay := overtimeDayHours.Mul(ots50).Div(fifthOfAnHour).Round(2)
	payroll.Artskoder.Dag.Sum = payroll.Artskoder.Dag.Sum.Add(overtimeDay)

	overtimeMorningHours := decimal.NewFromFloat(overtimeMorningMinutes).DivRound(minutesInHour, 0)
	payroll.Artskoder.Morgen.Hours = overtimeMorningHours.IntPart()
	overtimeMorning := overtimeMorningHours.Mul(ots100).Div(fifthOfAnHour).Round(2)
	payroll.Artskoder.Morgen.Sum = payroll.Artskoder.Morgen.Sum.Add(overtimeMorning)

	overtimeEveningHours := decimal.NewFromFloat(overtimeEveningMinutes).DivRound(minutesInHour, 0)
	payroll.Artskoder.Kveld.Hours = overtimeEveningHours.IntPart()
	overtimeEvening := overtimeEveningHours.Mul(ots100).Div(fifthOfAnHour).Round(2)
	payroll.Artskoder.Kveld.Sum = payroll.Artskoder.Kveld.Sum.Add(overtimeEvening)

	overtimeWeekendOrHolidayHours := decimal.NewFromFloat(overtimeWeekendOrHolidayMinutes).DivRound(minutesInHour, 0)
	payroll.Artskoder.Helg.Hours = overtimeWeekendOrHolidayHours.IntPart()
	overtimeWeekendOrHoliday := overtimeWeekendOrHolidayHours.Mul(ots100).Div(fifthOfAnHour).Round(2)
	payroll.Artskoder.Helg.Sum = payroll.Artskoder.Helg.Sum.Add(overtimeWeekendOrHoliday)
}
