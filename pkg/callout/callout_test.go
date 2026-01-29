package callout

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/shopspring/decimal"
)

func TestCalculate(t *testing.T) {
	type args struct {
		satser    models.Satser
		schedule  map[string][]models.Period
		timesheet map[string]models.TimeSheet
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
				schedule: map[string][]models.Period{
					"2022-10-29": {
						{
							Begin: time.Date(2022, 10, 29, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 30, 0, 0, 0, 0, time.UTC),
						},
					},
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
			name: "Flere korte utrykninger på en lørdag",
			args: args{
				satser: models.Satser{
					Helg:    decimal.NewFromInt(65),
					Dag:     decimal.NewFromInt(15),
					Natt:    decimal.NewFromInt(25),
					Utvidet: decimal.NewFromInt(25),
				},
				schedule: map[string][]models.Period{
					"2022-10-29": {
						{
							Begin: time.Date(2022, 10, 29, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 30, 0, 0, 0, 0, time.UTC),
						},
					},
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
								Out: time.Date(2022, 10, 29, 20, 20, 0, 0, time.UTC),
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
			name: "For kort utrykninger på en lørdag blir rundet ned",
			args: args{
				satser: models.Satser{
					Helg:    decimal.NewFromInt(65),
					Dag:     decimal.NewFromInt(15),
					Natt:    decimal.NewFromInt(25),
					Utvidet: decimal.NewFromInt(25),
				},
				schedule: map[string][]models.Period{
					"2022-10-29": {
						{
							Begin: time.Date(2022, 10, 29, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 30, 0, 0, 0, 0, time.UTC),
						},
					},
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
			name: "Utrykning i ukedag på kvelden gir ingen kompensasjon",
			args: args{
				satser: models.Satser{
					Helg:    decimal.NewFromInt(65),
					Dag:     decimal.NewFromInt(15),
					Natt:    decimal.NewFromInt(25),
					Utvidet: decimal.NewFromInt(25),
				},
				schedule: map[string][]models.Period{
					"2022-10-26": {
						{
							Begin: time.Date(2022, 10, 26, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 27, 0, 0, 0, 0, time.UTC),
						},
					},
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
								Out: time.Date(2022, 10, 26, 22, 0, 0, 0, time.UTC),
								OtG: true,
							},
						},
					},
				},
			},
			want: models.Artskoder{},
		},

		{
			name: "Utrykning i ukedag under utvidet arbeidstid gir kompensasjon",
			args: args{
				satser: models.Satser{
					Helg:    decimal.NewFromInt(65),
					Dag:     decimal.NewFromInt(15),
					Natt:    decimal.NewFromInt(25),
					Utvidet: decimal.NewFromInt(25),
				},
				schedule: map[string][]models.Period{
					"2022-10-26": {
						{
							Begin: time.Date(2022, 10, 26, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 27, 0, 0, 0, 0, time.UTC),
						},
					},
				},
				timesheet: map[string]models.TimeSheet{
					"2022-10-26": {
						Date:         time.Date(2022, 10, 26, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 10, 26, 6, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 10, 26, 7, 0, 0, 0, time.UTC),
								OtG: true,
							},
							{
								In:  time.Date(2022, 10, 26, 17, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 10, 26, 20, 0, 0, 0, time.UTC),
								OtG: true,
							},
						},
					},
				},
			},
			want: models.Artskoder{
				Utrykning: models.Artskode{
					Sum:   decimal.NewFromInt(100),
					Hours: 4,
				},
			},
		},
	}

	for _, tt := range tests {
		payroll := &models.Payroll{
			ID:         uuid.UUID{},
			ApproverID: "Scathan",
		}

		t.Run(tt.name, func(t *testing.T) {
			Calculate(tt.args.schedule, tt.args.timesheet, tt.args.satser, payroll)

			if diff := cmp.Diff(tt.want, payroll.Artskoder); diff != "" {
				t.Errorf("Calculate() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
