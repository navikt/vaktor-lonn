package compensation

import (
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/shopspring/decimal"
	"math"
)

func Calculate(report *models.Report, minutes map[string]models.GuardDuty, satser map[string]decimal.Decimal) decimal.Decimal {
	var compensationDuty models.GuardDuty

	for _, duty := range minutes {
		compensationDuty.Hvilende0620 += duty.Hvilende0620
		compensationDuty.Hvilende2006 += duty.Hvilende2006
		compensationDuty.Skifttillegg += duty.Skifttillegg
		compensationDuty.Helgetillegg += duty.Helgetillegg
	}

	report.GuardDutyMinutes.Hvilende0620 = compensationDuty.Hvilende0620
	report.GuardDutyMinutes.Hvilende2006 = compensationDuty.Hvilende2006
	report.GuardDutyMinutes.Skifttillegg = compensationDuty.Skifttillegg
	report.GuardDutyMinutes.Helgetillegg = compensationDuty.Helgetillegg
	report.GuardDutyHours.Hvilende0620 = int(math.Round(float64(compensationDuty.Hvilende0620 / 60)))
	report.GuardDutyHours.Hvilende2006 = int(math.Round(float64(compensationDuty.Hvilende2006 / 60)))
	report.GuardDutyHours.Skifttillegg = int(math.Round(float64(compensationDuty.Skifttillegg / 60)))
	report.GuardDutyHours.Helgetillegg = int(math.Round(float64(compensationDuty.Helgetillegg / 60)))

	minutesInHour := decimal.NewFromInt(60)
	compensation := decimal.NewFromInt(int64(compensationDuty.Hvilende0620)).Div(minutesInHour).Mul(satser["0620"]).
		Add(decimal.NewFromInt(int64(compensationDuty.Hvilende2006)).Div(minutesInHour).Mul(satser["2006"])).
		Add(decimal.NewFromInt(int64(compensationDuty.Helgetillegg)).Div(minutesInHour).Mul(satser["lørsøn"]).Div(decimal.NewFromInt(5))).
		Add(decimal.NewFromInt(int64(compensationDuty.Skifttillegg)).Div(minutesInHour).Mul(satser["utvidet"]).Div(decimal.NewFromInt(5)))

	return compensation.Round(2)
}
