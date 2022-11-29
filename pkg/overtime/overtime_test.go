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
			name: "Beredskapsvakt helg",
			args: args{
				payroll: &models.Payroll{},
				salary:  decimal.NewFromInt(750_000),
				minutes: map[string]models.GuardDuty{
					"2022-10-15": {
						Hvilende2000:                 360,
						Hvilende0006:                 840,
						Hvilende0620:                 240,
						Helgetillegg:                 1440,
						Skifttillegg:                 0,
						WeekendOrHolidayCompensation: true,
					},
					"2022-10-16": {
						Hvilende2000:                 360,
						Hvilende0006:                 840,
						Hvilende0620:                 240,
						Helgetillegg:                 1440,
						Skifttillegg:                 0,
						WeekendOrHolidayCompensation: true,
					},
				},
			},
			want: models.Artskoder{
				Morgen: models.Artskode{
					Sum:   decimal.NewFromInt(0),
					Hours: 0,
				},
				Dag: models.Artskode{
					Sum:   decimal.NewFromInt(0),
					Hours: 0,
				},
				Kveld: models.Artskode{
					Sum:   decimal.NewFromInt(0),
					Hours: 0,
				},
				Helg: models.Artskode{
					Sum:   decimal.NewFromFloat(7_783.78),
					Hours: 48,
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
