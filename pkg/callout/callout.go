package callout

import (
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/navikt/vaktor-lonn/pkg/ranges"
	"github.com/shopspring/decimal"
	"time"
)

func Calculate(schedule map[string][]models.Period, timesheet map[string]models.TimeSheet, satser models.Satser, payroll *models.Payroll) {
	minutesInHour := decimal.NewFromInt(60)
	fourFifthOfAnHour := decimal.NewFromFloat(0.8)
	guardMinutes := models.GuardDuty{}

	for day, sheet := range timesheet {
		date := sheet.Date
		for _, clocking := range sheet.Clockings {
			if clocking.OtG && (date.Weekday() == time.Saturday || date.Weekday() == time.Sunday) {
				for _, guardDutyPeriod := range schedule[day] {
					workRange := ranges.FromTime(clocking.In, clocking.Out)

					dutyRange := ranges.CreateForPeriod(guardDutyPeriod, models.Period{
						Begin: time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC),
						End:   time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC).Add(24 * time.Hour),
					})

					minutesWithGuardDuty := ranges.CalculateMinutesOverlapping(workRange, *dutyRange)
					guardMinutes.Helgetillegg += minutesWithGuardDuty
				}
			}
		}
	}

	hours := decimal.NewFromInt(int64(guardMinutes.Helgetillegg)).DivRound(minutesInHour, 0)
	payroll.Artskoder.Utrykning.Hours = hours.IntPart()
	compensation := hours.Mul(satser.Helg).Mul(fourFifthOfAnHour).Round(2)
	payroll.Artskoder.Utrykning.Sum = payroll.Artskoder.Utrykning.Sum.Add(compensation)
}
