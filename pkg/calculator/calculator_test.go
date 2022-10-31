package calculator

import (
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
