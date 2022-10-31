package ranges

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
			got := timeToMinutes(tt.args.clock.Hour(), tt.args.clock.Minute())
			if got != tt.want {
				t.Errorf("timeToMinutes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromTime(t *testing.T) {
	type args struct {
		in  time.Time
		out time.Time
	}
	tests := []struct {
		name string
		args args
		want Range
	}{
		{
			name: "",
			args: args{
				in:  time.Date(2022, 10, 31, 14, 30, 0, 0, time.UTC),
				out: time.Date(2022, 10, 31, 16, 30, 0, 0, time.UTC),
			},
			want: Range{
				Begin: 870,
				End:   990,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromTime(tt.args.in, tt.args.out); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateWorkInPeriode(t *testing.T) {
	type args struct {
		work   Range
		period Range
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
				period: Range{
					Begin: 06 * 60,
					End:   20 * 60,
				},
			},
			want: 0,
		},
		{
			name: "worked all day long",
			args: args{
				work: Range{6 * 60, 20 * 60},
				period: Range{
					Begin: 06 * 60,
					End:   20 * 60,
				},
			},
			want: 840,
		},
		{
			name: "some work done",
			args: args{
				work: Range{8 * 60, 15 * 60},
				period: Range{
					Begin: 06 * 60,
					End:   20 * 60,
				},
			},
			want: 420,
		},
		{
			name: "work done late",
			args: args{
				work: Range{19 * 60, 21 * 60},
				period: Range{
					Begin: 06 * 60,
					End:   20 * 60,
				},
			},
			want: 60,
		},
		{
			name: "work done before day",
			args: args{
				work: Range{3 * 60, 6 * 60},
				period: Range{
					Begin: 06 * 60,
					End:   20 * 60,
				},
			},
			want: 0,
		},
		{
			name: "work done after day",
			args: args{
				work: Range{20 * 60, 22 * 60},
				period: Range{
					Begin: 06 * 60,
					End:   20 * 60,
				},
			},
			want: 0,
		},
		{
			name: "work late nights",
			args: args{
				work: Range{23 * 60, 24 * 60},
				period: Range{
					Begin: 0,
					End:   24 * 60,
				},
			},
			want: 60,
		},
		{
			name: "work late nights part 2",
			args: args{
				work: Range{0, 2 * 60},
				period: Range{
					Begin: 0,
					End:   24 * 60,
				},
			},
			want: 120,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateMinutesOverlapping(tt.args.work, tt.args.period); got != tt.want {
				t.Errorf("CalculateMinutesOverlappingInPeriods() = %v, want %v", got, tt.want)
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
		{
			name: "work at end of month",
			args: args{
				period: models.Period{
					Begin: time.Date(1987, 9, 30, 20, 0, 0, 0, time.UTC),
					End:   time.Date(1987, 10, 1, 0, 0, 0, 0, time.UTC),
				},
				threshold: models.Period{
					Begin: time.Date(1987, 9, 30, 0, 0, 0, 0, time.UTC),
					End:   time.Date(1987, 10, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			want: &Range{Begin: 1200, End: 1440},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateForPeriod(tt.args.period, tt.args.threshold)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateRangeForPeriod() got = %v, want %v", got, tt.want)
			}
		})
	}
}