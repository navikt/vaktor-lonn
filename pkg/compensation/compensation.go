package compensation

import (
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/navikt/vaktor-lonn/pkg/ranges"
	"github.com/shopspring/decimal"
	"time"
)

func Calculate(minutes map[string]models.GuardDuty, satser models.Satser, payroll *models.Payroll) {
	compensationWeekendMinutes := 0.0
	compensationWeekendDayMinutes := 0.0
	compensationWeekendEveningMinutes := 0.0
	compensationWeekendMorningMinutes := 0.0
	compensationDayMinutes := 0.0
	compensationEveningMinutes := 0.0
	compensationMorningMinutes := 0.0
	compensationShiftMinutes := 0.0

	for _, duty := range minutes {
		if duty.WeekendCompensation {
			compensationWeekendMinutes += duty.Helgetillegg
			compensationWeekendDayMinutes += duty.Hvilende0620
			compensationWeekendMorningMinutes += duty.Hvilende0006
			compensationWeekendEveningMinutes += duty.Hvilende2000
		} else {
			compensationDayMinutes += duty.Hvilende0620
			compensationEveningMinutes += duty.Hvilende2000
			compensationMorningMinutes += duty.Hvilende0006
			compensationShiftMinutes += duty.Skifttillegg
		}
	}

	minutesInHour := decimal.NewFromInt(60)
	fifthOfAnHour := decimal.NewFromInt(5)

	compensationDayHours := decimal.NewFromInt(int64(compensationDayMinutes)).DivRound(minutesInHour, 0)
	payroll.Artskoder.Dag.Hours = compensationDayHours.IntPart()
	compensationDay := compensationDayHours.Mul(satser.Dag).Round(2)
	payroll.Artskoder.Dag.Sum = payroll.Artskoder.Dag.Sum.Add(compensationDay)

	compensationEveningHours := decimal.NewFromInt(int64(compensationEveningMinutes)).DivRound(minutesInHour, 0)
	payroll.Artskoder.Kveld.Hours = compensationEveningHours.IntPart()
	compensationEvening := compensationEveningHours.Mul(satser.Natt).Round(2)
	payroll.Artskoder.Kveld.Sum = payroll.Artskoder.Kveld.Sum.Add(compensationEvening)

	compensationMorningHours := decimal.NewFromInt(int64(compensationMorningMinutes)).DivRound(minutesInHour, 0)
	payroll.Artskoder.Morgen.Hours = compensationMorningHours.IntPart()
	compensationMorning := compensationMorningHours.Mul(satser.Natt).Round(2)
	payroll.Artskoder.Morgen.Sum = payroll.Artskoder.Morgen.Sum.Add(compensationMorning)

	compensationWeekendHours := decimal.NewFromInt(int64(compensationWeekendMinutes)).DivRound(minutesInHour, 0)
	payroll.Artskoder.Helg.Hours = compensationWeekendHours.IntPart()
	compensationWeekend := compensationWeekendHours.Mul(satser.Helg).Div(fifthOfAnHour).Round(2)
	payroll.Artskoder.Helg.Sum = payroll.Artskoder.Helg.Sum.Add(compensationWeekend)

	compensationWeekendDayHours := decimal.NewFromInt(int64(compensationWeekendDayMinutes)).DivRound(minutesInHour, 0)
	compensationWeekendDay := compensationWeekendDayHours.Mul(satser.Dag).Round(2)
	payroll.Artskoder.Helg.Sum = payroll.Artskoder.Helg.Sum.Add(compensationWeekendDay)

	compensationWeekendEveningHours := decimal.NewFromInt(int64(compensationWeekendEveningMinutes)).DivRound(minutesInHour, 0)
	compensationWeekendEvening := compensationWeekendEveningHours.Mul(satser.Natt).Round(2)
	payroll.Artskoder.Helg.Sum = payroll.Artskoder.Helg.Sum.Add(compensationWeekendEvening)

	compensationWeekendMorningHours := decimal.NewFromInt(int64(compensationWeekendMorningMinutes)).DivRound(minutesInHour, 0)
	compensationWeekendMorning := compensationWeekendMorningHours.Mul(satser.Natt).Round(2)
	payroll.Artskoder.Helg.Sum = payroll.Artskoder.Helg.Sum.Add(compensationWeekendMorning)

	compensationShiftHours := decimal.NewFromInt(int64(compensationShiftMinutes)).DivRound(minutesInHour, 0)
	payroll.Artskoder.Skift.Hours = compensationShiftHours.IntPart()
	compensationShift := compensationShiftHours.Mul(satser.Utvidet).Div(fifthOfAnHour).Round(2)
	payroll.Artskoder.Skift.Sum = payroll.Artskoder.Skift.Sum.Add(compensationShift)
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
					End:   time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC).Add(24 * time.Hour),
				})
				if dutyRange != nil {
					minutesWithGuardDuty := ranges.CalculateMinutesOverlapping(workRange, *dutyRange)
					guardMinutes.Hvilende2000 += minutesWithGuardDuty
				}

				if isWeekend(date) {
					// sjekk om man har vakt i perioden 00-24
					dutyRange = ranges.CreateForPeriod(guardDutyPeriod, models.Period{
						Begin: time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC),
						End:   time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC).Add(24 * time.Hour),
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

	hours := decimal.NewFromInt(int64(guardMinutes.Hvilende0006)).DivRound(minutesInHour, 0)
	payroll.Artskoder.Utrykning.Hours += hours.IntPart()
	compensation := hours.Mul(satser.Natt).Round(2)
	payroll.Artskoder.Utrykning.Sum = payroll.Artskoder.Utrykning.Sum.Add(compensation)

	hours = decimal.NewFromInt(int64(guardMinutes.Hvilende0620)).DivRound(minutesInHour, 0)
	payroll.Artskoder.Utrykning.Hours += hours.IntPart()
	compensation = hours.Mul(satser.Dag).Round(2)
	payroll.Artskoder.Utrykning.Sum = payroll.Artskoder.Utrykning.Sum.Add(compensation)

	hours = decimal.NewFromInt(int64(guardMinutes.Hvilende2000)).DivRound(minutesInHour, 0)
	payroll.Artskoder.Utrykning.Hours += hours.IntPart()
	compensation = hours.Mul(satser.Natt).Round(2)
	payroll.Artskoder.Utrykning.Sum = payroll.Artskoder.Utrykning.Sum.Add(compensation)

	hours = decimal.NewFromInt(int64(guardMinutes.Helgetillegg)).DivRound(minutesInHour, 0)
	compensation = hours.Mul(satser.Helg).Round(2)
	payroll.Artskoder.Utrykning.Sum = payroll.Artskoder.Utrykning.Sum.Add(compensation)

	hours = decimal.NewFromInt(int64(guardMinutes.Skifttillegg)).DivRound(minutesInHour, 0)
	compensation = hours.Mul(satser.Utvidet).Round(2)
	payroll.Artskoder.Utrykning.Sum = payroll.Artskoder.Utrykning.Sum.Add(compensation)
}
