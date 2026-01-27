package calculator

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/navikt/vaktor-lonn/pkg/compensation"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/navikt/vaktor-lonn/pkg/overtime"
	"github.com/shopspring/decimal"
)

func Test_calculateMinutesWithGuardDutyInPeriod(t *testing.T) {
	type args struct {
		dutyPeriod models.Period
		compPeriod models.Period
		timesheet  []models.Clocking
	}

	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "Vanlig arbeidsdag",
			args: args{
				dutyPeriod: models.Period{
					Begin: time.Date(2022, 10, 3, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2022, 10, 4, 0, 0, 0, 0, time.UTC),
				},
				compPeriod: models.Period{
					Begin: time.Date(2022, 10, 3, 9, 0, 0, 0, time.UTC),
					End:   time.Date(2022, 10, 3, 14, 30, 0, 0, time.UTC),
				},
				timesheet: []models.Clocking{
					{
						In:  time.Date(2022, 10, 3, 8, 0, 0, 0, time.UTC),
						Out: time.Date(2022, 10, 3, 15, 0, 0, 0, time.UTC),
					},
				},
			},
			want: 0,
		},
		{
			name: "Uvanlig kort arbeidsdag",
			args: args{
				dutyPeriod: models.Period{
					Begin: time.Date(2022, 10, 3, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2022, 10, 4, 0, 0, 0, 0, time.UTC),
				},
				compPeriod: models.Period{
					Begin: time.Date(2022, 10, 3, 9, 0, 0, 0, time.UTC),
					End:   time.Date(2022, 10, 3, 14, 30, 0, 0, time.UTC),
				},
				timesheet: []models.Clocking{
					{
						In:  time.Date(2022, 10, 3, 10, 0, 0, 0, time.UTC),
						Out: time.Date(2022, 10, 3, 14, 0, 0, 0, time.UTC),
					},
				},
			},
			want: 90,
		},
		{
			name: "Forskjøvet arbeidsdag",
			args: args{
				dutyPeriod: models.Period{
					Begin: time.Date(2022, 10, 3, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2022, 10, 4, 0, 0, 0, 0, time.UTC),
				},
				compPeriod: models.Period{
					Begin: time.Date(2022, 10, 3, 9, 0, 0, 0, time.UTC),
					End:   time.Date(2022, 10, 3, 14, 30, 0, 0, time.UTC),
				},
				timesheet: []models.Clocking{
					{
						In:  time.Date(2022, 10, 3, 10, 0, 0, 0, time.UTC),
						Out: time.Date(2022, 10, 3, 18, 0, 0, 0, time.UTC),
					},
				},
			},
			want: 60,
		},
		{
			name: "Morgenvakt",
			args: args{
				dutyPeriod: models.Period{
					Begin: time.Date(2022, 10, 3, 6, 0, 0, 0, time.UTC),
					End:   time.Date(2022, 10, 3, 9, 0, 0, 0, time.UTC),
				},
				compPeriod: models.Period{
					Begin: time.Date(2022, 10, 3, 20, 0, 0, 0, time.UTC),
					End:   time.Date(2022, 10, 4, 0, 0, 0, 0, time.UTC),
				},
				timesheet: []models.Clocking{
					{
						In:  time.Date(2022, 10, 3, 10, 0, 0, 0, time.UTC),
						Out: time.Date(2022, 10, 3, 18, 0, 0, 0, time.UTC),
					},
				},
			},
			want: 0,
		},
		{
			name: "Døgnvakt med kveldsarbeid",
			args: args{
				dutyPeriod: models.Period{
					Begin: time.Date(2022, 10, 3, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2022, 10, 4, 0, 0, 0, 0, time.UTC),
				},
				compPeriod: models.Period{
					Begin: time.Date(2022, 10, 3, 20, 0, 0, 0, time.UTC),
					End:   time.Date(2022, 10, 4, 0, 0, 0, 0, time.UTC),
				},
				timesheet: []models.Clocking{
					{
						In:  time.Date(2022, 10, 3, 23, 10, 0, 0, time.UTC),
						Out: time.Date(2022, 10, 4, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			want: 190,
		},
		{
			name: "Døgnvakt med kveldsarbeid ved månedskifte",
			args: args{
				dutyPeriod: models.Period{
					Begin: time.Date(2022, 10, 31, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2022, 11, 1, 0, 0, 0, 0, time.UTC),
				},
				compPeriod: models.Period{
					Begin: time.Date(2022, 10, 31, 20, 0, 0, 0, time.UTC),
					End:   time.Date(2022, 11, 1, 0, 0, 0, 0, time.UTC),
				},
				timesheet: []models.Clocking{
					{
						In:  time.Date(2022, 10, 31, 23, 10, 0, 0, time.UTC),
						Out: time.Date(2022, 11, 1, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			want: 190,
		},
		{
			name: "Døgnvakt ved månedskifte",
			args: args{
				dutyPeriod: models.Period{
					Begin: time.Date(2022, 7, 31, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2022, 8, 1, 0, 0, 0, 0, time.UTC),
				},
				compPeriod: models.Period{
					Begin: time.Date(2022, 7, 31, 20, 0, 0, 0, time.UTC),
					End:   time.Date(2022, 8, 1, 0, 0, 0, 0, time.UTC),
				},
				timesheet: []models.Clocking{},
			},
			want: 240,
		},
		{
			name: "Døgnvakt ved årskifte",
			args: args{
				dutyPeriod: models.Period{
					Begin: time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				compPeriod: models.Period{
					Begin: time.Date(2022, 12, 31, 20, 0, 0, 0, time.UTC),
					End:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				timesheet: []models.Clocking{},
			},
			want: 240,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateMinutesWithGuardDutyInPeriod(tt.args.dutyPeriod, tt.args.compPeriod, tt.args.timesheet)
			if got != tt.want {
				t.Errorf("calculateMinutesWithGuardDutyInPeriod() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateDaylightSavingTimeModifier(t *testing.T) {
	type args struct {
		date    time.Time
		periods []models.Period
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "Stiller klokken tilbake (normaltid)",
			args: args{
				date: time.Date(2022, 10, 30, 10, 0, 0, 0, time.UTC),
				periods: []models.Period{
					{
						Begin: time.Date(2022, 10, 30, 0, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 10, 31, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			want: 60,
		},
		{
			name: "Har _ikke_ vakt når vi stiller klokken tilbake (normaltid)",
			args: args{
				date: time.Date(2022, 10, 30, 10, 0, 0, 0, time.UTC),
				periods: []models.Period{
					{
						Begin: time.Date(2022, 10, 30, 10, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 10, 30, 20, 0, 0, 0, time.UTC),
					},
				},
			},
			want: 0,
		},
		{
			name: "Stiller klokken fremover (sommertid)",
			args: args{
				date: time.Date(2022, 3, 27, 10, 0, 0, 0, time.UTC),
				periods: []models.Period{
					{
						Begin: time.Date(2022, 3, 27, 0, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 3, 28, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			want: -60,
		},
		{
			name: "Har _ikke_ vakt når vi stiller klokken fremover (sommertid)",
			args: args{
				date: time.Date(2022, 3, 27, 10, 0, 0, 0, time.UTC),
				periods: []models.Period{
					{
						Begin: time.Date(2022, 3, 27, 10, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 3, 27, 20, 0, 0, 0, time.UTC),
					},
				},
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateDaylightSavingTimeModifier(tt.args.periods, tt.args.date)
			if got != tt.want {
				t.Errorf("calculateDaylightSavingTimeModifier() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createKjernetid(t *testing.T) {
	type args struct {
		date     time.Time
		formName string
	}
	tests := []struct {
		name string
		args args
		want models.Period
	}{
		{
			name: "Vanlig kjernetid",
			args: args{
				date:     time.Date(2022, 11, 6, 0, 0, 0, 0, time.UTC),
				formName: "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
			},
			want: models.Period{
				Begin: time.Date(2022, 11, 6, 9, 0, 0, 0, time.UTC),
				End:   time.Date(2022, 11, 6, 14, 30, 0, 0, time.UTC),
			},
		},
		{
			name: "Kjernetid for julaften",
			args: args{
				date:     time.Date(2021, 12, 24, 0, 0, 0, 0, time.UTC),
				formName: "Julaften 0800-1200 *",
			},
			want: models.Period{
				Begin: time.Date(2021, 12, 24, 8, 0, 0, 0, time.UTC),
				End:   time.Date(2021, 12, 24, 12, 0, 0, 0, time.UTC),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createKjernetid(tt.args.date, tt.args.formName)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("createKjernetid() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_calculateGuardDutyInKjernetid(t *testing.T) {
	type args struct {
		currentDay models.TimeSheet
		period     models.Period
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "Vanlig arbeid i kjernetid",
			args: args{
				currentDay: models.TimeSheet{
					Date:         time.Date(2022, 11, 7, 0, 0, 0, 0, time.UTC),
					WorkingHours: 7.75,
					WorkingDay:   "Virkedag",
					FormName:     "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 11, 7, 8, 0, 0, 0, time.UTC),
							Out: time.Date(2022, 11, 7, 15, 45, 0, 0, time.UTC),
						},
					},
				},
				period: models.Period{
					Begin: time.Date(2022, 11, 7, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2022, 11, 8, 0, 0, 0, 0, time.UTC),
				},
			},
			want: 0,
		},
		{
			name: "Ingen arbeid i kjernetid",
			args: args{
				currentDay: models.TimeSheet{
					Date:         time.Date(2022, 11, 7, 0, 0, 0, 0, time.UTC),
					WorkingHours: 7.75,
					WorkingDay:   "Virkedag",
					FormName:     "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 11, 7, 15, 0, 0, 0, time.UTC),
							Out: time.Date(2022, 11, 7, 20, 45, 0, 0, time.UTC),
						},
					},
				},
				period: models.Period{
					Begin: time.Date(2022, 11, 7, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2022, 11, 8, 0, 0, 0, 0, time.UTC),
				},
			},
			want: 330,
		},
		{
			name: "Kom sent på jobb",
			args: args{
				currentDay: models.TimeSheet{
					Date:         time.Date(2022, 11, 7, 0, 0, 0, 0, time.UTC),
					WorkingHours: 7.75,
					WorkingDay:   "Virkedag",
					FormName:     "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 11, 7, 10, 0, 0, 0, time.UTC),
							Out: time.Date(2022, 11, 7, 17, 45, 0, 0, time.UTC),
						},
					},
				},
				period: models.Period{
					Begin: time.Date(2022, 11, 7, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2022, 11, 8, 0, 0, 0, 0, time.UTC),
				},
			},
			want: 60,
		},

		{
			name: "Julaften på en lørdag",
			args: args{
				currentDay: models.TimeSheet{
					Date:       time.Date(2022, 12, 24, 0, 0, 0, 0, time.UTC),
					WorkingDay: "Lørdag",
					FormName:   "BV Lørdag IKT",
					Clockings:  []models.Clocking{},
				},
				period: models.Period{
					Begin: time.Date(2022, 12, 24, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2022, 12, 25, 0, 0, 0, 0, time.UTC),
				},
			},
			want: 0,
		},

		{
			name: "Julaften på en fredag (litt sen klokking)",
			args: args{
				currentDay: models.TimeSheet{
					Date:         time.Date(2021, 12, 24, 0, 0, 0, 0, time.UTC),
					WorkingHours: 4,
					WorkingDay:   "Virkedag",
					FormName:     "Julaften 0800-1200 *",
					Clockings: []models.Clocking{
						{
							In:  time.Date(2021, 12, 24, 9, 0, 0, 0, time.UTC),
							Out: time.Date(2021, 12, 24, 14, 0, 0, 0, time.UTC),
						},
					},
				},
				period: models.Period{
					Begin: time.Date(2021, 12, 24, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2021, 12, 25, 0, 0, 0, 0, time.UTC),
				},
			},
			want: 60,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateGuardDutyInKjernetid(tt.args.currentDay, tt.args.period); got != tt.want {
				t.Errorf("calculateGuardDutyInKjernetid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateMinutesToBePaid(t *testing.T) {
	type args struct {
		schedule  map[string][]models.Period
		timesheet map[string]models.TimeSheet
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]models.GuardDuty
		wantErr bool
	}{
		{
			name: "Julaften på en lørdag",
			args: args{
				schedule: map[string][]models.Period{
					"2022-12-24": {
						{
							Begin: time.Date(2022, 12, 24, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 12, 25, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-12-25": {
						{
							Begin: time.Date(2022, 12, 25, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 12, 26, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-12-26": {
						{
							Begin: time.Date(2022, 12, 26, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 12, 27, 0, 0, 0, 0, time.UTC),
						},
					},
				},
				timesheet: map[string]models.TimeSheet{
					"2022-12-24": {
						Date:         time.Date(2022, 12, 24, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Lørdag",
						FormName:     "BV Lørdag IKT",
						Salary:       decimal.NewFromInt(500_000),
						Clockings:    nil,
					},
					"2022-12-25": {
						Date:         time.Date(2022, 12, 25, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "1. Juledag",
						FormName:     "Helligdag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings:    nil,
					},
					"2022-12-26": {
						Date:         time.Date(2022, 12, 26, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "2. Juledag",
						FormName:     "Helligdag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings:    nil,
					},
				},
			},
			want: map[string]models.GuardDuty{
				"2022-12-24": {
					Hvilende2000:        240,
					Hvilende0006:        360,
					Hvilende0620:        840,
					Helgetillegg:        1440,
					Skifttillegg:        0,
					WeekendCompensation: true,
				},
				"2022-12-25": {
					Hvilende2000:        240,
					Hvilende0006:        360,
					Hvilende0620:        840,
					Helgetillegg:        1440,
					Skifttillegg:        0,
					WeekendCompensation: true,
				},
				"2022-12-26": {
					Hvilende2000:        240,
					Hvilende0006:        360,
					Helligdag0620:       840,
					Helgetillegg:        0,
					Skifttillegg:        240,
					WeekendCompensation: false,
				},
			},
		},

		{
			name: "Nyttårsaften på en lørdag",
			args: args{
				schedule: map[string][]models.Period{
					"2022-12-31": {
						{
							Begin: time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
						},
					},
				},
				timesheet: map[string]models.TimeSheet{
					"2022-12-31": {
						Date:         time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Lørdag",
						FormName:     "BV Lørdag IKT",
						Salary:       decimal.NewFromInt(500_000),
						Clockings:    nil,
					},
				},
			},
			want: map[string]models.GuardDuty{
				"2022-12-31": {
					Hvilende2000:        240,
					Hvilende0006:        360,
					Hvilende0620:        840,
					Helgetillegg:        1440,
					Skifttillegg:        0,
					WeekendCompensation: true,
				},
			},
		},

		{
			name: "Julaften på en fredag",
			args: args{
				schedule: map[string][]models.Period{
					"2021-12-24": {
						{
							Begin: time.Date(2021, 12, 24, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2021, 12, 25, 0, 0, 0, 0, time.UTC),
						},
					},
					"2021-12-25": {
						{
							Begin: time.Date(2021, 12, 25, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2021, 12, 26, 0, 0, 0, 0, time.UTC),
						},
					},
					"2021-12-26": {
						{
							Begin: time.Date(2021, 12, 26, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2021, 12, 27, 0, 0, 0, 0, time.UTC),
						},
					},
				},
				timesheet: map[string]models.TimeSheet{
					"2021-12-24": {
						Date:         time.Date(2021, 12, 24, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Virkedag",
						FormName:     "Julaften 0800-1200 *",
						Salary:       decimal.NewFromInt(500_000),
						Clockings:    nil,
					},
					"2021-12-25": {
						Date:         time.Date(2021, 12, 25, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "1. Juledag",
						FormName:     "Helligdag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings:    nil,
					},
					"2021-12-26": {
						Date:         time.Date(2021, 12, 26, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "2. Juledag",
						FormName:     "Helligdag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings:    nil,
					},
				},
			},
			want: map[string]models.GuardDuty{
				"2021-12-24": {
					Hvilende2000:  240,
					Hvilende0006:  360,
					Helligdag0620: 480,
					Hvilende0620:  120,
					Helgetillegg:  0,
					Skifttillegg:  240,
				},
				"2021-12-25": {
					Hvilende2000:        240,
					Hvilende0006:        360,
					Hvilende0620:        840,
					Helgetillegg:        1440,
					Skifttillegg:        0,
					WeekendCompensation: true,
				},
				"2021-12-26": {
					Hvilende2000:        240,
					Hvilende0006:        360,
					Hvilende0620:        840,
					Helgetillegg:        1440,
					Skifttillegg:        0,
					WeekendCompensation: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateMinutesToBePaid(tt.args.schedule, tt.args.timesheet)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateMinutesToBePaid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("calculateMinutesToBePaid() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestHoursGuarddutySalary tester kun summering av timer. For tester av lønn se filen guard_duty_salary_test.go
func TestHoursGuarddutySalary(t *testing.T) {
	type args struct {
		timesheet   map[string]models.TimeSheet
		guardPeriod map[string][]models.Period
	}
	tests := []struct {
		name    string
		args    args
		want    models.Artskoder
		wantErr bool
	}{
		{
			name: "En døgnkontinuerlig vaktuke (vintertid)",
			args: args{
				timesheet: map[string]models.TimeSheet{
					"2022-10-12": {
						Date:         time.Date(2022, 10, 12, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						FormName:     "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Salary:       decimal.NewFromInt(750_000),
						Koststed:     "000000",
						Formal:       "000000",
						Aktivitet:    "000000",
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 10, 12, 8, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 10, 12, 15, 45, 0, 0, time.UTC),
							},
						},
					},
					"2022-10-13": {
						Date:         time.Date(2022, 10, 13, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						FormName:     "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Salary:       decimal.NewFromInt(750_000),
						Koststed:     "000000",
						Formal:       "000000",
						Aktivitet:    "000000",
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 10, 13, 8, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 10, 13, 15, 45, 0, 0, time.UTC),
							},
						},
					},
					"2022-10-14": {
						Date:         time.Date(2022, 10, 14, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						FormName:     "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Salary:       decimal.NewFromInt(750_000),
						Koststed:     "000000",
						Formal:       "000000",
						Aktivitet:    "000000",
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 10, 14, 8, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 10, 14, 15, 45, 0, 0, time.UTC),
							},
						},
					},
					"2022-10-15": {
						Date:         time.Date(2022, 10, 15, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Lørdag",
						FormName:     "BV Lørdag IKT",
						Salary:       decimal.NewFromInt(750_000),
						Koststed:     "000000",
						Formal:       "000000",
						Aktivitet:    "000000",
						Clockings:    []models.Clocking{},
					},
					"2022-10-16": {
						Date:         time.Date(2022, 10, 16, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Søndag",
						FormName:     "BV Søndag IKT",
						Salary:       decimal.NewFromInt(750_000),
						Koststed:     "000000",
						Formal:       "000000",
						Aktivitet:    "000000",
						Clockings:    []models.Clocking{},
					},
					"2022-10-17": {
						Date:         time.Date(2022, 10, 17, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						FormName:     "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Salary:       decimal.NewFromInt(750_000),
						Koststed:     "000000",
						Formal:       "000000",
						Aktivitet:    "000000",
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 10, 17, 8, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 10, 17, 15, 45, 0, 0, time.UTC),
							},
						},
					},
					"2022-10-18": {
						Date:         time.Date(2022, 10, 18, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						FormName:     "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Salary:       decimal.NewFromInt(750_000),
						Koststed:     "000000",
						Formal:       "000000",
						Aktivitet:    "000000",
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 10, 18, 8, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 10, 18, 15, 45, 0, 0, time.UTC),
							},
						},
					},
					"2022-10-19": {
						Date:         time.Date(2022, 10, 19, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						FormName:     "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Salary:       decimal.NewFromInt(750_000),
						Koststed:     "000000",
						Formal:       "000000",
						Aktivitet:    "000000",
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 10, 19, 8, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 10, 19, 15, 45, 0, 0, time.UTC),
							},
						},
					},
				},
				guardPeriod: map[string][]models.Period{
					"2022-10-12": {
						{
							Begin: time.Date(2022, 10, 12, 12, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 13, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-10-13": {
						{
							Begin: time.Date(2022, 10, 13, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 14, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-10-14": {
						{
							Begin: time.Date(2022, 10, 14, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 15, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-10-15": {
						{
							Begin: time.Date(2022, 10, 15, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 16, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-10-16": {
						{
							Begin: time.Date(2022, 10, 16, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 17, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-10-17": {
						{
							Begin: time.Date(2022, 10, 17, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 18, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-10-18": {
						{
							Begin: time.Date(2022, 10, 18, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 19, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-10-19": {
						{
							Begin: time.Date(2022, 10, 19, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 19, 12, 0, 0, 0, time.UTC),
						},
					},
				},
			},
			want: models.Artskoder{
				Morgen: models.Artskode{
					Sum:   decimal.NewFromFloat(5614.86),
					Hours: 30,
				},
				Kveld: models.Artskode{
					Sum:   decimal.NewFromFloat(3743.24),
					Hours: 20,
				},
				Dag: models.Artskode{
					Sum:   decimal.NewFromFloat(4235.27),
					Hours: 31,
				},
				Helg: models.Artskode{
					Sum:   decimal.NewFromFloat(9327.78),
					Hours: 48,
				},
				Skift: models.Artskode{
					Sum:   decimal.NewFromFloat(100),
					Hours: 20,
				},
			},
		},

		{
			name: "En døgnkontinuerlig vaktuke (normaltid)",
			args: args{
				timesheet: map[string]models.TimeSheet{
					"2022-10-12": {
						Date:         time.Date(2022, 10, 12, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.5,
						WorkingDay:   "Virkedag",
						FormName:     "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Salary:       decimal.NewFromInt(750_000),
						Koststed:     "000000",
						Formal:       "000000",
						Aktivitet:    "000000",
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 10, 12, 8, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 10, 12, 15, 30, 0, 0, time.UTC),
							},
						},
					},
					"2022-10-13": {
						Date:         time.Date(2022, 10, 13, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.5,
						WorkingDay:   "Virkedag",
						FormName:     "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Salary:       decimal.NewFromInt(750_000),
						Koststed:     "000000",
						Formal:       "000000",
						Aktivitet:    "000000",
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 10, 13, 8, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 10, 13, 15, 30, 0, 0, time.UTC),
							},
						},
					},
					"2022-10-14": {
						Date:         time.Date(2022, 10, 14, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.5,
						WorkingDay:   "Virkedag",
						FormName:     "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Salary:       decimal.NewFromInt(750_000),
						Koststed:     "000000",
						Formal:       "000000",
						Aktivitet:    "000000",
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 10, 14, 8, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 10, 14, 15, 30, 0, 0, time.UTC),
							},
						},
					},
					"2022-10-15": {
						Date:         time.Date(2022, 10, 15, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Lørdag",
						FormName:     "BV Lørdag IKT",
						Salary:       decimal.NewFromInt(750_000),
						Koststed:     "000000",
						Formal:       "000000",
						Aktivitet:    "000000",
						Clockings:    []models.Clocking{},
					},
					"2022-10-16": {
						Date:         time.Date(2022, 10, 16, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Søndag",
						FormName:     "BV Søndag IKT",
						Salary:       decimal.NewFromInt(750_000),
						Koststed:     "000000",
						Formal:       "000000",
						Aktivitet:    "000000",
						Clockings:    []models.Clocking{},
					},
					"2022-10-17": {
						Date:         time.Date(2022, 10, 17, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.5,
						WorkingDay:   "Virkedag",
						FormName:     "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Salary:       decimal.NewFromInt(750_000),
						Koststed:     "000000",
						Formal:       "000000",
						Aktivitet:    "000000",
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 10, 17, 8, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 10, 17, 15, 30, 0, 0, time.UTC),
							},
						},
					},
					"2022-10-18": {
						Date:         time.Date(2022, 10, 18, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.5,
						WorkingDay:   "Virkedag",
						FormName:     "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Salary:       decimal.NewFromInt(750_000),
						Koststed:     "000000",
						Formal:       "000000",
						Aktivitet:    "000000",
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 10, 18, 8, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 10, 18, 15, 30, 0, 0, time.UTC),
							},
						},
					},
					"2022-10-19": {
						Date:         time.Date(2022, 10, 19, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.5,
						WorkingDay:   "Virkedag",
						FormName:     "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Salary:       decimal.NewFromInt(750_000),
						Koststed:     "000000",
						Formal:       "000000",
						Aktivitet:    "000000",
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 10, 19, 8, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 10, 19, 15, 30, 0, 0, time.UTC),
							},
						},
					},
				},
				guardPeriod: map[string][]models.Period{
					"2022-10-12": {
						{
							Begin: time.Date(2022, 10, 12, 12, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 13, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-10-13": {
						{
							Begin: time.Date(2022, 10, 13, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 14, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-10-14": {
						{
							Begin: time.Date(2022, 10, 14, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 15, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-10-15": {
						{
							Begin: time.Date(2022, 10, 15, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 16, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-10-16": {
						{
							Begin: time.Date(2022, 10, 16, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 17, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-10-17": {
						{
							Begin: time.Date(2022, 10, 17, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 18, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-10-18": {
						{
							Begin: time.Date(2022, 10, 18, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 19, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-10-19": {
						{
							Begin: time.Date(2022, 10, 19, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 19, 12, 0, 0, 0, time.UTC),
						},
					},
				},
			},
			want: models.Artskoder{
				Morgen: models.Artskode{
					Sum:   decimal.NewFromFloat(5614.86),
					Hours: 30,
				},
				Kveld: models.Artskode{
					Sum:   decimal.NewFromFloat(3743.24),
					Hours: 20,
				},
				Dag: models.Artskode{
					Sum:   decimal.NewFromFloat(4508.51),
					Hours: 33,
				},
				Helg: models.Artskode{
					Sum:   decimal.NewFromFloat(9327.78),
					Hours: 48,
				},
				Skift: models.Artskode{
					Sum:   decimal.NewFromFloat(100),
					Hours: 20,
				},
			},
		},
		{
			name: "Vakt en lørdag med utrykning",
			args: args{
				timesheet: map[string]models.TimeSheet{
					"2026-01-10": {
						Date:         time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.5,
						WorkingDay:   "Lørdag",
						FormName:     "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Salary:       decimal.NewFromInt(500_000),
						Koststed:     "000000",
						Formal:       "000000",
						Aktivitet:    "000000",
						Clockings: []models.Clocking{
							{
								In:  time.Date(2026, 1, 10, 20, 0, 0, 0, time.UTC),
								Out: time.Date(2026, 1, 10, 22, 0, 0, 0, time.UTC),
								OtG: true,
							},
						},
					},
				},
				guardPeriod: map[string][]models.Period{
					"2026-01-10": {
						{
							Begin: time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC),
						},
					},
				},
			},
			want: models.Artskoder{
				Helg: models.Artskode{
					Sum:   decimal.NewFromFloat(3366.59),
					Hours: 24,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vaktplan := models.Vaktplan{
				ID:       uuid.UUID{},
				Ident:    "A123456",
				Schedule: tt.args.guardPeriod,
			}

			minWinTid := models.MinWinTid{
				Ident:        "A123456",
				ResourceID:   "123456",
				ApproverID:   "M654321",
				ApproverName: "Kalpana, Bran",
				Timesheet:    tt.args.timesheet,
				Satser: models.Satser{
					Helg:    decimal.NewFromInt(65),
					Dag:     decimal.NewFromInt(15),
					Natt:    decimal.NewFromInt(25),
					Utvidet: decimal.NewFromInt(25),
				},
			}

			got, err := GuarddutySalary(vaktplan, minWinTid)
			if (err != nil) != tt.wantErr {
				t.Errorf("GuarddutySalary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, got.Artskoder); diff != "" {
				t.Errorf("GuarddutySalary() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCalculateIsEqualForHolidayOnSaturdayAndRegularSaturday(t *testing.T) {
	salary := decimal.NewFromInt(500_000)
	satser := models.Satser{
		Helg:    decimal.NewFromInt(65),
		Dag:     decimal.NewFromInt(15),
		Natt:    decimal.NewFromInt(25),
		Utvidet: decimal.NewFromInt(25),
	}

	holidayPayroll := &models.Payroll{}
	holidayMinutes := map[string]models.GuardDuty{
		"2022-10-15": {
			Hvilende2000:        240,
			Hvilende0006:        360,
			Hvilende0620:        840,
			Helgetillegg:        1440,
			Skifttillegg:        0,
			WeekendCompensation: true,
		},
	}
	compensation.Calculate(holidayMinutes, satser, holidayPayroll)
	overtime.Calculate(holidayMinutes, salary, holidayPayroll)

	regularPayroll := &models.Payroll{}
	regularMinutes := map[string]models.GuardDuty{
		"2022-10-15": {
			Hvilende2000:        240,
			Hvilende0006:        360,
			Hvilende0620:        840,
			Helgetillegg:        1440,
			Skifttillegg:        0,
			WeekendCompensation: true,
		},
	}
	compensation.Calculate(regularMinutes, satser, regularPayroll)
	overtime.Calculate(regularMinutes, salary, regularPayroll)

	if diff := cmp.Diff(holidayPayroll.Artskoder, regularPayroll.Artskoder); diff != "" {
		t.Errorf("Calculate() holiday on saturday did not return the same as a normal saturday")
		t.Errorf("Calculate() mismatch (-want +got):\n%s", diff)
	}
}

// TestCalculateIsNotEqualForHolidayOnMondayAndRegularMonday tester at man ikke får den samme lønnen for en vanlig arbeidsdag
// uten arbeid og en arbeidsdag som er en helligdag.
func TestCalculateIsNotEqualForHolidayOnMondayAndRegularMonday(t *testing.T) {
	salary := decimal.NewFromInt(500_000)
	satser := models.Satser{
		Helg:    decimal.NewFromInt(65),
		Dag:     decimal.NewFromInt(15),
		Natt:    decimal.NewFromInt(25),
		Utvidet: decimal.NewFromInt(25),
	}

	holidayPayroll := &models.Payroll{}
	holidayMinutes := map[string]models.GuardDuty{
		"2022-10-17": {
			Hvilende2000:  240,
			Hvilende0006:  360,
			Helligdag0620: 840,
			Helgetillegg:  0,
			Skifttillegg:  240,
		},
	}
	compensation.Calculate(holidayMinutes, satser, holidayPayroll)
	overtime.Calculate(holidayMinutes, salary, holidayPayroll)

	regularPayroll := &models.Payroll{}
	regularMinutes := map[string]models.GuardDuty{
		"2022-10-17": {
			Hvilende2000: 240,
			Hvilende0006: 360,
			Hvilende0620: 840, // this is not a legal work day with guard duty
			Helgetillegg: 0,
			Skifttillegg: 240,
		},
	}
	compensation.Calculate(regularMinutes, satser, regularPayroll)
	overtime.Calculate(regularMinutes, salary, regularPayroll)

	if diff := cmp.Diff(holidayPayroll.Artskoder, regularPayroll.Artskoder); diff == "" {
		t.Errorf("Calculate() holiday on monday returns the same as a normal monday")
		t.Errorf("Calculate() mismatch (-want +got):\n%s", diff)
	}
}
