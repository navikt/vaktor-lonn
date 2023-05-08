package compensation

import (
	"github.com/google/go-cmp/cmp"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/shopspring/decimal"
	"testing"
)

func TestCalculate(t *testing.T) {
	type args struct {
		minutes map[string]models.GuardDuty
		satser  models.Satser
		payroll *models.Payroll
	}
	tests := []struct {
		name string
		args args
		want models.Artskoder
	}{
		{
			name: "Beredskapsvakt en uke",
			args: args{
				payroll: &models.Payroll{},
				satser: models.Satser{
					Helg:    decimal.NewFromInt(65),
					Dag:     decimal.NewFromInt(15),
					Natt:    decimal.NewFromInt(25),
					Utvidet: decimal.NewFromInt(25),
				},
				minutes: map[string]models.GuardDuty{
					"2022-10-12": {
						Hvilende2000:        240,
						Hvilende0006:        0,
						Hvilende0620:        255,
						Helgetillegg:        0,
						Skifttillegg:        180,
						WeekendCompensation: false,
					},
					"2022-10-13": {
						Hvilende2000:        240,
						Hvilende0006:        360,
						Hvilende0620:        375,
						Helgetillegg:        0,
						Skifttillegg:        240,
						WeekendCompensation: false,
					},
					"2022-10-14": {
						Hvilende2000:        240,
						Hvilende0006:        360,
						Hvilende0620:        375,
						Helgetillegg:        0,
						Skifttillegg:        240,
						WeekendCompensation: false,
					},
					"2022-10-15": {
						Hvilende2000:        240,
						Hvilende0006:        360,
						Hvilende0620:        840,
						Helgetillegg:        1440,
						Skifttillegg:        0,
						WeekendCompensation: true,
					},
					"2022-10-16": {
						Hvilende2000:        240,
						Hvilende0006:        360,
						Hvilende0620:        840,
						Helgetillegg:        1440,
						Skifttillegg:        0,
						WeekendCompensation: true,
					},
					"2022-10-17": {
						Hvilende2000:        240,
						Hvilende0006:        360,
						Hvilende0620:        375,
						Helgetillegg:        0,
						Skifttillegg:        240,
						WeekendCompensation: false,
					},
					"2022-10-18": {
						Hvilende2000:        240,
						Hvilende0006:        360,
						Hvilende0620:        375,
						Helgetillegg:        0,
						Skifttillegg:        240,
						WeekendCompensation: false,
					},
					"2022-10-19": {
						Hvilende2000:        0,
						Hvilende0006:        360,
						Hvilende0620:        120,
						Helgetillegg:        0,
						Skifttillegg:        60,
						WeekendCompensation: false,
					},
				},
			},
			want: models.Artskoder{
				Morgen: models.Artskode{
					Sum:   decimal.NewFromInt(750),
					Hours: 30,
				},
				Dag: models.Artskode{
					Sum:   decimal.NewFromInt(465),
					Hours: 31,
				},
				Kveld: models.Artskode{
					Sum:   decimal.NewFromInt(500),
					Hours: 20,
				},
				Helg: models.Artskode{
					Sum:   decimal.NewFromInt(1544),
					Hours: 48,
				},
				Skift: models.Artskode{
					Sum:   decimal.NewFromInt(100),
					Hours: 20,
				},
			},
		},

		{
			name: "Utrykning i helg (søndag)",
			args: args{
				payroll: &models.Payroll{},
				satser: models.Satser{
					Helg:    decimal.NewFromInt(65),
					Dag:     decimal.NewFromInt(15),
					Natt:    decimal.NewFromInt(25),
					Utvidet: decimal.NewFromInt(25),
				},
				minutes: map[string]models.GuardDuty{
					"2022-10-15": {
						Hvilende2000:        240,
						Hvilende0006:        360,
						Hvilende0620:        840,
						Helgetillegg:        1440,
						Skifttillegg:        0,
						WeekendCompensation: true,
					},
					"2022-10-16": {
						Hvilende2000:        120,
						Hvilende0006:        360,
						Hvilende0620:        840,
						Helgetillegg:        1320,
						Skifttillegg:        0,
						WeekendCompensation: true,
					},
				},
			},
			want: models.Artskoder{
				Helg: models.Artskode{
					Sum:   decimal.NewFromInt(1468),
					Hours: 46,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Calculate(tt.args.minutes, tt.args.satser, tt.args.payroll)

			if diff := cmp.Diff(tt.want, tt.args.payroll.Artskoder); diff != "" {
				t.Errorf("Calculate() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
