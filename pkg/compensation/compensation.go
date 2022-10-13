package compensation

import (
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/shopspring/decimal"
)

func Calculate(minutes map[string]models.GuardDuty, satser map[string]decimal.Decimal, payroll models.Payroll) {
	var compensationDuty models.GuardDuty

	for _, duty := range minutes {
		compensationDuty.Hvilende0620 += duty.Hvilende0620
		compensationDuty.Hvilende2000 += duty.Hvilende2000
		compensationDuty.Hvilende0006 += duty.Hvilende0006
		compensationDuty.Skifttillegg += duty.Skifttillegg
		compensationDuty.Helgetillegg += duty.Helgetillegg
	}

	minutesInHour := decimal.NewFromInt(60)
	payroll.TypeCodes[models.ArtskodeDag] = payroll.TypeCodes[models.ArtskodeDag].Add(decimal.NewFromInt(int64(compensationDuty.Hvilende0620)).Div(minutesInHour).Mul(satser["0620"]).Round(2))
	payroll.TypeCodes[models.ArtskodeKveld] = payroll.TypeCodes[models.ArtskodeKveld].Add(decimal.NewFromInt(int64(compensationDuty.Hvilende2000)).Div(minutesInHour).Mul(satser["2006"]).Round(2))
	payroll.TypeCodes[models.ArtskodeMorgen] = payroll.TypeCodes[models.ArtskodeMorgen].Add(decimal.NewFromInt(int64(compensationDuty.Hvilende0006)).Div(minutesInHour).Mul(satser["2006"]).Round(2))
	payroll.TypeCodes[models.ArtskodeHelg] = payroll.TypeCodes[models.ArtskodeHelg].Add(decimal.NewFromInt(int64(compensationDuty.Helgetillegg)).Div(minutesInHour).Mul(satser["lørsøn"]).Div(decimal.NewFromInt(5)).Round(2))
	payroll.TypeCodes[models.ArtskodeDag] = payroll.TypeCodes[models.ArtskodeDag].Add(decimal.NewFromInt(int64(compensationDuty.Skifttillegg)).Div(minutesInHour).Mul(satser["utvidet"]).Div(decimal.NewFromInt(5)).Round(2))
}
