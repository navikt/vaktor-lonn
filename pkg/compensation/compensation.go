package compensation

import (
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/navikt/vaktor-lonn/pkg/ranges"
	"github.com/shopspring/decimal"
	"time"
)

func Calculate(minutes map[string]models.GuardDuty, satser models.Satser, payroll *models.Payroll) {
	var compensation models.GuardDuty

	for _, duty := range minutes {
		compensation.Hvilende0620 += duty.Hvilende0620
		compensation.Hvilende2000 += duty.Hvilende2000
		compensation.Hvilende0006 += duty.Hvilende0006
		compensation.Skifttillegg += duty.Skifttillegg
		compensation.Helgetillegg += duty.Helgetillegg
	}

	minutesInHour := decimal.NewFromInt(60)
	fifthOfAnHour := decimal.NewFromInt(5)

	compensationDay := decimal.NewFromInt(int64(compensation.Hvilende0620)).DivRound(minutesInHour, 0).Mul(satser.Dag).Round(2)
	payroll.Artskoder.Dag = payroll.Artskoder.Dag.Add(compensationDay)

	compensationEvening := decimal.NewFromInt(int64(compensation.Hvilende2000)).DivRound(minutesInHour, 0).Mul(satser.Natt).Round(2)
	payroll.Artskoder.Kveld = payroll.Artskoder.Kveld.Add(compensationEvening)

	compensationMorning := decimal.NewFromInt(int64(compensation.Hvilende0006)).DivRound(minutesInHour, 0).Mul(satser.Natt).Round(2)
	payroll.Artskoder.Morgen = payroll.Artskoder.Morgen.Add(compensationMorning)

	compensationWeekend := decimal.NewFromInt(int64(compensation.Helgetillegg)).DivRound(minutesInHour, 0).Mul(satser.Helg).Div(fifthOfAnHour).Round(2)
	payroll.Artskoder.Helg = payroll.Artskoder.Helg.Add(compensationWeekend)

	compensationShift := decimal.NewFromInt(int64(compensation.Skifttillegg)).DivRound(minutesInHour, 0).Mul(satser.Utvidet).Div(fifthOfAnHour).Round(2)
	payroll.Artskoder.Dag = payroll.Artskoder.Dag.Add(compensationShift)
}

func isWeekend(date time.Time) bool {
	return date.Weekday() == time.Saturday || date.Weekday() == time.Sunday
}

func CalculateCallOut(timesheet map[string]models.TimeSheet, satser models.Satser, payroll *models.Payroll) {
	// TODO: Validering av at man har vakt i perioden man har overtid med kommentaren BV (for eks. i kjernetid)
	minutesInHour := decimal.NewFromInt(60)
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

				// 00-06
				dutyRange := ranges.CreateForPeriod(guardDutyPeriod, models.Period{
					Begin: time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC),
					End:   time.Date(date.Year(), date.Month(), date.Day(), 6, 0, 0, 0, time.UTC),
				})
				if dutyRange != nil {
					minutesWithGuardDuty := ranges.CalculateMinutesOverlapping(workRange, *dutyRange)
					guardMinutes.Hvilende0006 += minutesWithGuardDuty
				}

				// 06-20
				dutyRange = ranges.CreateForPeriod(guardDutyPeriod, models.Period{
					Begin: time.Date(date.Year(), date.Month(), date.Day(), 6, 0, 0, 0, time.UTC),
					End:   time.Date(date.Year(), date.Month(), date.Day(), 20, 0, 0, 0, time.UTC),
				})
				if dutyRange != nil {
					minutesWithGuardDuty := ranges.CalculateMinutesOverlapping(workRange, *dutyRange)
					guardMinutes.Hvilende0620 += minutesWithGuardDuty
				}

				// 20-00
				dutyRange = ranges.CreateForPeriod(guardDutyPeriod, models.Period{
					Begin: time.Date(date.Year(), date.Month(), date.Day(), 20, 0, 0, 0, time.UTC),
					End:   time.Date(date.Year(), date.Month(), date.Day()+1, 0, 0, 0, 0, time.UTC),
				})
				if dutyRange != nil {
					minutesWithGuardDuty := ranges.CalculateMinutesOverlapping(workRange, *dutyRange)
					guardMinutes.Hvilende2000 += minutesWithGuardDuty
				}

				if isWeekend(date) {
					// sjekk om man har vakt i perioden 00-24
					dutyRange = ranges.CreateForPeriod(guardDutyPeriod, models.Period{
						Begin: time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC),
						End:   time.Date(date.Year(), date.Month(), date.Day()+1, 0, 0, 0, 0, time.UTC),
					})

					minutesWithGuardDuty := ranges.CalculateMinutesOverlapping(workRange, *dutyRange)
					guardMinutes.Helgetillegg += minutesWithGuardDuty
				} else {
					minutesWithGuardDuty := 0.0
					// sjekk om man har vakt i perioden 06-07
					dutyRange = ranges.CreateForPeriod(guardDutyPeriod, models.Period{
						Begin: time.Date(date.Year(), date.Month(), date.Day(), 6, 0, 0, 0, time.UTC),
						End:   time.Date(date.Year(), date.Month(), date.Day(), 7, 0, 0, 0, time.UTC),
					})
					if dutyRange != nil {
						minutesWithGuardDuty = ranges.CalculateMinutesOverlapping(workRange, *dutyRange)
						guardMinutes.Skifttillegg += minutesWithGuardDuty
					}

					// sjekk om man har vakt i perioden 17-20
					dutyRange = ranges.CreateForPeriod(guardDutyPeriod, models.Period{
						Begin: time.Date(date.Year(), date.Month(), date.Day(), 17, 0, 0, 0, time.UTC),
						End:   time.Date(date.Year(), date.Month(), date.Day(), 20, 0, 0, 0, time.UTC),
					})
					if dutyRange != nil {
						minutesWithGuardDuty += ranges.CalculateMinutesOverlapping(workRange, *dutyRange)
						guardMinutes.Skifttillegg += minutesWithGuardDuty
					}
				}
			}
		}
	}

	compensation := decimal.NewFromInt(int64(guardMinutes.Hvilende0006)).DivRound(minutesInHour, 0).Mul(satser.Natt).Round(2)
	payroll.Artskoder.Morgen = payroll.Artskoder.Morgen.Add(compensation)

	compensation = decimal.NewFromInt(int64(guardMinutes.Hvilende0620)).DivRound(minutesInHour, 0).Mul(satser.Dag).Round(2)
	payroll.Artskoder.Dag = payroll.Artskoder.Dag.Add(compensation)

	compensation = decimal.NewFromInt(int64(guardMinutes.Hvilende2000)).DivRound(minutesInHour, 0).Mul(satser.Natt).Round(2)
	payroll.Artskoder.Kveld = payroll.Artskoder.Kveld.Add(compensation)

	compensation = decimal.NewFromInt(int64(guardMinutes.Helgetillegg)).DivRound(minutesInHour, 0).Mul(satser.Helg).Round(2)
	payroll.Artskoder.Helg = payroll.Artskoder.Helg.Add(compensation)

	compensation = decimal.NewFromInt(int64(guardMinutes.Skifttillegg)).DivRound(minutesInHour, 0).Mul(satser.Utvidet).Round(2)
	payroll.Artskoder.Dag = payroll.Artskoder.Dag.Add(compensation)
}
