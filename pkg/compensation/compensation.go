package compensation

import (
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/shopspring/decimal"
)

func Calculate(minutes map[string]models.GuardDuty, satser map[string]decimal.Decimal) decimal.Decimal {
	var compensationDuty models.GuardDuty

	for _, duty := range minutes {
		compensationDuty.Hvilende0620 += duty.Hvilende0620
		compensationDuty.Hvilende2000 += duty.Hvilende2000
		compensationDuty.Hvilende0006 += duty.Hvilende0006
		compensationDuty.Skifttillegg += duty.Skifttillegg
		compensationDuty.Helgetillegg += duty.Helgetillegg
	}

	minutesInHour := decimal.NewFromInt(60)
	compensation := decimal.NewFromInt(int64(compensationDuty.Hvilende0620)).Div(minutesInHour).Mul(satser["0620"]).
		Add(decimal.NewFromInt(int64(compensationDuty.Hvilende2000 + compensationDuty.Hvilende0006)).Div(minutesInHour).Mul(satser["2006"])).
		Add(decimal.NewFromInt(int64(compensationDuty.Helgetillegg)).Div(minutesInHour).Mul(satser["lørsøn"]).Div(decimal.NewFromInt(5))).
		Add(decimal.NewFromInt(int64(compensationDuty.Skifttillegg)).Div(minutesInHour).Mul(satser["utvidet"]).Div(decimal.NewFromInt(5)))

	return compensation.Round(2)
}
