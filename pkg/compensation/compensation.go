package compensation

import (
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/shopspring/decimal"
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
