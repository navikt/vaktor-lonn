package calculator

import (
	"github.com/navikt/vaktor-lonn/pkg/dummy"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"reflect"
	"testing"
)

func Test_timeToMinutes(t *testing.T) {
	type args struct {
		clock string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "first test",
			args: args{
				"02:33",
			},
			want:    153,
			wantErr: false,
		},
		{
			name: "no leading zero padding for hour",
			args: args{
				"7:03",
			},
			want:    423,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := timeToMinutes(tt.args.clock)
			if (err != nil) != tt.wantErr {
				t.Errorf("timeToMinutes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("timeToMinutes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateEarnings(t *testing.T) {
	type args struct {
		minutes map[string]guardDutyMinutes
		salary  int
	}
	pocPeriode := map[string]models.Period{
		"14.03.2022": {
			Fra:       "00:00",
			Til:       "24:00",
			Helligdag: false,
		},
		"15.03.2022": {
			Fra:       "00:00",
			Til:       "24:00",
			Helligdag: false,
		},
		"16.03.2022": {
			Fra:       "00:00",
			Til:       "24:00",
			Helligdag: false,
		},
		"17.03.2022": {
			Fra:       "00:00",
			Til:       "24:00",
			Helligdag: true,
		},
		"18.03.2022": {
			Fra:       "00:00",
			Til:       "24:00",
			Helligdag: false,
		},
		"19.03.2022": {
			Fra:       "00:00",
			Til:       "24:00",
			Helligdag: false,
		},
		"20.03.2022": {
			Fra:       "00:00",
			Til:       "24:00",
			Helligdag: false,
		},
	}
	minutes, _ := ParsePeriode(pocPeriode, dummy.GetMinWinTidPOC())
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "vs poc",
			args: args{
				minutes: minutes,
				salary:  500_000,
			},
			want: 15482.82,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateEarnings(tt.args.minutes, tt.args.salary)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateEarnings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CalculateEarnings() got = %v, want %v", got, tt.want)
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
		want int
	}{
		// TODO: Add test cases.
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
		dutyBegin string
		dutyEnd   string
		begin     string
		end       string
	}
	tests := []struct {
		name    string
		args    args
		want    *Range
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "døgn duty",
			args: args{
				dutyBegin: "00:00",
				dutyEnd:   "24:00",
				begin:     "06:00",
				end:       "20:00",
			},
			want: &Range{Begin: 360, End: 1200},
		},
		{
			name: "short duty",
			args: args{
				dutyBegin: "04:00",
				dutyEnd:   "07:00",
				begin:     "06:00",
				end:       "20:00",
			},
			want: &Range{Begin: 360, End: 420},
		},
		{
			name: "no duty",
			args: args{
				dutyBegin: "07:00",
				dutyEnd:   "17:00",
				begin:     "00:00",
				end:       "06:00",
			},
			want: nil,
		},
		{
			name: "late work duty",
			args: args{
				dutyBegin: "14:00",
				dutyEnd:   "24:00",
				begin:     "06:00",
				end:       "20:00",
			},
			want: &Range{Begin: 840, End: 1200},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createRangeForPeriod("09.07.1987", tt.args.dutyBegin, tt.args.dutyEnd, tt.args.begin, tt.args.end)
			if (err != nil) != tt.wantErr {
				t.Errorf("createRangeForPeriod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createRangeForPeriod() got = %v, want %v", got, tt.want)
			}
		})
	}
}
