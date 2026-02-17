package kronetillegg

import (
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/shopspring/decimal"
)

func Calculate(minutes map[string]models.GuardDuty, satser models.Satser, payroll *models.Payroll) {
	kronetilleggWeekendMinutes := 0.0
	kronetilleggWeekendDayMinutes := 0.0
	kronetilleggWeekendEveningMinutes := 0.0
	kronetilleggWeekendMorningMinutes := 0.0
	kronetilleggDayMinutes := 0.0
	kronetilleggEveningMinutes := 0.0
	kronetilleggMorningMinutes := 0.0
	kronetilleggShiftMinutes := 0.0

	for _, duty := range minutes {
		if duty.IsWeekend {
			kronetilleggWeekendMinutes += duty.Helgetillegg
			kronetilleggWeekendDayMinutes += duty.Hvilende0620 + duty.Helligdag0620
			kronetilleggWeekendMorningMinutes += duty.Hvilende0006
			kronetilleggWeekendEveningMinutes += duty.Hvilende2000
		} else {
			kronetilleggDayMinutes += duty.Hvilende0620 + duty.Helligdag0620
			kronetilleggEveningMinutes += duty.Hvilende2000
			kronetilleggMorningMinutes += duty.Hvilende0006
			kronetilleggShiftMinutes += duty.Skifttillegg
		}
	}

	minutesInHour := decimal.NewFromInt(60)
	fifthOfAnHour := decimal.NewFromInt(5)

	kronetilleggDayHours := decimal.NewFromInt(int64(kronetilleggDayMinutes)).DivRound(minutesInHour, 0)
	kronetilleggDay := kronetilleggDayHours.Mul(satser.Dag).Round(2)
	payroll.Artskoder.Dag.Sum = payroll.Artskoder.Dag.Sum.Add(kronetilleggDay)

	kronetilleggEveningHours := decimal.NewFromInt(int64(kronetilleggEveningMinutes)).DivRound(minutesInHour, 0)
	kronetilleggEvening := kronetilleggEveningHours.Mul(satser.Natt).Round(2)
	payroll.Artskoder.Kveld.Sum = payroll.Artskoder.Kveld.Sum.Add(kronetilleggEvening)

	kronetilleggMorningHours := decimal.NewFromInt(int64(kronetilleggMorningMinutes)).DivRound(minutesInHour, 0)
	kronetilleggMorning := kronetilleggMorningHours.Mul(satser.Natt).Round(2)
	payroll.Artskoder.Morgen.Sum = payroll.Artskoder.Morgen.Sum.Add(kronetilleggMorning)

	kronetilleggWeekendHours := decimal.NewFromInt(int64(kronetilleggWeekendMinutes)).DivRound(minutesInHour, 0)
	kronetilleggWeekend := kronetilleggWeekendHours.Mul(satser.Helg).Div(fifthOfAnHour).Round(2)
	payroll.Artskoder.Helg.Sum = payroll.Artskoder.Helg.Sum.Add(kronetilleggWeekend)

	kronetilleggWeekendDayHours := decimal.NewFromInt(int64(kronetilleggWeekendDayMinutes)).DivRound(minutesInHour, 0)
	kronetilleggWeekendDay := kronetilleggWeekendDayHours.Mul(satser.Dag).Round(2)
	payroll.Artskoder.Helg.Sum = payroll.Artskoder.Helg.Sum.Add(kronetilleggWeekendDay)

	kronetilleggWeekendEveningHours := decimal.NewFromInt(int64(kronetilleggWeekendEveningMinutes)).DivRound(minutesInHour, 0)
	kronetilleggWeekendEvening := kronetilleggWeekendEveningHours.Mul(satser.Natt).Round(2)
	payroll.Artskoder.Helg.Sum = payroll.Artskoder.Helg.Sum.Add(kronetilleggWeekendEvening)

	kronetilleggWeekendMorningHours := decimal.NewFromInt(int64(kronetilleggWeekendMorningMinutes)).DivRound(minutesInHour, 0)
	kronetilleggWeekendMorning := kronetilleggWeekendMorningHours.Mul(satser.Natt).Round(2)
	payroll.Artskoder.Helg.Sum = payroll.Artskoder.Helg.Sum.Add(kronetilleggWeekendMorning)

	kronetilleggShiftHours := decimal.NewFromInt(int64(kronetilleggShiftMinutes)).DivRound(minutesInHour, 0)
	payroll.Artskoder.Skift.Hours = kronetilleggShiftHours.IntPart()
	kronetilleggShift := kronetilleggShiftHours.Mul(satser.Utvidet).Div(fifthOfAnHour).Round(2)
	payroll.Artskoder.Skift.Sum = payroll.Artskoder.Skift.Sum.Add(kronetilleggShift)
}
