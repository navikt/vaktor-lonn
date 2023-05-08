package callout

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/shopspring/decimal"
	"testing"
	"time"
)

func TestCalculate(t *testing.T) {
	type args struct {
		timesheet map[string]models.TimeSheet
		satser    models.Satser
	}
	tests := []struct {
		name string
		args args
		want models.Artskoder
	}{
		{
			name: "Utrykning i helg",
			args: args{
				satser: models.Satser{
					Helg:    decimal.NewFromInt(65),
					Dag:     decimal.NewFromInt(15),
					Natt:    decimal.NewFromInt(25),
					Utvidet: decimal.NewFromInt(25),
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
				Utrykning: models.Artskode{
					Sum:   decimal.NewFromInt(130),
					Hours: 2,
				},
			},
		},

		{
			name: "Korte utrykninger på lørdag",
			args: args{
				satser: models.Satser{
					Helg:    decimal.NewFromInt(65),
					Dag:     decimal.NewFromInt(15),
					Natt:    decimal.NewFromInt(25),
					Utvidet: decimal.NewFromInt(25),
				},
				timesheet: map[string]models.TimeSheet{
					"2022-10-29": {
						Date:         time.Date(2022, 10, 29, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Lørdag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 10, 29, 10, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 10, 29, 10, 20, 0, 0, time.UTC),
								OtG: true,
							},
							{
								In:  time.Date(2022, 10, 29, 20, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 10, 29, 20, 40, 0, 0, time.UTC),
								OtG: true,
							},
						},
					},
				},
			},
			want: models.Artskoder{
				Utrykning: models.Artskode{
					Sum:   decimal.NewFromInt(65),
					Hours: 1,
				},
			},
		},

		{
			name: "Korte utrykninger over flere dager",
			args: args{
				satser: models.Satser{
					Helg:    decimal.NewFromInt(65),
					Dag:     decimal.NewFromInt(15),
					Natt:    decimal.NewFromInt(25),
					Utvidet: decimal.NewFromInt(25),
				},
				timesheet: map[string]models.TimeSheet{
					"2022-10-26": {
						Date:         time.Date(2022, 10, 26, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 10, 26, 20, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 10, 26, 20, 20, 0, 0, time.UTC),
								OtG: true,
							},
						},
					},
					"2022-10-27": {
						Date:         time.Date(2022, 10, 27, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 10, 27, 20, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 10, 27, 20, 20, 0, 0, time.UTC),
								OtG: true,
							},
						},
					},
					"2022-10-28": {
						Date:         time.Date(2022, 10, 28, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 10, 28, 5, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 10, 28, 5, 20, 0, 0, time.UTC),
								OtG: true,
							},
						},
					},
				},
			},
			want: models.Artskoder{
				Utrykning: models.Artskode{
					Sum:   decimal.NewFromInt(0),
					Hours: 0,
				},
			},
		},

		{
			name: "Utrykning i utvidet arbeidstid",
			args: args{
				satser: models.Satser{
					Helg:    decimal.NewFromInt(65),
					Dag:     decimal.NewFromInt(15),
					Natt:    decimal.NewFromInt(25),
					Utvidet: decimal.NewFromInt(25),
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
				Utrykning: models.Artskode{
					Sum:   decimal.NewFromInt(0),
					Hours: 0,
				},
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
			Calculate(tt.args.timesheet, tt.args.satser, payroll)

			if diff := cmp.Diff(tt.want, payroll.Artskoder); diff != "" {
				t.Errorf("Calculate() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
