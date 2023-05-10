package callout

import (
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/navikt/vaktor-lonn/pkg/ranges"
	"github.com/shopspring/decimal"
	"time"
)

func Calculate(timesheet map[string]models.TimeSheet, satser models.Satser, payroll *models.Payroll) {
	// TODO: Validering av at man har vakt i perioden man har overtid med kommentaren BV (for eks. i kjernetid)
	minutesInHour := decimal.NewFromInt(60)
	fourFifthOfAnHour := decimal.NewFromFloat(0.8)
	guardMinutes := models.GuardDuty{}

	for _, sheet := range timesheet {
		date := sheet.Date
		for _, clocking := range sheet.Clockings {
			if clocking.OtG {
				guardDutyPeriod := models.Period{
					Begin: clocking.In,
					End:   clocking.Out,
				}
				workRange := ranges.FromTime(clocking.In, clocking.Out)

				if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
					// sjekk om man har vakt i perioden 00-24 i helgen
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
