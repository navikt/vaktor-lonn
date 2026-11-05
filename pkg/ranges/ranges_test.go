package ranges

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/navikt/vaktor-lonn/pkg/testhelpers"
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
				testhelpers.At("2022-10-03T02:33"),
			},
			want: 153,
		},
		{
			name: "no leading zero padding for hour",
			args: args{
				testhelpers.At("2022-10-03T07:03"),
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
			name: "2 timers range",
			args: args{
				in:  testhelpers.At("2022-10-31T14:30"),
				out: testhelpers.At("2022-10-31T16:30"),
			},
			want: Range{
				Begin: 870,
				End:   990,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromTime(tt.args.in, tt.args.out)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("FromTime() mismatch (-want +got):\n%s", diff)
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
				work:   Range{0, 0},
				period: Range{Begin: 6 * 60, End: 20 * 60},
			},
			want: 0,
		},
		{
			name: "worked all day long",
			args: args{
				work:   Range{6 * 60, 20 * 60},
				period: Range{Begin: 6 * 60, End: 20 * 60},
			},
			want: 840,
		},
		{
			name: "some work done",
			args: args{
				work:   Range{8 * 60, 15 * 60},
				period: Range{Begin: 6 * 60, End: 20 * 60},
			},
			want: 420,
		},
		{
			name: "work done late",
			args: args{
				work:   Range{19 * 60, 21 * 60},
				period: Range{Begin: 6 * 60, End: 20 * 60},
			},
			want: 60,
		},
		{
			name: "work done before day",
			args: args{
				work:   Range{3 * 60, 6 * 60},
				period: Range{Begin: 6 * 60, End: 20 * 60},
			},
			want: 0,
		},
		{
			name: "work done after day",
			args: args{
				work:   Range{20 * 60, 22 * 60},
				period: Range{Begin: 6 * 60, End: 20 * 60},
			},
			want: 0,
		},
		{
			name: "work late nights",
			args: args{
				work:   Range{23 * 60, 24 * 60},
				period: Range{Begin: 0, End: 24 * 60},
			},
			want: 60,
		},
		{
			name: "work late nights part 2",
			args: args{
				work:   Range{0, 2 * 60},
				period: Range{Begin: 0, End: 24 * 60},
			},
			want: 120,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateMinutesOverlapping(tt.args.work, tt.args.period); got != tt.want {
				t.Errorf("CalculateMinutesOverlapping() = %v, want %v", got, tt.want)
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
				period:    testhelpers.Period("1987-07-09T00:00", "1987-07-10T00:00"),
				threshold: testhelpers.Period("1987-07-09T06:00", "1987-07-09T20:00"),
			},
			want: &Range{Begin: 360, End: 1200},
		},
		{
			name: "døgnvakt 20-00",
			args: args{
				period:    testhelpers.Period("1987-07-09T00:00", "1987-07-10T00:00"),
				threshold: testhelpers.Period("1987-07-09T20:00", "1987-07-10T00:00"),
			},
			want: &Range{Begin: 1200, End: 1440},
		},
		{
			name: "short duty",
			args: args{
				period:    testhelpers.Period("1987-07-09T04:00", "1987-07-09T07:00"),
				threshold: testhelpers.Period("1987-07-09T06:00", "1987-07-09T20:00"),
			},
			want: &Range{Begin: 360, End: 420},
		},
		{
			name: "no duty",
			args: args{
				period:    testhelpers.Period("1987-07-09T07:00", "1987-07-09T17:00"),
				threshold: testhelpers.Period("1987-07-09T00:00", "1987-07-09T06:00"),
			},
			want: nil,
		},
		{
			name: "late work duty",
			args: args{
				period:    testhelpers.Period("1987-07-09T14:00", "1987-07-10T00:00"),
				threshold: testhelpers.Period("1987-07-09T06:00", "1987-07-09T20:00"),
			},
			want: &Range{Begin: 840, End: 1200},
		},
		{
			name: "work till duty begins",
			args: args{
				period:    testhelpers.Period("1987-07-09T06:00", "1987-07-09T09:00"),
				threshold: testhelpers.Period("1987-07-09T00:00", "1987-07-09T06:00"),
			},
			want: nil,
		},
		{
			name: "work outside of duty",
			args: args{
				period:    testhelpers.Period("1987-07-09T06:00", "1987-07-09T09:59:59"),
				threshold: testhelpers.Period("1987-07-09T20:00", "1987-07-10T00:00"),
			},
			want: nil,
		},
		{
			name: "work at end of month",
			args: args{
				period:    testhelpers.Period("1987-09-30T20:00", "1987-10-01T00:00"),
				threshold: testhelpers.Period("1987-09-30T00:00", "1987-10-01T00:00"),
			},
			want: &Range{Begin: 1200, End: 1440},
		},
		{
			name: "work at end of year 20-00 (torsdag-fredag)",
			args: args{
				period:    testhelpers.Period("1987-12-31T00:00", "1988-01-01T00:00"),
				threshold: testhelpers.Period("1987-12-31T20:00", "1988-01-01T00:00"),
			},
			want: &Range{Begin: 1200, End: 1440},
		},
		{
			name: "work at end of year 20-00 (lørdag til søndag)",
			args: args{
				period:    testhelpers.Period("2022-12-31T00:00", "2023-01-01T00:00"),
				threshold: testhelpers.Period("2022-12-31T20:00", "2023-01-01T00:00"),
			},
			want: &Range{Begin: 1200, End: 1440},
		},
		{
			name: "work at end of year 00-00 (lørdag til søndag)",
			args: args{
				period:    testhelpers.Period("1987-12-31T00:00", "1988-01-01T00:00"),
				threshold: testhelpers.Period("1987-12-31T00:00", "1988-01-01T00:00"),
			},
			want: &Range{Begin: 0, End: 1440},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateForPeriod(tt.args.period, tt.args.threshold)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("CreateForPeriod() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
