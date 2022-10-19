package calculator

import (
	"github.com/navikt/vaktor-lonn/pkg/models"
	"reflect"
	"testing"
	"time"
)

func Test_timeToMinutes(t *testing.T) {
	type args struct {
		clock time.Time
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "first test",
			args: args{
				time.Date(2022, 10, 3, 2, 33, 0, 0, time.UTC),
			},
			want: 153,
		},
		{
			name: "no leading zero padding for hour",
			args: args{
				time.Date(2022, 10, 3, 7, 3, 0, 0, time.UTC),
			},
			want: 423,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := timeToMinutes(tt.args.clock)
			if got != tt.want {
				t.Errorf("timeToMinutes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateWorkInPeriode(t *testing.T) {
	type args struct {
		work   Range
		period Range
	}
	periodRange := Range{
		Begin: 06 * 60,
		End:   20 * 60,
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "no work done",
			args: args{
				work: Range{0, 0},
			},
			want: 0,
		},
		{
			name: "worked all day long",
			args: args{
				work: Range{6 * 60, 20 * 60},
			},
			want: 840,
		},
		{
			name: "some work done",
			args: args{
				work: Range{8 * 60, 15 * 60},
			},
			want: 420,
		},
		{
			name: "work done late",
			args: args{
				work: Range{19 * 60, 21 * 60},
			},
			want: 60,
		},
		{
			name: "work done before day",
			args: args{
				work: Range{3 * 60, 6 * 60},
			},
			want: 0,
		},
		{
			name: "work done after day",
			args: args{
				work: Range{20 * 60, 22 * 60},
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateMinutesOverlappingInPeriods(tt.args.work, periodRange); got != tt.want {
				t.Errorf("calculateMinutesOverlappingInPeriods() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createRangeForPeriod(t *testing.T) {
	type args struct {
		period    models.Period
		threshold models.Period
	}
	tests := []struct {
		name string
		args args
		want *Range
	}{
		{
			name: "døgnvakt 06-20",
			args: args{
				period: models.Period{
					Begin: time.Date(1987, 7, 9, 0, 0, 0, 0, time.UTC),
					End:   time.Date(1987, 7, 10, 0, 0, 0, 0, time.UTC),
				},
				threshold: models.Period{
					Begin: time.Date(1987, 7, 9, 6, 0, 0, 0, time.UTC),
					End:   time.Date(1987, 7, 9, 20, 0, 0, 0, time.UTC),
				},
			},
			want: &Range{Begin: 360, End: 1200},
		},
		{
			name: "døgnvakt 20-00",
			args: args{
				period: models.Period{
					Begin: time.Date(1987, 7, 9, 0, 0, 0, 0, time.UTC),
					End:   time.Date(1987, 7, 10, 0, 0, 0, 0, time.UTC),
				},
				threshold: models.Period{
					Begin: time.Date(1987, 7, 9, 20, 0, 0, 0, time.UTC),
					End:   time.Date(1987, 7, 10, 0, 0, 0, 0, time.UTC),
				},
			},
			want: &Range{Begin: 1200, End: 1440},
		},
		{
			name: "short duty",
			args: args{
				period: models.Period{
					Begin: time.Date(1987, 7, 9, 4, 0, 0, 0, time.UTC),
					End:   time.Date(1987, 7, 9, 7, 0, 0, 0, time.UTC),
				},
				threshold: models.Period{
					Begin: time.Date(1987, 7, 9, 6, 0, 0, 0, time.UTC),
					End:   time.Date(1987, 7, 9, 20, 0, 0, 0, time.UTC),
				},
			},
			want: &Range{Begin: 360, End: 420},
		},
		{
			name: "no duty",
			args: args{
				period: models.Period{
					Begin: time.Date(1987, 7, 9, 7, 0, 0, 0, time.UTC),
					End:   time.Date(1987, 7, 9, 17, 0, 0, 0, time.UTC),
				},
				threshold: models.Period{
					Begin: time.Date(1987, 7, 9, 0, 0, 0, 0, time.UTC),
					End:   time.Date(1987, 7, 9, 6, 0, 0, 0, time.UTC),
				},
			},
			want: nil,
		},
		{
			name: "late work duty",
			args: args{
				period: models.Period{
					Begin: time.Date(1987, 7, 9, 14, 0, 0, 0, time.UTC),
					End:   time.Date(1987, 7, 10, 0, 0, 0, 0, time.UTC),
				},
				threshold: models.Period{
					Begin: time.Date(1987, 7, 9, 6, 0, 0, 0, time.UTC),
					End:   time.Date(1987, 7, 9, 20, 0, 0, 0, time.UTC),
				},
			},
			want: &Range{Begin: 840, End: 1200},
		},
		{
			name: "work till duty begins",
			args: args{
				period: models.Period{
					Begin: time.Date(1987, 7, 9, 6, 0, 0, 0, time.UTC),
					End:   time.Date(1987, 7, 9, 9, 0, 0, 0, time.UTC),
				},
				threshold: models.Period{
					Begin: time.Date(1987, 7, 9, 0, 0, 0, 0, time.UTC),
					End:   time.Date(1987, 7, 9, 6, 0, 0, 0, time.UTC),
				},
			},
			want: nil,
		},
		{
			name: "work outside of duty",
			args: args{
				period: models.Period{
					Begin: time.Date(1987, 7, 9, 6, 0, 0, 0, time.UTC),
					End:   time.Date(1987, 7, 9, 9, 59, 59, 0, time.UTC),
				},
				threshold: models.Period{
					Begin: time.Date(1987, 7, 9, 20, 0, 0, 0, time.UTC),
					End:   time.Date(1987, 7, 10, 0, 0, 0, 0, time.UTC),
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createRangeForPeriod(tt.args.period, tt.args.threshold)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createRangeForPeriod() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateMinutesWithGuardDutyInPeriod(t *testing.T) {
	type args struct {
		day        string
		dutyPeriod models.Period
		compPeriod models.Period
		timesheet  []models.Clocking
	}

	day := "2022-08-08"

	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "Vanlig arbeidsdag",
			args: args{
				day: day,
				dutyPeriod: models.Period{
					Begin: time.Date(2022, 10, 3, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2022, 10, 3, 23, 59, 59, 0, time.UTC),
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
				day: day,
				dutyPeriod: models.Period{
					Begin: time.Date(2022, 10, 3, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2022, 10, 3, 23, 59, 59, 0, time.UTC),
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
				day: day,
				dutyPeriod: models.Period{
					Begin: time.Date(2022, 10, 3, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2022, 10, 3, 23, 59, 59, 0, time.UTC),
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
				day: day,
				dutyPeriod: models.Period{
					Begin: time.Date(2022, 10, 3, 6, 0, 0, 0, time.UTC),
					End:   time.Date(2022, 10, 3, 9, 0, 0, 0, time.UTC),
				},
				compPeriod: models.Period{
					Begin: time.Date(2022, 10, 3, 20, 0, 0, 0, time.UTC),
					End:   time.Date(2022, 10, 3, 23, 59, 59, 0, time.UTC),
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
