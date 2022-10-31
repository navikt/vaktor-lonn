package compensation

import (
	"github.com/google/uuid"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/shopspring/decimal"
	"testing"
	"time"
)

func TestCalculateCallOut(t *testing.T) {
	type args struct {
		timesheet map[string]models.TimeSheet
		satser    map[string]decimal.Decimal
	}
	tests := []struct {
		name string
		args args
		want map[string]decimal.Decimal
	}{
		{
			name: "Utrykning i helg",
			args: args{
				satser: map[string]decimal.Decimal{
					"lørsøn":  decimal.NewFromInt(55),
					"0620":    decimal.NewFromInt(10),
					"2006":    decimal.NewFromInt(20),
					"utvidet": decimal.NewFromInt(15),
				},
				timesheet: map[string]models.TimeSheet{
					"2022-03-14": {
						Date:         time.Date(2022, 10, 29, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Lørdag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 10, 29, 20, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 10, 29, 22, 0, 0, 0, time.UTC),
								OtG: true,
							},
						},
					},
				},
			},
			want: map[string]decimal.Decimal{
				models.ArtskodeMorgen: {},
				models.ArtskodeDag:    {},
				models.ArtskodeKveld:  decimal.NewFromInt(40),
				models.ArtskodeHelg:   decimal.NewFromInt(110),
			},
		},

		{
			name: "Utrykning i utvidet arbeidstid",
			args: args{
				satser: map[string]decimal.Decimal{
					"lørsøn":  decimal.NewFromInt(55),
					"0620":    decimal.NewFromInt(10),
					"2006":    decimal.NewFromInt(20),
					"utvidet": decimal.NewFromInt(15),
				},
				timesheet: map[string]models.TimeSheet{
					"2022-03-14": {
						Date:         time.Date(2022, 10, 31, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 10, 31, 6, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 10, 31, 8, 0, 0, 0, time.UTC),
								OtG: true,
							},
						},
					},
				},
			},
			want: map[string]decimal.Decimal{
				models.ArtskodeMorgen: {},
				models.ArtskodeDag:    decimal.NewFromInt(35),
				models.ArtskodeKveld:  {},
				models.ArtskodeHelg:   {},
			},
		},
	}

	for _, tt := range tests {
		var payroll *models.Payroll
		payroll = &models.Payroll{
			ID:         uuid.UUID{},
			ApproverID: "Scathan",
			TypeCodes: map[string]decimal.Decimal{
				models.ArtskodeMorgen: {},
				models.ArtskodeDag:    {},
				models.ArtskodeKveld:  {},
				models.ArtskodeHelg:   {},
			},
		}

		t.Run(tt.name, func(t *testing.T) {
			CalculateCallOut(tt.args.timesheet, tt.args.satser, payroll)

			for code, value := range payroll.TypeCodes {
				if !value.Equal(tt.want[code]) {
					t.Errorf("CalculateCallOut(%v) got = %v, want %v", code, value, tt.want[code])
				}
			}
		})
	}
}