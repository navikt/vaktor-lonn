package calculator

import (
	"encoding/json"
	"fmt"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"testing"
)

func TestDøgnvakt(t *testing.T) {
	type args struct {
		report  *models.Report
		minutes map[string]models.GuardDuty
		salary  int
	}
	pocPeriode := map[string][]models.Period{
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
	}
	minWinTid := map[string][]string{
		"14.03.2022": {"07:00-15:00"},
		"15.03.2022": {"07:00-16:00"},
		// TODO: tiden 01-03 er vakt, så man skal ha tillegg for det. Må finne ut hvordan dette representeres i MinWinTid.
		"16.03.2022": { /*"01:00-03:00", */ "07:00-15:00"},
		"17.03.2022": {"08:00-16:00"},
		"18.03.2022": {"07:00-16:00"},
		"19.03.2022": {},
		"20.03.2022": {},
	}
	tests := struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
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
		want: 15_739.036,
	}

	tests.args.report.Salary = float64(tests.args.salary)
	for day, work := range minWinTid {
		timesheet := models.Timesheet{
			Schedule: pocPeriode[day],
			Work:     work,
		}
		tests.args.report.TimesheetEachDay[day] = timesheet
	}
	minutes, _ := ParsePeriode(tests.args.report, pocPeriode, minWinTid)
	tests.args.minutes = minutes

	t.Run(tests.name, func(t *testing.T) {
		err := CalculateEarnings(tests.args.report, tests.args.minutes, tests.args.salary)
		if (err != nil) != tests.wantErr {
			t.Errorf("CalculateEarnings() error = %v, wantErr %v", err, tests.wantErr)
			return
		}
		bytes, err := json.Marshal(tests.args.report)
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Println(string(bytes))
		if tests.args.report.Earnings.Total != tests.want {
			t.Errorf("CalculateEarnings() got = %v, want %v", tests.args.report.Earnings.Total, tests.want)
		}
	})
}
