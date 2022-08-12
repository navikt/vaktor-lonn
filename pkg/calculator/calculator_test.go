package calculator

import (
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

func Test_calculateMinutesWithGuardDutyInPeriod(t *testing.T) {
	type args struct {
		report     *models.Report
		day        string
		dutyPeriod models.Period
		compPeriod models.Period
		timesheet  []string
	}

	day := "08.08.2022"
	report := &models.Report{
		TimesheetEachDay: map[string]models.Timesheet{},
	}
	timesheet := models.Timesheet{}
	report.TimesheetEachDay[day] = timesheet

	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "Vanlig arbeidsdag",
			args: args{
				report:     report,
				day:        day,
				dutyPeriod: models.Period{Begin: "00:00", End: "24:00"},
				compPeriod: models.Period{Begin: "09:00", End: "14:30"},
				timesheet:  []string{"08:00-15:00"},
			},
			want: 0,
		},
		{
			name: "Uvanlig kort arbeidsdag",
			args: args{
				report:     report,
				day:        day,
				dutyPeriod: models.Period{Begin: "00:00", End: "24:00"},
				compPeriod: models.Period{Begin: "09:00", End: "14:30"},
				timesheet:  []string{"10:00-14:00"},
			},
			want: 90,
		},
		{
			name: "Forskjøvet arbeidsdag",
			args: args{
				report:     report,
				day:        day,
				dutyPeriod: models.Period{Begin: "00:00", End: "24:00"},
				compPeriod: models.Period{Begin: "09:00", End: "14:30"},
				timesheet:  []string{"10:00-18:00"},
			},
			want: 60,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateMinutesWithGuardDutyInPeriod(tt.args.report, tt.args.day, tt.args.dutyPeriod, tt.args.compPeriod, tt.args.timesheet)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateMinutesWithGuardDutyInPeriod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("calculateMinutesWithGuardDutyInPeriod() got = %v, want %v", got, tt.want)
			}
		})
	}
}
