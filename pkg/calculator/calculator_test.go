package calculator

import (
	"github.com/google/go-cmp/cmp"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"testing"
	"time"
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
		date       time.Time
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
					WorkingDay: "Virkedag",
					FormName:   "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 11, 6, 8, 0, 0, 0, time.UTC),
							Out: time.Date(2022, 11, 6, 15, 45, 0, 0, time.UTC),
						},
					},
				},
				date: time.Date(2022, 11, 6, 0, 0, 0, 0, time.UTC),
				period: models.Period{
					Begin: time.Date(2022, 11, 6, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2022, 11, 7, 0, 0, 0, 0, time.UTC),
				},
			},
			want: 0,
		},
		{
			name: "Ingen arbeid i kjernetid",
			args: args{
				currentDay: models.TimeSheet{
					WorkingDay: "Virkedag",
					FormName:   "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 11, 6, 15, 0, 0, 0, time.UTC),
							Out: time.Date(2022, 11, 6, 20, 45, 0, 0, time.UTC),
						},
					},
				},
				date: time.Date(2022, 11, 6, 0, 0, 0, 0, time.UTC),
				period: models.Period{
					Begin: time.Date(2022, 11, 6, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2022, 11, 7, 0, 0, 0, 0, time.UTC),
				},
			},
			want: 330,
		},
		{
			name: "Kom sent på jobb",
			args: args{
				currentDay: models.TimeSheet{
					WorkingDay: "Virkedag",
					FormName:   "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 11, 6, 10, 0, 0, 0, time.UTC),
							Out: time.Date(2022, 11, 6, 17, 45, 0, 0, time.UTC),
						},
					},
				},
				date: time.Date(2022, 11, 6, 0, 0, 0, 0, time.UTC),
				period: models.Period{
					Begin: time.Date(2022, 11, 6, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2022, 11, 7, 0, 0, 0, 0, time.UTC),
				},
			},
			want: 60,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateGuardDutyInKjernetid(tt.args.currentDay, tt.args.date, tt.args.period); got != tt.want {
				t.Errorf("calculateGuardDutyInKjernetid() = %v, want %v", got, tt.want)
			}
		})
	}
}
