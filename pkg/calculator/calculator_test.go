package calculator

import (
	"github.com/navikt/vaktor-lonn/pkg/models"
	"reflect"
	"testing"
	"time"
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
		name    string
		args    args
		want    *Range
		wantErr bool
	}{
		{
			name: "døgn duty",
			args: args{
				period: models.Period{
					Begin: time.Date(1987, 7, 9, 0, 0, 0, 0, time.UTC),
					End:   time.Date(1987, 7, 9, 23, 59, 59, 0, time.UTC),
				},
				threshold: models.Period{
					Begin: time.Date(1987, 7, 9, 6, 0, 0, 0, time.UTC),
					End:   time.Date(1987, 7, 9, 20, 0, 0, 0, time.UTC),
				},
			},
			want: &Range{Begin: 360, End: 1200},
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
					End:   time.Date(1987, 7, 9, 23, 59, 59, 0, time.UTC),
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
					End:   time.Date(1987, 7, 9, 23, 59, 59, 0, time.UTC),
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createRangeForPeriod(tt.args.period, tt.args.threshold)
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
				report: report,
				day:    day,
				dutyPeriod: models.Period{
					Begin: time.Date(1987, 7, 9, 0, 0, 0, 0, time.UTC),
					End:   time.Date(1987, 7, 9, 23, 59, 59, 0, time.UTC),
				},
				compPeriod: models.Period{
					Begin: time.Date(1987, 7, 9, 9, 0, 0, 0, time.UTC),
					End:   time.Date(1987, 7, 9, 14, 30, 0, 0, time.UTC),
				},
				timesheet: []string{"08:00-15:00"},
			},
			want: 0,
		},
		{
			name: "Uvanlig kort arbeidsdag",
			args: args{
				report: report,
				day:    day,
				dutyPeriod: models.Period{
					Begin: time.Date(1987, 7, 9, 0, 0, 0, 0, time.UTC),
					End:   time.Date(1987, 7, 9, 23, 59, 59, 0, time.UTC),
				},
				compPeriod: models.Period{
					Begin: time.Date(1987, 7, 9, 9, 0, 0, 0, time.UTC),
					End:   time.Date(1987, 7, 9, 14, 30, 0, 0, time.UTC),
				},
				timesheet: []string{"10:00-14:00"},
			},
			want: 90,
		},
		{
			name: "Forskjøvet arbeidsdag",
			args: args{
				report: report,
				day:    day,
				dutyPeriod: models.Period{
					Begin: time.Date(1987, 7, 9, 0, 0, 0, 0, time.UTC),
					End:   time.Date(1987, 7, 9, 23, 59, 59, 0, time.UTC),
				},
				compPeriod: models.Period{
					Begin: time.Date(1987, 7, 9, 9, 0, 0, 0, time.UTC),
					End:   time.Date(1987, 7, 9, 14, 30, 0, 0, time.UTC),
				},
				timesheet: []string{"10:00-18:00"},
			},
			want: 60,
		},
		{
			name: "Morgenvakt",
			args: args{
				report: report,
				day:    day,
				dutyPeriod: models.Period{
					Begin: time.Date(1987, 7, 9, 6, 0, 0, 0, time.UTC),
					End:   time.Date(1987, 7, 9, 9, 0, 0, 0, time.UTC),
				},
				compPeriod: models.Period{
					Begin: time.Date(1987, 7, 9, 20, 0, 0, 0, time.UTC),
					End:   time.Date(1987, 7, 9, 23, 59, 59, 0, time.UTC),
				},
				timesheet: []string{"10:00-18:00"},
			},
			want: 0,
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
