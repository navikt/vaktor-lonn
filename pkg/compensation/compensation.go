package compensation

import (
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/navikt/vaktor-lonn/pkg/ranges"
	"github.com/shopspring/decimal"
	"time"
)

func Calculate(minutes map[string]models.GuardDuty, satser map[string]decimal.Decimal, payroll *models.Payroll) {
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

	compensationDay := decimal.NewFromInt(int64(compensation.Hvilende0620)).DivRound(minutesInHour, 0).Mul(satser["0620"]).Round(2)
	payroll.TypeCodes[models.ArtskodeDag] = payroll.TypeCodes[models.ArtskodeDag].Add(compensationDay)

	compensationEvening := decimal.NewFromInt(int64(compensation.Hvilende2000)).DivRound(minutesInHour, 0).Mul(satser["2006"]).Round(2)
	payroll.TypeCodes[models.ArtskodeKveld] = payroll.TypeCodes[models.ArtskodeKveld].Add(compensationEvening)

	compensationMorning := decimal.NewFromInt(int64(compensation.Hvilende0006)).DivRound(minutesInHour, 0).Mul(satser["2006"]).Round(2)
	payroll.TypeCodes[models.ArtskodeMorgen] = payroll.TypeCodes[models.ArtskodeMorgen].Add(compensationMorning)

	compensationWeekend := decimal.NewFromInt(int64(compensation.Helgetillegg)).DivRound(minutesInHour, 0).Mul(satser["lørsøn"]).Div(fifthOfAnHour).Round(2)
	payroll.TypeCodes[models.ArtskodeHelg] = payroll.TypeCodes[models.ArtskodeHelg].Add(compensationWeekend)

	compensationShift := decimal.NewFromInt(int64(compensation.Skifttillegg)).DivRound(minutesInHour, 0).Mul(satser["utvidet"]).Div(fifthOfAnHour).Round(2)
	payroll.TypeCodes[models.ArtskodeDag] = payroll.TypeCodes[models.ArtskodeDag].Add(compensationShift)
}

func isWeekend(date time.Time) bool {
	return date.Weekday() == time.Saturday || date.Weekday() == time.Sunday
}

func CalculateCallOut(timesheet map[string]models.TimeSheet, satser map[string]decimal.Decimal, payroll *models.Payroll) {
	// TODO: Validering av at man har vakt i perioden man har overtid med kommentaren BV (for eks. i kjernetid)
	minutesInHour := decimal.NewFromInt(60)

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

					compensation := decimal.NewFromInt(int64(minutesWithGuardDuty)).DivRound(minutesInHour, 0).Mul(satser["2006"]).Round(2)
					payroll.TypeCodes[models.ArtskodeMorgen] = payroll.TypeCodes[models.ArtskodeMorgen].Add(compensation)
				}

				// 06-20
				dutyRange = ranges.CreateForPeriod(guardDutyPeriod, models.Period{
					Begin: time.Date(date.Year(), date.Month(), date.Day(), 6, 0, 0, 0, time.UTC),
					End:   time.Date(date.Year(), date.Month(), date.Day(), 20, 0, 0, 0, time.UTC),
				})
				if dutyRange != nil {
					minutesWithGuardDuty := ranges.CalculateMinutesOverlapping(workRange, *dutyRange)

					compensation := decimal.NewFromInt(int64(minutesWithGuardDuty)).DivRound(minutesInHour, 0).Mul(satser["0620"]).Round(2)
					payroll.TypeCodes[models.ArtskodeDag] = payroll.TypeCodes[models.ArtskodeDag].Add(compensation)
				}

				// 20-00
				dutyRange = ranges.CreateForPeriod(guardDutyPeriod, models.Period{
					Begin: time.Date(date.Year(), date.Month(), date.Day(), 20, 0, 0, 0, time.UTC),
					End:   time.Date(date.Year(), date.Month(), date.Day()+1, 0, 0, 0, 0, time.UTC),
				})
				if dutyRange != nil {
					minutesWithGuardDuty := ranges.CalculateMinutesOverlapping(workRange, *dutyRange)

					compensation := decimal.NewFromInt(int64(minutesWithGuardDuty)).DivRound(minutesInHour, 0).Mul(satser["2006"]).Round(2)
					payroll.TypeCodes[models.ArtskodeKveld] = payroll.TypeCodes[models.ArtskodeKveld].Add(compensation)
				}

				if isWeekend(date) {
					// sjekk om man har vakt i perioden 00-24
					dutyRange = ranges.CreateForPeriod(guardDutyPeriod, models.Period{
						Begin: time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC),
						End:   time.Date(date.Year(), date.Month(), date.Day()+1, 0, 0, 0, 0, time.UTC),
					})

					minutesWithGuardDuty := ranges.CalculateMinutesOverlapping(workRange, *dutyRange)

					compensation := decimal.NewFromInt(int64(minutesWithGuardDuty)).DivRound(minutesInHour, 0).Mul(satser["lørsøn"]).Round(2)
					payroll.TypeCodes[models.ArtskodeHelg] = payroll.TypeCodes[models.ArtskodeHelg].Add(compensation)
				} else {
					minutesWithGuardDuty := 0.0
					// sjekk om man har vakt i perioden 06-07
					dutyRange = ranges.CreateForPeriod(guardDutyPeriod, models.Period{
						Begin: time.Date(date.Year(), date.Month(), date.Day(), 6, 0, 0, 0, time.UTC),
						End:   time.Date(date.Year(), date.Month(), date.Day(), 7, 0, 0, 0, time.UTC),
					})
					if dutyRange != nil {
						minutesWithGuardDuty = ranges.CalculateMinutesOverlapping(workRange, *dutyRange)
					}

					// sjekk om man har vakt i perioden 17-20
					dutyRange = ranges.CreateForPeriod(guardDutyPeriod, models.Period{
						Begin: time.Date(date.Year(), date.Month(), date.Day(), 17, 0, 0, 0, time.UTC),
						End:   time.Date(date.Year(), date.Month(), date.Day(), 20, 0, 0, 0, time.UTC),
					})
					if dutyRange != nil {
						minutesWithGuardDuty += ranges.CalculateMinutesOverlapping(workRange, *dutyRange)
					}

					compensation := decimal.NewFromInt(int64(minutesWithGuardDuty)).DivRound(minutesInHour, 0).Mul(satser["utvidet"]).Round(2)
					payroll.TypeCodes[models.ArtskodeDag] = payroll.TypeCodes[models.ArtskodeDag].Add(compensation)
				}
			}
		}
	}
}
