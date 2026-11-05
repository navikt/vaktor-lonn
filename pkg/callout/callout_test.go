package callout

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/navikt/vaktor-lonn/pkg/testhelpers"
)

func TestCalculate(t *testing.T) {
	type args struct {
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
				schedule: map[string][]models.Period{
					"2022-10-29": {testhelpers.Period("2022-10-29T00:00", "2022-10-30T00:00")},
				},
				timesheet: map[string]models.TimeSheet{
					"2022-10-29": {
						Date:         testhelpers.D("2022-10-29"),
						WorkingHours: 0,
						WorkingDay:   "Lørdag",
						Salary:       testhelpers.Salary500k,
						Clockings: []models.Clocking{
							testhelpers.ClockingOtG("2022-10-29T20:00", "2022-10-29T22:00"),
						},
					},
				},
			},
			want: models.Artskoder{
				Utrykning: models.Artskode{Sum: testhelpers.DecI(130), Hours: 2},
			},
		},

		{
			name: "Flere korte utrykninger på en lørdag",
			args: args{
				schedule: map[string][]models.Period{
					"2022-10-29": {testhelpers.Period("2022-10-29T00:00", "2022-10-30T00:00")},
				},
				timesheet: map[string]models.TimeSheet{
					"2022-10-29": {
						Date:         testhelpers.D("2022-10-29"),
						WorkingHours: 0,
						WorkingDay:   "Lørdag",
						Salary:       testhelpers.Salary500k,
						Clockings: []models.Clocking{
							testhelpers.ClockingOtG("2022-10-29T10:00", "2022-10-29T10:20"),
							testhelpers.ClockingOtG("2022-10-29T20:00", "2022-10-29T20:20"),
						},
					},
				},
			},
			want: models.Artskoder{
				Utrykning: models.Artskode{Sum: testhelpers.DecI(65), Hours: 1},
			},
		},

		{
			name: "For kort utrykninger på en lørdag blir rundet ned",
			args: args{
				schedule: map[string][]models.Period{
					"2022-10-29": {testhelpers.Period("2022-10-29T00:00", "2022-10-30T00:00")},
				},
				timesheet: map[string]models.TimeSheet{
					"2022-10-29": {
						Date:         testhelpers.D("2022-10-29"),
						WorkingHours: 0,
						WorkingDay:   "Lørdag",
						Salary:       testhelpers.Salary500k,
						Clockings: []models.Clocking{
							testhelpers.ClockingOtG("2022-10-29T10:00", "2022-10-29T10:20"),
						},
					},
				},
			},
			want: models.Artskoder{
				Utrykning: models.Artskode{Sum: testhelpers.DecI(0), Hours: 0},
			},
		},

		{
			name: "Utrykning i ukedag på kvelden gir ingen kompensasjon",
			args: args{
				schedule: map[string][]models.Period{
					"2022-10-26": {testhelpers.Period("2022-10-26T00:00", "2022-10-27T00:00")},
				},
				timesheet: map[string]models.TimeSheet{
					"2022-10-26": {
						Date:         testhelpers.D("2022-10-26"),
						WorkingHours: 0,
						WorkingDay:   "Virkedag",
						Salary:       testhelpers.Salary500k,
						Clockings: []models.Clocking{
							testhelpers.ClockingOtG("2022-10-26T20:00", "2022-10-26T22:00"),
						},
					},
				},
			},
			want: models.Artskoder{},
		},

		{
			name: "Utrykning i ukedag under utvidet arbeidstid gir kompensasjon",
			args: args{
				schedule: map[string][]models.Period{
					"2022-10-26": {testhelpers.Period("2022-10-26T00:00", "2022-10-27T00:00")},
				},
				timesheet: map[string]models.TimeSheet{
					"2022-10-26": {
						Date:         testhelpers.D("2022-10-26"),
						WorkingHours: 0,
						WorkingDay:   "Virkedag",
						Salary:       testhelpers.Salary500k,
						Clockings: []models.Clocking{
							testhelpers.ClockingOtG("2022-10-26T06:00", "2022-10-26T07:00"),
							testhelpers.ClockingOtG("2022-10-26T17:00", "2022-10-26T20:00"),
						},
					},
				},
			},
			want: models.Artskoder{
				Utrykning: models.Artskode{Sum: testhelpers.DecI(100), Hours: 4},
			},
		},

		{
			name: "Vakt hverdag utenfor utvidet arbeidstid",
			args: args{
				schedule: map[string][]models.Period{
					"2026-01-01": {testhelpers.Period("2026-01-01T10:00", "2026-01-01T15:00")},
				},
				timesheet: map[string]models.TimeSheet{
					"2026-01-01": {
						Date:         testhelpers.D("2026-01-01"),
						WorkingHours: 0,
						WorkingDay:   "Helligdag",
						Salary:       testhelpers.Salary500k,
						Clockings: []models.Clocking{
							testhelpers.ClockingOtG("2026-10-26T10:00", "2026-10-26T11:00"),
						},
					},
				},
			},
			want: models.Artskoder{},
		},
	}

	for _, tt := range tests {
		payroll := &models.Payroll{
			ID:         uuid.UUID{},
			ApproverID: "Scathan",
		}

		t.Run(tt.name, func(t *testing.T) {
			Calculate(tt.args.schedule, tt.args.timesheet, testhelpers.DefaultSatser, payroll)

			if diff := cmp.Diff(tt.want, payroll.Artskoder); diff != "" {
				t.Errorf("Calculate() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
