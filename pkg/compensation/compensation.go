package compensation

import (
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/shopspring/decimal"
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
