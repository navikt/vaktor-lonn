package calculator

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/navikt/vaktor-lonn/pkg/kronetillegg"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/navikt/vaktor-lonn/pkg/overtime"
	"github.com/navikt/vaktor-lonn/pkg/testhelpers"
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
				dutyPeriod: testhelpers.Period("2022-10-03T00:00", "2022-10-04T00:00"),
				compPeriod: testhelpers.Period("2022-10-03T09:00", "2022-10-03T14:30"),
				timesheet: []models.Clocking{
					testhelpers.Clocking("2022-10-03T08:00", "2022-10-03T15:00"),
				},
			},
			want: 0,
		},
		{
			name: "Uvanlig kort arbeidsdag",
			args: args{
				dutyPeriod: testhelpers.Period("2022-10-03T00:00", "2022-10-04T00:00"),
				compPeriod: testhelpers.Period("2022-10-03T09:00", "2022-10-03T14:30"),
				timesheet: []models.Clocking{
					testhelpers.Clocking("2022-10-03T10:00", "2022-10-03T14:00"),
				},
			},
			want: 90,
		},
		{
			name: "Forskjøvet arbeidsdag",
			args: args{
				dutyPeriod: testhelpers.Period("2022-10-03T00:00", "2022-10-04T00:00"),
				compPeriod: testhelpers.Period("2022-10-03T09:00", "2022-10-03T14:30"),
				timesheet: []models.Clocking{
					testhelpers.Clocking("2022-10-03T10:00", "2022-10-03T18:00"),
				},
			},
			want: 60,
		},
		{
			name: "Morgenvakt",
			args: args{
				dutyPeriod: testhelpers.Period("2022-10-03T06:00", "2022-10-03T09:00"),
				compPeriod: testhelpers.Period("2022-10-03T20:00", "2022-10-04T00:00"),
				timesheet: []models.Clocking{
					testhelpers.Clocking("2022-10-03T10:00", "2022-10-03T18:00"),
				},
			},
			want: 0,
		},
		{
			name: "Døgnvakt med kveldsarbeid",
			args: args{
				dutyPeriod: testhelpers.Period("2022-10-03T00:00", "2022-10-04T00:00"),
				compPeriod: testhelpers.Period("2022-10-03T20:00", "2022-10-04T00:00"),
				timesheet: []models.Clocking{
					testhelpers.Clocking("2022-10-03T23:10", "2022-10-04T00:00"),
				},
			},
			want: 190,
		},
		{
			name: "Døgnvakt med kveldsarbeid ved månedskifte",
			args: args{
				dutyPeriod: testhelpers.Period("2022-10-31T00:00", "2022-11-01T00:00"),
				compPeriod: testhelpers.Period("2022-10-31T20:00", "2022-11-01T00:00"),
				timesheet: []models.Clocking{
					testhelpers.Clocking("2022-10-31T23:10", "2022-11-01T00:00"),
				},
			},
			want: 190,
		},
		{
			name: "Døgnvakt ved månedskifte",
			args: args{
				dutyPeriod: testhelpers.Period("2022-07-31T00:00", "2022-08-01T00:00"),
				compPeriod: testhelpers.Period("2022-07-31T20:00", "2022-08-01T00:00"),
				timesheet:  []models.Clocking{},
			},
			want: 240,
		},
		{
			name: "Døgnvakt ved årskifte",
			args: args{
				dutyPeriod: testhelpers.Period("2022-12-31T00:00", "2023-01-01T00:00"),
				compPeriod: testhelpers.Period("2022-12-31T20:00", "2023-01-01T00:00"),
				timesheet:  []models.Clocking{},
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
				date: testhelpers.At("2022-10-30T10:00"),
				periods: []models.Period{
					testhelpers.Period("2022-10-30T00:00", "2022-10-31T00:00"),
				},
			},
			want: 60,
		},
		{
			name: "Har _ikke_ vakt når vi stiller klokken tilbake (normaltid)",
			args: args{
				date: testhelpers.At("2022-10-30T10:00"),
				periods: []models.Period{
					testhelpers.Period("2022-10-30T10:00", "2022-10-30T20:00"),
				},
			},
			want: 0,
		},
		{
			name: "Stiller klokken fremover (sommertid)",
			args: args{
				date: testhelpers.At("2022-03-27T10:00"),
				periods: []models.Period{
					testhelpers.Period("2022-03-27T00:00", "2022-03-28T00:00"),
				},
			},
			want: -60,
		},
		{
			name: "Har _ikke_ vakt når vi stiller klokken fremover (sommertid)",
			args: args{
				date: testhelpers.At("2022-03-27T10:00"),
				periods: []models.Period{
					testhelpers.Period("2022-03-27T10:00", "2022-03-27T20:00"),
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
				date:     testhelpers.D("2022-11-06"),
				formName: "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
			},
			want: testhelpers.Period("2022-11-06T09:00", "2022-11-06T14:30"),
		},
		{
			name: "Kjernetid for julaften",
			args: args{
				date:     testhelpers.D("2021-12-24"),
				formName: "Julaften 0800-1200 *",
			},
			want: testhelpers.Period("2021-12-24T08:00", "2021-12-24T12:00"),
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
				currentDay: testhelpers.WorkDay("2022-11-07", 7.75, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary500k,
					testhelpers.Clocking("2022-11-07T08:00", "2022-11-07T15:45")),
				period: testhelpers.Period("2022-11-07T00:00", "2022-11-08T00:00"),
			},
			want: 0,
		},
		{
			name: "Ingen arbeid i kjernetid",
			args: args{
				currentDay: testhelpers.WorkDay("2022-11-07", 7.75, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary500k,
					testhelpers.Clocking("2022-11-07T15:00", "2022-11-07T20:45")),
				period: testhelpers.Period("2022-11-07T00:00", "2022-11-08T00:00"),
			},
			want: 330,
		},
		{
			name: "Kom sent på jobb",
			args: args{
				currentDay: testhelpers.WorkDay("2022-11-07", 7.75, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary500k,
					testhelpers.Clocking("2022-11-07T10:00", "2022-11-07T17:45")),
				period: testhelpers.Period("2022-11-07T00:00", "2022-11-08T00:00"),
			},
			want: 60,
		},

		{
			name: "Julaften på en lørdag",
			args: args{
				currentDay: models.TimeSheet{
					Date:       testhelpers.D("2022-12-24"),
					WorkingDay: "Lørdag",
					FormName:   "BV Lørdag IKT",
					Clockings:  []models.Clocking{},
				},
				period: testhelpers.Period("2022-12-24T00:00", "2022-12-25T00:00"),
			},
			want: 0,
		},

		{
			name: "Julaften på en fredag (litt sen klokking)",
			args: args{
				currentDay: testhelpers.WorkDay("2021-12-24", 4, "Virkedag", "Julaften 0800-1200 *", testhelpers.Salary500k,
					testhelpers.Clocking("2021-12-24T09:00", "2021-12-24T14:00")),
				period: testhelpers.Period("2021-12-24T00:00", "2021-12-25T00:00"),
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
					"2022-12-24": {testhelpers.Period("2022-12-24T00:00", "2022-12-25T00:00")},
					"2022-12-25": {testhelpers.Period("2022-12-25T00:00", "2022-12-26T00:00")},
					"2022-12-26": {testhelpers.Period("2022-12-26T00:00", "2022-12-27T00:00")},
				},
				timesheet: map[string]models.TimeSheet{
					"2022-12-24": testhelpers.OffDay("2022-12-24", "Lørdag", "BV Lørdag IKT", testhelpers.Salary500k),
					"2022-12-25": testhelpers.OffDay("2022-12-25", "1. Juledag", "Helligdag", testhelpers.Salary500k),
					"2022-12-26": testhelpers.OffDay("2022-12-26", "2. Juledag", "Helligdag", testhelpers.Salary500k),
				},
			},
			want: map[string]models.GuardDuty{
				"2022-12-24": {
					Hvilende2000: 240,
					Hvilende0006: 360,
					Hvilende0620: 840,
					Helgetillegg: 1440,
					Skifttillegg: 0,
					IsWeekend:    true,
				},
				"2022-12-25": {
					Hvilende2000: 240,
					Hvilende0006: 360,
					Hvilende0620: 840,
					Helgetillegg: 1440,
					Skifttillegg: 0,
					IsWeekend:    true,
				},
				"2022-12-26": {
					Hvilende2000:  240,
					Hvilende0006:  360,
					Helligdag0620: 840,
					Helgetillegg:  0,
					Skifttillegg:  240,
					IsWeekend:     false,
				},
			},
		},

		{
			name: "Nyttårsaften på en lørdag",
			args: args{
				schedule: map[string][]models.Period{
					"2022-12-31": {testhelpers.Period("2022-12-31T00:00", "2023-01-01T00:00")},
				},
				timesheet: map[string]models.TimeSheet{
					"2022-12-31": testhelpers.OffDay("2022-12-31", "Lørdag", "BV Lørdag IKT", testhelpers.Salary500k),
				},
			},
			want: map[string]models.GuardDuty{
				"2022-12-31": {
					Hvilende2000: 240,
					Hvilende0006: 360,
					Hvilende0620: 840,
					Helgetillegg: 1440,
					Skifttillegg: 0,
					IsWeekend:    true,
				},
			},
		},

		{
			name: "Julaften på en fredag",
			args: args{
				schedule: map[string][]models.Period{
					"2021-12-24": {testhelpers.Period("2021-12-24T00:00", "2021-12-25T00:00")},
					"2021-12-25": {testhelpers.Period("2021-12-25T00:00", "2021-12-26T00:00")},
					"2021-12-26": {testhelpers.Period("2021-12-26T00:00", "2021-12-27T00:00")},
				},
				timesheet: map[string]models.TimeSheet{
					"2021-12-24": testhelpers.OffDay("2021-12-24", "Virkedag", "Julaften 0800-1200 *", testhelpers.Salary500k),
					"2021-12-25": testhelpers.OffDay("2021-12-25", "1. Juledag", "Helligdag", testhelpers.Salary500k),
					"2021-12-26": testhelpers.OffDay("2021-12-26", "2. Juledag", "Helligdag", testhelpers.Salary500k),
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
					Hvilende2000: 240,
					Hvilende0006: 360,
					Hvilende0620: 840,
					Helgetillegg: 1440,
					Skifttillegg: 0,
					IsWeekend:    true,
				},
				"2021-12-26": {
					Hvilende2000: 240,
					Hvilende0006: 360,
					Hvilende0620: 840,
					Helgetillegg: 1440,
					Skifttillegg: 0,
					IsWeekend:    true,
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
					"2022-10-12": testhelpers.WorkDay("2022-10-12", 7.75, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary750k,
						testhelpers.Clocking("2022-10-12T08:00", "2022-10-12T15:45")),
					"2022-10-13": testhelpers.WorkDay("2022-10-13", 7.75, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary750k,
						testhelpers.Clocking("2022-10-13T08:00", "2022-10-13T15:45")),
					"2022-10-14": testhelpers.WorkDay("2022-10-14", 7.75, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary750k,
						testhelpers.Clocking("2022-10-14T08:00", "2022-10-14T15:45")),
					"2022-10-15": testhelpers.OffDay("2022-10-15", "Lørdag", "BV Lørdag IKT", testhelpers.Salary750k),
					"2022-10-16": testhelpers.OffDay("2022-10-16", "Søndag", "BV Søndag IKT", testhelpers.Salary750k),
					"2022-10-17": testhelpers.WorkDay("2022-10-17", 7.75, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary750k,
						testhelpers.Clocking("2022-10-17T08:00", "2022-10-17T15:45")),
					"2022-10-18": testhelpers.WorkDay("2022-10-18", 7.75, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary750k,
						testhelpers.Clocking("2022-10-18T08:00", "2022-10-18T15:45")),
					"2022-10-19": testhelpers.WorkDay("2022-10-19", 7.75, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary750k,
						testhelpers.Clocking("2022-10-19T08:00", "2022-10-19T15:45")),
				},
				guardPeriod: map[string][]models.Period{
					"2022-10-12": {testhelpers.Period("2022-10-12T12:00", "2022-10-13T00:00")},
					"2022-10-13": {testhelpers.Period("2022-10-13T00:00", "2022-10-14T00:00")},
					"2022-10-14": {testhelpers.Period("2022-10-14T00:00", "2022-10-15T00:00")},
					"2022-10-15": {testhelpers.Period("2022-10-15T00:00", "2022-10-16T00:00")},
					"2022-10-16": {testhelpers.Period("2022-10-16T00:00", "2022-10-17T00:00")},
					"2022-10-17": {testhelpers.Period("2022-10-17T00:00", "2022-10-18T00:00")},
					"2022-10-18": {testhelpers.Period("2022-10-18T00:00", "2022-10-19T00:00")},
					"2022-10-19": {testhelpers.Period("2022-10-19T00:00", "2022-10-19T12:00")},
				},
			},
			want: models.Artskoder{
				Morgen: models.Artskode{Sum: testhelpers.Dec(5614.86), Hours: 30},
				Kveld:  models.Artskode{Sum: testhelpers.Dec(3743.24), Hours: 20},
				Dag:    models.Artskode{Sum: testhelpers.Dec(4235.27), Hours: 31},
				Helg:   models.Artskode{Sum: testhelpers.Dec(9327.78), Hours: 48},
				Skift:  models.Artskode{Sum: testhelpers.DecI(100), Hours: 20},
			},
		},

		{
			name: "En døgnkontinuerlig vaktuke (normaltid)",
			args: args{
				timesheet: map[string]models.TimeSheet{
					"2022-10-12": testhelpers.WorkDay("2022-10-12", 7.5, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary750k,
						testhelpers.Clocking("2022-10-12T08:00", "2022-10-12T15:30")),
					"2022-10-13": testhelpers.WorkDay("2022-10-13", 7.5, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary750k,
						testhelpers.Clocking("2022-10-13T08:00", "2022-10-13T15:30")),
					"2022-10-14": testhelpers.WorkDay("2022-10-14", 7.5, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary750k,
						testhelpers.Clocking("2022-10-14T08:00", "2022-10-14T15:30")),
					"2022-10-15": testhelpers.OffDay("2022-10-15", "Lørdag", "BV Lørdag IKT", testhelpers.Salary750k),
					"2022-10-16": testhelpers.OffDay("2022-10-16", "Søndag", "BV Søndag IKT", testhelpers.Salary750k),
					"2022-10-17": testhelpers.WorkDay("2022-10-17", 7.5, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary750k,
						testhelpers.Clocking("2022-10-17T08:00", "2022-10-17T15:30")),
					"2022-10-18": testhelpers.WorkDay("2022-10-18", 7.5, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary750k,
						testhelpers.Clocking("2022-10-18T08:00", "2022-10-18T15:30")),
					"2022-10-19": testhelpers.WorkDay("2022-10-19", 7.5, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary750k,
						testhelpers.Clocking("2022-10-19T08:00", "2022-10-19T15:30")),
				},
				guardPeriod: map[string][]models.Period{
					"2022-10-12": {testhelpers.Period("2022-10-12T12:00", "2022-10-13T00:00")},
					"2022-10-13": {testhelpers.Period("2022-10-13T00:00", "2022-10-14T00:00")},
					"2022-10-14": {testhelpers.Period("2022-10-14T00:00", "2022-10-15T00:00")},
					"2022-10-15": {testhelpers.Period("2022-10-15T00:00", "2022-10-16T00:00")},
					"2022-10-16": {testhelpers.Period("2022-10-16T00:00", "2022-10-17T00:00")},
					"2022-10-17": {testhelpers.Period("2022-10-17T00:00", "2022-10-18T00:00")},
					"2022-10-18": {testhelpers.Period("2022-10-18T00:00", "2022-10-19T00:00")},
					"2022-10-19": {testhelpers.Period("2022-10-19T00:00", "2022-10-19T12:00")},
				},
			},
			want: models.Artskoder{
				Morgen: models.Artskode{Sum: testhelpers.Dec(5614.86), Hours: 30},
				Kveld:  models.Artskode{Sum: testhelpers.Dec(3743.24), Hours: 20},
				Dag:    models.Artskode{Sum: testhelpers.Dec(4508.51), Hours: 33},
				Helg:   models.Artskode{Sum: testhelpers.Dec(9327.78), Hours: 48},
				Skift:  models.Artskode{Sum: testhelpers.DecI(100), Hours: 20},
			},
		},
		{
			name: "Vakt en lørdag med utrykning",
			args: args{
				timesheet: map[string]models.TimeSheet{
					"2026-01-10": {
						Date:         testhelpers.D("2026-01-10"),
						WorkingHours: 7.5,
						WorkingDay:   "Lørdag",
						FormName:     "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Salary:       testhelpers.Salary500k,
						Clockings: []models.Clocking{
							testhelpers.ClockingOtG("2026-01-10T20:00", "2026-01-10T22:00"),
						},
					},
				},
				guardPeriod: map[string][]models.Period{
					"2026-01-10": {testhelpers.Period("2026-01-10T00:00", "2026-01-11T00:00")},
				},
			},
			want: models.Artskoder{
				Helg:      models.Artskode{Sum: testhelpers.Dec(3366.59), Hours: 24},
				Utrykning: models.Artskode{Sum: testhelpers.DecI(130), Hours: 2},
			},
		},
		{
			name: "Vakt en arbeidsdag med utrykning under utvidetarbeidstid",
			args: args{
				timesheet: map[string]models.TimeSheet{
					"2026-01-27": {
						Date:         testhelpers.D("2026-01-27"),
						WorkingHours: 7.5,
						WorkingDay:   "Lørdag",
						FormName:     "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Salary:       testhelpers.Salary500k,
						Clockings: []models.Clocking{
							testhelpers.ClockingOtG("2026-01-27T17:00", "2026-01-27T19:00"),
						},
					},
				},
				guardPeriod: map[string][]models.Period{
					"2026-01-27": {testhelpers.Period("2026-01-27T00:00", "2026-01-28T00:00")},
				},
			},
			want: models.Artskoder{
				Morgen:    models.Artskode{Sum: testhelpers.Dec(798.65), Hours: 6},
				Kveld:     models.Artskode{Sum: testhelpers.Dec(532.43), Hours: 4},
				Dag:       models.Artskode{Sum: testhelpers.Dec(672.57), Hours: 7},
				Helg:      models.Artskode{},
				Skift:     models.Artskode{Sum: testhelpers.DecI(20), Hours: 4},
				Utrykning: models.Artskode{Sum: testhelpers.DecI(50), Hours: 2},
			},
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
				Satser:       testhelpers.DefaultSatser,
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
	salary := testhelpers.Salary500k
	satser := testhelpers.DefaultSatser

	holidayPayroll := &models.Payroll{}
	holidayMinutes := map[string]models.GuardDuty{
		"2022-10-15": {
			Hvilende2000: 240,
			Hvilende0006: 360,
			Hvilende0620: 840,
			Helgetillegg: 1440,
			Skifttillegg: 0,
			IsWeekend:    true,
		},
	}
	kronetillegg.Calculate(holidayMinutes, satser, holidayPayroll)
	overtime.Calculate(holidayMinutes, salary, holidayPayroll)

	regularPayroll := &models.Payroll{}
	regularMinutes := map[string]models.GuardDuty{
		"2022-10-15": {
			Hvilende2000: 240,
			Hvilende0006: 360,
			Hvilende0620: 840,
			Helgetillegg: 1440,
			Skifttillegg: 0,
			IsWeekend:    true,
		},
	}
	kronetillegg.Calculate(regularMinutes, satser, regularPayroll)
	overtime.Calculate(regularMinutes, salary, regularPayroll)

	if diff := cmp.Diff(holidayPayroll.Artskoder, regularPayroll.Artskoder); diff != "" {
		t.Errorf("Calculate() holiday on saturday did not return the same as a normal saturday")
		t.Errorf("Calculate() mismatch (-want +got):\n%s", diff)
	}
}

// TestCalculateIsNotEqualForHolidayOnMondayAndRegularMonday tester at man ikke får den samme lønnen for en vanlig arbeidsdag
// uten arbeid og en arbeidsdag som er en helligdag.
func TestCalculateIsNotEqualForHolidayOnMondayAndRegularMonday(t *testing.T) {
	salary := testhelpers.Salary500k
	satser := testhelpers.DefaultSatser

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
	kronetillegg.Calculate(holidayMinutes, satser, holidayPayroll)
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
	kronetillegg.Calculate(regularMinutes, satser, regularPayroll)
	overtime.Calculate(regularMinutes, salary, regularPayroll)

	if diff := cmp.Diff(holidayPayroll.Artskoder, regularPayroll.Artskoder); diff == "" {
		t.Errorf("Calculate() holiday on monday returns the same as a normal monday")
		t.Errorf("Calculate() mismatch (-want +got):\n%s", diff)
	}
}
