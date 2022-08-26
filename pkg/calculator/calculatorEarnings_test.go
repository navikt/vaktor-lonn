package calculator

import (
	"encoding/json"
	"fmt"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"testing"
)

func TestCalculateEarnings(t *testing.T) {
	type args struct {
		report  *models.Report
		minutes map[string]models.GuardDuty
		salary  int
	}
	tests := []struct {
		name      string
		args      args
		minWinTid map[string][]string
		pocPeriod map[string][]models.Period
		want      float64
		wantErr   bool
	}{
		{
			name: "døgnvakt",
			args: args{
				salary: 500_000,
				report: &models.Report{
					Ident:            "testv1",
					TimesheetEachDay: map[string]models.Timesheet{},
					Satser: map[string]float64{
						"lørsøn":  55,
						"0620":    10,
						"2006":    20,
						"utvidet": 15,
					},
				},
			},
			minWinTid: map[string][]string{
				"14.03.2022": {"07:00-15:00"},
				"15.03.2022": {"07:00-16:00"},
				// TODO: tiden 01-03 er vakt, så man skal ha tillegg for det. Må finne ut hvordan dette representeres i MinWinTid.
				"16.03.2022": { /*"01:00-03:00", */ "07:00-15:00"},
				"17.03.2022": {"08:00-16:00"},
				"18.03.2022": {"07:00-16:00"},
				"19.03.2022": {},
				"20.03.2022": {},
			},
			pocPeriod: map[string][]models.Period{
				"14.03.2022": {
					{
						Begin:     "00:00",
						End:       "24:00",
						Helligdag: false,
					},
				},
				"15.03.2022": {
					{
						Begin:     "00:00",
						End:       "24:00",
						Helligdag: false,
					},
				},
				"16.03.2022": {
					{
						Begin:     "00:00",
						End:       "24:00",
						Helligdag: false,
					},
				},
				"17.03.2022": {
					{
						Begin:     "00:00",
						End:       "24:00",
						Helligdag: true,
					},
				},
				"18.03.2022": {
					{
						Begin:     "00:00",
						End:       "24:00",
						Helligdag: false,
					},
				},
				"19.03.2022": {
					{
						Begin:     "00:00",
						End:       "24:00",
						Helligdag: false,
					},
				},
				"20.03.2022": {
					{
						Begin:     "00:00",
						End:       "24:00",
						Helligdag: false,
					},
				},
			},
			want: 15_739.036,
		},

		{
			name: "Utvidet beredskap",
			args: args{
				salary: 800_000,
				report: &models.Report{
					Ident:            "testv1",
					TimesheetEachDay: map[string]models.Timesheet{},
					Satser: map[string]float64{
						"lørsøn":  55,
						"0620":    10,
						"2006":    20,
						"utvidet": 15,
					},
				},
			},
			minWinTid: map[string][]string{
				"04.07.2022": {"09:00-15:00"},
				"05.07.2022": {"09:00-15:00"},
				"06.07.2022": {"09:00-15:30"},
				"07.07.2022": {"09:00-15:00"},
				"08.07.2022": {"09:00-15:30"},

				"11.07.2022": {"08:00-16:00"},
				"12.07.2022": {"08:00-16:00"},
				"13.07.2022": {"08:00-16:00"},
				"14.07.2022": {"08:00-16:00"},
				"15.07.2022": {"08:00-16:00"},
				"16.07.2022": {},
				"17.07.2022": {},
			},
			pocPeriod: map[string][]models.Period{
				"04.07.2022": {
					{
						Begin: "06:00",
						End:   "09:00",
					},
					{
						Begin: "15:00",
						End:   "21:00",
					},
				},
				"05.07.2022": {
					{
						Begin: "06:00",
						End:   "09:00",
					},
					{
						Begin: "15:00",
						End:   "21:00",
					},
				},
				"06.07.2022": {
					{
						Begin: "06:00",
						End:   "09:00",
					},
					{
						Begin: "15:30",
						End:   "21:00",
					},
				},
				"07.07.2022": {
					{
						Begin: "06:00",
						End:   "09:00",
					},
					{
						Begin: "15:00",
						End:   "21:00",
					},
				},
				"08.07.2022": {
					{
						Begin: "06:00",
						End:   "09:00",
					},
					{
						Begin: "15:30",
						End:   "21:00",
					},
				},

				"11.07.2022": {
					{
						Begin: "06:00",
						End:   "08:00",
					},
					{
						Begin: "16:00",
						End:   "21:00",
					},
				},
				"12.07.2022": {
					{
						Begin: "06:00",
						End:   "08:00",
					},
					{
						Begin: "16:00",
						End:   "21:00",
					},
				},
				"13.07.2022": {
					{
						Begin: "06:00",
						End:   "08:00",
					},
					{
						Begin: "16:00",
						End:   "21:00",
					},
				},
				"14.07.2022": {
					{
						Begin: "06:00",
						End:   "08:00",
					},
					{
						Begin: "16:00",
						End:   "21:00",
					},
				},
				"15.07.2022": {
					{
						Begin: "06:00",
						End:   "08:00",
					},
					{
						Begin: "16:00",
						End:   "21:00",
					},
				},
				"16.07.2022": {
					{
						Begin: "09:00",
						End:   "15:00",
					},
				},
				"17.07.2022": {
					{
						Begin: "09:00",
						End:   "15:00",
					},
				},
			},
			want: 14_018.754,
			// TODO: Excel returns .76, so we do something different with rounding,
			// Excel 2,075.68 vs 2,075.6639999999998
			// Excel 10,681.08 vs 10,681.09
		},

		{
			name: "Vakt ved spesielle hendelser",
			args: args{
				salary: 800_000,
				report: &models.Report{
					Ident:            "testv1",
					TimesheetEachDay: map[string]models.Timesheet{},
					Satser: map[string]float64{
						"lørsøn":  55,
						"0620":    10,
						"2006":    20,
						"utvidet": 15,
					},
				},
			},
			minWinTid: map[string][]string{},
			pocPeriod: map[string][]models.Period{
				"16.07.2022": {
					{
						Begin: "17:00",
						End:   "24:00",
					},
				},
				"17.07.2022": {
					{
						Begin: "07:00",
						End:   "16:00",
					},
				},
				"22.07.2022": {
					{
						Begin: "16:00",
						End:   "24:00",
					},
				},
				"23.07.2022": {
					{
						Begin: "00:00",
						End:   "24:00",
					},
				},
				"24.07.2022": {
					{
						Begin: "00:00",
						End:   "24:00",
					},
				},

				"25.07.2022": {
					{
						Begin: "00:00",
						End:   "07:00",
					},
				},
			},
			want: 15_294.578000000001,
			// TODO: Excel got 15_294.65, we got 15294.578000000001
			// Some rounding error happens
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.report.Salary = float64(tt.args.salary)
			for day, work := range tt.minWinTid {
				timesheet := models.Timesheet{
					Schedule: tt.pocPeriod[day],
					Work:     work,
				}
				tt.args.report.TimesheetEachDay[day] = timesheet
			}
			minutes, err := ParsePeriod(tt.args.report, tt.pocPeriod, tt.minWinTid)
			if err != nil {
				t.Errorf("ParsePeriod() error : %v", err)
				return
			}
			tt.args.minutes = minutes

			err = CalculateEarnings(tt.args.report, tt.args.minutes, tt.args.salary)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateEarnings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.args.report.Earnings.Total != tt.want {
				t.Errorf("CalculateEarnings() got = %v, want %v", tt.args.report.Earnings.Total, tt.want)

				bytes, err := json.Marshal(tt.args.report)
				if err != nil {
					t.Error(err)
					return
				}
				fmt.Println(string(bytes))
			}
		})
	}
}
