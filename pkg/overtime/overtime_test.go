package overtime

import (
	"github.com/google/go-cmp/cmp"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/shopspring/decimal"
	"testing"
	"time"
)

func TestCalculate(t *testing.T) {
	type args struct {
		minutes   map[string]models.GuardDuty
		timesheet map[string]models.TimeSheet
		payroll   *models.Payroll
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]decimal.Decimal
		wantErr bool
	}{
		{
			name: "Utrykning i helg",
			args: args{
				payroll: &models.Payroll{
					TypeCodes: map[string]decimal.Decimal{},
				},
				timesheet: map[string]models.TimeSheet{
					"2022-10-15": {
						Date:       time.Date(2022, 10, 15, 0, 0, 0, 0, time.UTC),
						WorkingDay: "Lørdag",
						FormName:   "BV Lørdag IKT",
						Salary:     decimal.NewFromInt(750_000),
						Koststed:   "855130",
						Formal:     "000000",
						Aktivitet:  "000000",
						Clockings:  []models.Clocking{},
					},
					"2022-10-16": {
						Date:       time.Date(2022, 10, 16, 0, 0, 0, 0, time.UTC),
						WorkingDay: "Søndag",
						FormName:   "BV Søndag IKT",
						Salary:     decimal.NewFromInt(750_000),
						Koststed:   "855130",
						Formal:     "000000",
						Aktivitet:  "000000",
						Clockings:  []models.Clocking{},
					},
				},
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
			want: map[string]decimal.Decimal{
				models.ArtskodeMorgen: decimal.NewFromInt(0),
				models.ArtskodeDag:    decimal.NewFromInt(0),
				models.ArtskodeKveld:  decimal.NewFromInt(0),
				models.ArtskodeHelg:   decimal.NewFromFloat(7_783.78),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Calculate(tt.args.minutes, tt.args.timesheet, tt.args.payroll); (err != nil) != tt.wantErr {
				t.Errorf("Calculate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, tt.args.payroll.TypeCodes); diff != "" {
				t.Errorf("Calculate() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
