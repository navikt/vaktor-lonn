package compensation

import (
	"github.com/google/go-cmp/cmp"
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
		want models.Artskoder
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
					"2022-10-29": {
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
			want: models.Artskoder{
				Kveld: decimal.NewFromInt(40),
				Helg:  decimal.NewFromInt(110),
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
					"2022-10-31": {
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
			want: models.Artskoder{
				Dag: decimal.NewFromInt(35),
			},
		},
	}

	for _, tt := range tests {
		var payroll *models.Payroll
		payroll = &models.Payroll{
			ID:         uuid.UUID{},
			ApproverID: "Scathan",
		}

		t.Run(tt.name, func(t *testing.T) {
			CalculateCallOut(tt.args.timesheet, tt.args.satser, payroll)

			if diff := cmp.Diff(tt.want, payroll.Artskoder); diff != "" {
				t.Errorf("CalculateCallOut() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCalculate(t *testing.T) {
	type args struct {
		minutes map[string]models.GuardDuty
		satser  map[string]decimal.Decimal
		payroll *models.Payroll
	}
	tests := []struct {
		name string
		args args
		want models.Artskoder
	}{
		{
			name: "Utrykning i helg",
			args: args{
				payroll: &models.Payroll{},
				satser: map[string]decimal.Decimal{
					"lørsøn":  decimal.NewFromInt(65),
					"0620":    decimal.NewFromInt(15),
					"2006":    decimal.NewFromInt(25),
					"utvidet": decimal.NewFromInt(25),
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
			want: models.Artskoder{
				Morgen: decimal.NewFromInt(700),
				Dag:    decimal.NewFromInt(120),
				Kveld:  decimal.NewFromInt(300),
				Helg:   decimal.NewFromInt(624),
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
