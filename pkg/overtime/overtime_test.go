package overtime

import (
	"github.com/google/go-cmp/cmp"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/shopspring/decimal"
	"testing"
)

func TestCalculate(t *testing.T) {
	type args struct {
		minutes map[string]models.GuardDuty
		salary  decimal.Decimal
		payroll *models.Payroll
	}
	tests := []struct {
		name    string
		args    args
		want    models.Artskoder
		wantErr bool
	}{
		{
			name: "Beredskapsvakt en uke",
			args: args{
				payroll: &models.Payroll{},
				salary:  decimal.NewFromInt(750_000),
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
					Sum:   decimal.NewFromFloat(4_864.86),
					Hours: 30,
				},
				Dag: models.Artskode{
					Sum:   decimal.NewFromFloat(3_770.27),
					Hours: 31,
				},
				Kveld: models.Artskode{
					Sum:   decimal.NewFromFloat(3_243.24),
					Hours: 20,
				},
				Helg: models.Artskode{
					Sum:   decimal.NewFromFloat(7_783.78),
					Hours: 48,
				},
			},
			wantErr: false,
		},

		{
			name: "Utrykning i helg (søndag)",
			args: args{
				payroll: &models.Payroll{},
				salary:  decimal.NewFromInt(750_000),
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
					Sum:   decimal.NewFromFloat(7_459.46),
					Hours: 46,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Calculate(tt.args.minutes, tt.args.salary, tt.args.payroll)
			if diff := cmp.Diff(tt.want, tt.args.payroll.Artskoder); diff != "" {
				t.Errorf("Calculate() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
