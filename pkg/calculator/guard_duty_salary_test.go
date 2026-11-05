package calculator

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/navikt/vaktor-lonn/pkg/testhelpers"
	"github.com/shopspring/decimal"
)

func TestGuarddutySalary(t *testing.T) {
	type args struct {
		satser      models.Satser
		timesheet   map[string]models.TimeSheet
		guardPeriod map[string][]models.Period
	}
	type want struct {
		sum     decimal.Decimal
		details *models.Artskoder
	}
	tests := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "døgnvakt",
			args: args{
				satser: testhelpers.DefaultSatser,
				timesheet: map[string]models.TimeSheet{
					"2022-03-14": testhelpers.WorkDay("2022-03-14", 7.75, "Virkedag", "", testhelpers.Salary500k,
						testhelpers.Clocking("2022-03-14T07:00", "2022-03-14T15:00")),
					"2022-03-15": testhelpers.WorkDay("2022-03-15", 7.75, "Virkedag", "", testhelpers.Salary500k,
						testhelpers.Clocking("2022-03-15T07:00", "2022-03-15T16:00")),
					"2022-03-16": testhelpers.WorkDay("2022-03-16", 7.75, "Virkedag", "", testhelpers.Salary500k,
						testhelpers.Clocking("2022-03-16T07:00", "2022-03-16T15:00")),
					"2022-03-17": testhelpers.WorkDay("2022-03-17", 7.75, "Virkedag", "", testhelpers.Salary500k,
						testhelpers.Clocking("2022-03-17T08:00", "2022-03-17T16:00")),
					"2022-03-18": testhelpers.WorkDay("2022-03-18", 7.75, "Virkedag", "", testhelpers.Salary500k,
						testhelpers.Clocking("2022-03-18T07:00", "2022-03-18T16:00")),
					"2022-03-19": testhelpers.OffDay("2022-03-19", "Lørdag", "", testhelpers.Salary500k),
					"2022-03-20": testhelpers.OffDay("2022-03-20", "Søndag", "", testhelpers.Salary500k),
				},
				guardPeriod: map[string][]models.Period{
					"2022-03-14": {testhelpers.Period("2022-03-14T00:00", "2022-03-15T00:00")},
					"2022-03-15": {testhelpers.Period("2022-03-15T00:00", "2022-03-16T00:00")},
					"2022-03-16": {testhelpers.Period("2022-03-16T00:00", "2022-03-17T00:00")},
					"2022-03-17": {testhelpers.Period("2022-03-17T00:00", "2022-03-18T00:00")},
					"2022-03-18": {testhelpers.Period("2022-03-18T00:00", "2022-03-19T00:00")},
					"2022-03-19": {testhelpers.Period("2022-03-19T00:00", "2022-03-20T00:00")},
					"2022-03-20": {testhelpers.Period("2022-03-20T00:00", "2022-03-21T00:00")},
				},
			},
			want: want{
				sum: testhelpers.Dec(16_178.86),
			},
		},

		{
			name: "døgnvakt uten stemplinger",
			args: args{
				satser: testhelpers.DefaultSatser,
				timesheet: map[string]models.TimeSheet{
					"2022-03-14": testhelpers.WorkDay("2022-03-14", 7.75, "Virkedag", "", testhelpers.Salary500k),
					"2022-03-15": testhelpers.WorkDay("2022-03-15", 7.75, "Virkedag", "", testhelpers.Salary500k),
					"2022-03-16": testhelpers.WorkDay("2022-03-16", 7.75, "Virkedag", "", testhelpers.Salary500k),
					"2022-03-17": testhelpers.WorkDay("2022-03-17", 7.75, "Virkedag", "", testhelpers.Salary500k),
					"2022-03-18": testhelpers.WorkDay("2022-03-18", 7.75, "Virkedag", "", testhelpers.Salary500k),
					"2022-03-19": testhelpers.OffDay("2022-03-19", "Lørdag", "", testhelpers.Salary500k),
					"2022-03-20": testhelpers.OffDay("2022-03-20", "Søndag", "", testhelpers.Salary500k),
				},
				guardPeriod: map[string][]models.Period{
					"2022-03-14": {testhelpers.Period("2022-03-14T00:00", "2022-03-15T00:00")},
					"2022-03-15": {testhelpers.Period("2022-03-15T00:00", "2022-03-16T00:00")},
					"2022-03-16": {testhelpers.Period("2022-03-16T00:00", "2022-03-17T00:00")},
					"2022-03-17": {testhelpers.Period("2022-03-17T00:00", "2022-03-18T00:00")},
					"2022-03-18": {testhelpers.Period("2022-03-18T00:00", "2022-03-19T00:00")},
					"2022-03-19": {testhelpers.Period("2022-03-19T00:00", "2022-03-20T00:00")},
					"2022-03-20": {testhelpers.Period("2022-03-20T00:00", "2022-03-21T00:00")},
				},
			},
			want: want{
				sum: testhelpers.Dec(16_467.1),
			},
		},

		{
			name: "døgnvakt med perfekt stempling",
			args: args{
				satser: testhelpers.DefaultSatser,
				timesheet: map[string]models.TimeSheet{
					"2022-03-14": testhelpers.WorkDay("2022-03-14", 7.75, "Virkedag", "", testhelpers.Salary500k,
						testhelpers.Clocking("2022-03-14T08:00", "2022-03-14T15:45")),
					"2022-03-15": testhelpers.WorkDay("2022-03-15", 7.75, "Virkedag", "", testhelpers.Salary500k,
						testhelpers.Clocking("2022-03-15T08:00", "2022-03-15T15:45")),
					"2022-03-16": testhelpers.WorkDay("2022-03-16", 7.75, "Virkedag", "", testhelpers.Salary500k,
						testhelpers.Clocking("2022-03-16T08:00", "2022-03-16T15:45")),
					"2022-03-17": testhelpers.WorkDay("2022-03-17", 7.75, "Virkedag", "", testhelpers.Salary500k,
						testhelpers.Clocking("2022-03-17T08:00", "2022-03-17T15:45")),
					"2022-03-18": testhelpers.WorkDay("2022-03-18", 7.75, "Virkedag", "", testhelpers.Salary500k,
						testhelpers.Clocking("2022-03-18T08:00", "2022-03-18T15:45")),
					"2022-03-19": testhelpers.OffDay("2022-03-19", "Lørdag", "", testhelpers.Salary500k),
					"2022-03-20": testhelpers.OffDay("2022-03-20", "Søndag", "", testhelpers.Salary500k),
				},
				guardPeriod: map[string][]models.Period{
					"2022-03-14": {testhelpers.Period("2022-03-14T00:00", "2022-03-15T00:00")},
					"2022-03-15": {testhelpers.Period("2022-03-15T00:00", "2022-03-16T00:00")},
					"2022-03-16": {testhelpers.Period("2022-03-16T00:00", "2022-03-17T00:00")},
					"2022-03-17": {testhelpers.Period("2022-03-17T00:00", "2022-03-18T00:00")},
					"2022-03-18": {testhelpers.Period("2022-03-18T00:00", "2022-03-19T00:00")},
					"2022-03-19": {testhelpers.Period("2022-03-19T00:00", "2022-03-20T00:00")},
					"2022-03-20": {testhelpers.Period("2022-03-20T00:00", "2022-03-21T00:00")},
				},
			},
			want: want{
				sum: testhelpers.Dec(16_467.1),
			},
		},

		{
			name: "Utvidet beredskap",
			args: args{
				satser: testhelpers.DefaultSatser,
				timesheet: map[string]models.TimeSheet{
					"2022-07-04": testhelpers.WorkDay("2022-07-04", 7.75, "Virkedag", "", testhelpers.Salary800k,
						testhelpers.Clocking("2022-07-04T09:00", "2022-07-04T15:00")),
					"2022-07-05": testhelpers.WorkDay("2022-07-05", 7.75, "Virkedag", "", testhelpers.Salary800k,
						testhelpers.Clocking("2022-07-05T09:00", "2022-07-05T15:00")),
					"2022-07-06": testhelpers.WorkDay("2022-07-06", 7.75, "Virkedag", "", testhelpers.Salary800k,
						testhelpers.Clocking("2022-07-06T09:00", "2022-07-06T15:30")),
					"2022-07-07": testhelpers.WorkDay("2022-07-07", 7.75, "Virkedag", "", testhelpers.Salary800k,
						testhelpers.Clocking("2022-07-07T09:00", "2022-07-07T15:00")),
					"2022-07-08": testhelpers.WorkDay("2022-07-08", 7.75, "Virkedag", "", testhelpers.Salary800k,
						testhelpers.Clocking("2022-07-08T09:00", "2022-07-08T15:30")),
					"2022-07-11": testhelpers.WorkDay("2022-07-11", 7.75, "Virkedag", "", testhelpers.Salary800k,
						testhelpers.Clocking("2022-07-11T08:00", "2022-07-11T16:00")),
					"2022-07-12": testhelpers.WorkDay("2022-07-12", 7.75, "Virkedag", "", testhelpers.Salary800k,
						testhelpers.Clocking("2022-07-12T08:00", "2022-07-12T16:00")),
					"2022-07-13": testhelpers.WorkDay("2022-07-13", 7.75, "Virkedag", "", testhelpers.Salary800k,
						testhelpers.Clocking("2022-07-13T08:00", "2022-07-13T16:00")),
					"2022-07-14": testhelpers.WorkDay("2022-07-14", 7.75, "Virkedag", "", testhelpers.Salary800k,
						testhelpers.Clocking("2022-07-14T08:00", "2022-07-14T16:00")),
					"2022-07-15": testhelpers.WorkDay("2022-07-15", 7.75, "Virkedag", "", testhelpers.Salary800k,
						testhelpers.Clocking("2022-07-15T08:00", "2022-07-15T16:00")),
					"2022-07-16": testhelpers.OffDay("2022-07-16", "Lørdag", "", testhelpers.Salary800k),
					"2022-07-17": testhelpers.OffDay("2022-07-17", "Søndag", "", testhelpers.Salary800k),
				},
				guardPeriod: map[string][]models.Period{
					"2022-07-04": {
						testhelpers.Period("2022-07-04T06:00", "2022-07-04T09:00"),
						testhelpers.Period("2022-07-04T15:00", "2022-07-04T21:00"),
					},
					"2022-07-05": {
						testhelpers.Period("2022-07-05T06:00", "2022-07-05T09:00"),
						testhelpers.Period("2022-07-05T15:00", "2022-07-05T21:00"),
					},
					"2022-07-06": {
						testhelpers.Period("2022-07-06T06:00", "2022-07-06T09:00"),
						testhelpers.Period("2022-07-06T15:30", "2022-07-06T21:00"),
					},
					"2022-07-07": {
						testhelpers.Period("2022-07-07T06:00", "2022-07-07T09:00"),
						testhelpers.Period("2022-07-07T15:00", "2022-07-07T21:00"),
					},
					"2022-07-08": {
						testhelpers.Period("2022-07-08T06:00", "2022-07-08T09:00"),
						testhelpers.Period("2022-07-08T15:30", "2022-07-08T21:00"),
					},
					"2022-07-11": {
						testhelpers.Period("2022-07-11T06:00", "2022-07-11T08:00"),
						testhelpers.Period("2022-07-11T16:00", "2022-07-11T21:00"),
					},
					"2022-07-12": {
						testhelpers.Period("2022-07-12T06:00", "2022-07-12T08:00"),
						testhelpers.Period("2022-07-12T16:00", "2022-07-12T21:00"),
					},
					"2022-07-13": {
						testhelpers.Period("2022-07-13T06:00", "2022-07-13T08:00"),
						testhelpers.Period("2022-07-13T16:00", "2022-07-13T21:00"),
					},
					"2022-07-14": {
						testhelpers.Period("2022-07-14T06:00", "2022-07-14T08:00"),
						testhelpers.Period("2022-07-14T16:00", "2022-07-14T21:00"),
					},
					"2022-07-15": {
						testhelpers.Period("2022-07-15T06:00", "2022-07-15T08:00"),
						testhelpers.Period("2022-07-15T16:00", "2022-07-15T21:00"),
					},
					"2022-07-16": {
						testhelpers.Period("2022-07-16T09:00", "2022-07-16T15:00"),
					},
					"2022-07-17": {
						testhelpers.Period("2022-07-17T09:00", "2022-07-17T15:00"),
					},
				},
			},
			want: want{
				sum: testhelpers.Dec(14_577.76),
			},
		},

		{
			name: "Vakt ved spesielle hendelser",
			args: args{
				satser: testhelpers.DefaultSatser,
				timesheet: map[string]models.TimeSheet{
					"2022-07-16": testhelpers.OffDay("2022-07-16", "Lørdag", "", testhelpers.Salary800k),
					"2022-07-17": testhelpers.OffDay("2022-07-17", "Søndag", "", testhelpers.Salary800k),
					"2022-07-22": testhelpers.OffDay("2022-07-22", "Virkedag", "", testhelpers.Salary800k),
					"2022-07-23": testhelpers.OffDay("2022-07-23", "Lørdag", "", testhelpers.Salary800k),
					"2022-07-24": testhelpers.OffDay("2022-07-24", "Søndag", "", testhelpers.Salary800k),
					"2022-07-25": testhelpers.OffDay("2022-07-25", "Virkedag", "", testhelpers.Salary800k),
				},
				guardPeriod: map[string][]models.Period{
					"2022-07-16": {testhelpers.Period("2022-07-16T17:00", "2022-07-17T00:00")},
					"2022-07-17": {testhelpers.Period("2022-07-17T07:00", "2022-07-17T16:00")},
					"2022-07-22": {testhelpers.Period("2022-07-22T16:00", "2022-07-23T00:00")},
					"2022-07-23": {testhelpers.Period("2022-07-23T00:00", "2022-07-24T00:00")},
					"2022-07-24": {testhelpers.Period("2022-07-24T00:00", "2022-07-25T00:00")},
					"2022-07-25": {testhelpers.Period("2022-07-25T00:00", "2022-07-25T07:00")},
				},
			},
			want: want{
				sum: testhelpers.Dec(15_825.65),
			},
		},

		{
			name: "Vakt når klokka stilles til sommertid",
			args: args{
				satser: testhelpers.DefaultSatser,
				timesheet: map[string]models.TimeSheet{
					"2022-03-27": testhelpers.OffDay("2022-03-27", "Søndag", "", testhelpers.Salary500k),
				},
				guardPeriod: map[string][]models.Period{
					"2022-03-27": {testhelpers.Period("2022-03-27T00:00", "2022-03-28T00:00")},
				},
			},
			want: want{
				sum: testhelpers.Dec(3_220.49),
			},
		},

		{
			name: "Vakt når klokka stilles til normaltid",
			args: args{
				satser: testhelpers.DefaultSatser,
				timesheet: map[string]models.TimeSheet{
					"2022-10-30": testhelpers.OffDay("2022-10-30", "Søndag", "", testhelpers.Salary500k),
				},
				guardPeriod: map[string][]models.Period{
					"2022-10-30": {testhelpers.Period("2022-10-30T00:00", "2022-10-31T00:00")},
				},
			},
			want: want{
				sum: testhelpers.Dec(3_512.70),
			},
		},

		{
			name: "Utvidet åpningstid dagen klokka stilles til normaltid",
			args: args{
				satser: testhelpers.DefaultSatser,
				timesheet: map[string]models.TimeSheet{
					"2022-10-30": testhelpers.OffDay("2022-10-30", "Søndag", "", testhelpers.Salary500k),
				},
				guardPeriod: map[string][]models.Period{
					"2022-10-30": {testhelpers.Period("2022-10-30T08:00", "2022-10-30T14:00")},
				},
			},
			want: want{
				sum: testhelpers.Dec(816.65),
			},
		},

		{
			name: "Helgevakt med utrykning",
			args: args{
				satser: testhelpers.DefaultSatser,
				timesheet: map[string]models.TimeSheet{
					"2022-09-24": {
						Date:         testhelpers.D("2022-09-24"),
						WorkingHours: 0,
						WorkingDay:   "Lørdag",
						Salary:       testhelpers.Salary500k,
						Clockings: []models.Clocking{
							testhelpers.ClockingOtG("2022-09-24T20:30", "2022-09-24T22:30"),
						},
					},
					"2022-09-25": testhelpers.OffDay("2022-09-25", "Søndag", "", testhelpers.Salary500k),
				},
				guardPeriod: map[string][]models.Period{
					"2022-09-24": {testhelpers.Period("2022-09-24T00:00", "2022-09-25T00:00")},
					"2022-09-25": {testhelpers.Period("2022-09-25T00:00", "2022-09-26T00:00")},
				},
			},
			want: want{
				sum: testhelpers.Dec(6_863.19),
			},
		},

		{
			name: "Helgevakt uten utrykning",
			args: args{
				satser: testhelpers.DefaultSatser,
				timesheet: map[string]models.TimeSheet{
					"2022-09-24": testhelpers.OffDay("2022-09-24", "Lørdag", "BV Lørdag IKT", testhelpers.Salary500k),
					"2022-09-25": testhelpers.OffDay("2022-09-25", "Søndag", "BV Søndag IKT", testhelpers.Salary500k),
				},
				guardPeriod: map[string][]models.Period{
					"2022-09-24": {testhelpers.Period("2022-09-24T00:00", "2022-09-25T00:00")},
					"2022-09-25": {testhelpers.Period("2022-09-25T00:00", "2022-09-26T00:00")},
				},
			},
			want: want{
				sum: testhelpers.Dec(6_733.19),
			},
		},

		{
			name: "vakt en dag med utrykning",
			args: args{
				satser: testhelpers.DefaultSatser,
				timesheet: map[string]models.TimeSheet{
					"2022-03-14": {
						Date:         testhelpers.D("2022-03-14"),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						Salary:       testhelpers.Salary500k,
						Clockings: []models.Clocking{
							testhelpers.Clocking("2022-03-14T07:00", "2022-03-14T15:00"),
							testhelpers.ClockingOtG("2022-03-14T20:00", "2022-03-14T22:00"),
						},
					},
				},
				guardPeriod: map[string][]models.Period{
					"2022-03-14": {testhelpers.Period("2022-03-14T00:00", "2022-03-15T00:00")},
				},
			},
			want: want{
				sum: testhelpers.Dec(1_927.57),
			},
		},

		{
			name: "En tilfeldig døgnkontinuerlig vaktuke",
			args: args{
				satser: testhelpers.DefaultSatser,
				timesheet: map[string]models.TimeSheet{
					"2022-10-05": testhelpers.WorkDay("2022-10-05", 7.75, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary725k,
						testhelpers.Clocking("2022-10-05T07:21:42", "2022-10-05T15:24:14")),
					"2022-10-06": testhelpers.WorkDay("2022-10-06", 7.75, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary725k,
						testhelpers.Clocking("2022-10-06T07:13:24", "2022-10-06T15:03:51")),
					"2022-10-07": testhelpers.WorkDay("2022-10-07", 7.75, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary725k,
						testhelpers.Clocking("2022-10-07T07:18:52", "2022-10-07T15:06:59")),
					"2022-10-08": testhelpers.OffDay("2022-10-08", "Lørdag", "BV Lørdag IKT", testhelpers.Salary725k),
					"2022-10-09": testhelpers.OffDay("2022-10-09", "Søndag", "BV Søndag IKT", testhelpers.Salary725k),
					"2022-10-10": testhelpers.WorkDay("2022-10-10", 7.75, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary725k,
						testhelpers.Clocking("2022-10-10T07:18:32", "2022-10-10T15:25:00")),
					"2022-10-11": testhelpers.WorkDay("2022-10-11", 7.75, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary725k,
						testhelpers.Clocking("2022-10-11T07:09:58", "2022-10-11T15:23:41")),
					"2022-10-12": testhelpers.WorkDay("2022-10-12", 7.75, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary725k,
						testhelpers.Clocking("2022-10-12T08:00", "2022-10-12T09:00")),
				},
				guardPeriod: map[string][]models.Period{
					"2022-10-05": {testhelpers.Period("2022-10-05T12:00", "2022-10-06T00:00")},
					"2022-10-06": {testhelpers.Period("2022-10-06T00:00", "2022-10-07T00:00")},
					"2022-10-07": {testhelpers.Period("2022-10-07T00:00", "2022-10-08T00:00")},
					"2022-10-08": {testhelpers.Period("2022-10-08T00:00", "2022-10-09T00:00")},
					"2022-10-09": {testhelpers.Period("2022-10-09T00:00", "2022-10-10T00:00")},
					"2022-10-10": {testhelpers.Period("2022-10-10T00:00", "2022-10-11T00:00")},
					"2022-10-11": {testhelpers.Period("2022-10-11T00:00", "2022-10-12T00:00")},
					"2022-10-12": {testhelpers.Period("2022-10-12T00:00", "2022-10-12T12:00")},
				},
			},
			want: want{
				sum: testhelpers.Dec(22_365.75),
			},
		},

		{
			name: "Overtid utenfor vaktperioden",
			args: args{
				satser: testhelpers.DefaultSatser,
				guardPeriod: map[string][]models.Period{
					"2023-05-12": {testhelpers.Period("2023-05-12T00:00", "2023-05-13T00:00")},
					"2023-05-13": {testhelpers.Period("2023-05-13T00:00", "2023-05-13T12:00")},
				},
				timesheet: map[string]models.TimeSheet{
					"2023-05-12": testhelpers.WorkDay("2023-05-12", 0, "Virkedag", "BV 0800-1545 m", testhelpers.Salary500k,
						testhelpers.Clocking("2023-05-12T08:00", "2023-05-12T15:45")),
					"2023-05-13": {
						Date:         testhelpers.D("2023-05-13"),
						WorkingHours: 0,
						WorkingDay:   "Lørdag",
						FormName:     "BV Lørdag IKT",
						Salary:       testhelpers.Salary500k,
						Clockings: []models.Clocking{
							testhelpers.ClockingOtG("2023-05-13T22:00", "2023-05-13T23:20"),
						},
					},
				},
			},
			want: want{
				sum: testhelpers.Dec(3_620.87),
				details: &models.Artskoder{
					Morgen: models.Artskode{Sum: testhelpers.Dec(798.65), Hours: 6},
					Kveld:  models.Artskode{Sum: testhelpers.Dec(532.43), Hours: 4},
					Dag:    models.Artskode{Sum: testhelpers.Dec(576.49), Hours: 6},
					Helg:   models.Artskode{Sum: testhelpers.Dec(1_693.3), Hours: 12},
					Skift:  models.Artskode{Sum: testhelpers.DecI(20), Hours: 4},
				},
			},
		},

		{
			name: "To vaktdager med forskjellig lønn",
			args: args{
				satser: testhelpers.DefaultSatser,
				guardPeriod: map[string][]models.Period{
					"2022-10-05": {testhelpers.Period("2022-10-05T00:00", "2022-10-06T00:00")},
					"2022-10-06": {testhelpers.Period("2022-10-06T00:00", "2022-10-07T00:00")},
				},
				timesheet: map[string]models.TimeSheet{
					"2022-10-05": testhelpers.WorkDay("2022-10-05", 7.75, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary725k,
						testhelpers.Clocking("2022-10-05T08:00", "2022-10-05T15:45")),
					"2022-10-06": testhelpers.WorkDay("2022-10-06", 7.75, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary750k,
						testhelpers.Clocking("2022-10-06T08:00", "2022-10-06T15:45")),
				},
			},
			want: want{
				sum: testhelpers.Dec(5_359.33),
				details: &models.Artskoder{
					Morgen: models.Artskode{Sum: testhelpers.Dec(2213.51), Hours: 12},
					Kveld:  models.Artskode{Sum: testhelpers.Dec(1475.68), Hours: 8},
					Dag:    models.Artskode{Sum: testhelpers.Dec(1630.14), Hours: 12},
					Skift:  models.Artskode{Sum: testhelpers.DecI(40), Hours: 8},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vaktplan := models.Vaktplan{
				ID:       uuid.UUID{},
				Ident:    "A123456",
				Schedule: tt.args.guardPeriod,
			}

			minWinTid := models.MinWinTid{
				Ident:        "A123456",
				ResourceID:   "123456",
				ApproverID:   "M654321",
				ApproverName: "Kalpana, Bran",
				Timesheet:    tt.args.timesheet,
				Satser:       tt.args.satser,
			}

			payroll, err := GuarddutySalary(vaktplan, minWinTid)
			if (err != nil) != tt.wantErr {
				t.Errorf("GuarddutySalary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			artskoder := payroll.Artskoder
			total := artskoder.Morgen.Sum.
				Add(artskoder.Kveld.Sum.
					Add(artskoder.Dag.Sum.
						Add(artskoder.Helg.Sum.
							Add(artskoder.Skift.Sum.
								Add(artskoder.Utrykning.Sum)))))

			if !total.Equal(tt.want.sum) {
				t.Errorf("GuarddutySalary() got = %v, want %v", total, tt.want.sum)
				t.Errorf("Morgen: %v, Dag: %v, Kveld: %v, Helg: %v, Skift: %v, Utrykning: %v\n", payroll.Artskoder.Morgen,
					payroll.Artskoder.Dag, payroll.Artskoder.Kveld, payroll.Artskoder.Helg, payroll.Artskoder.Skift,
					payroll.Artskoder.Utrykning)
			}

			if tt.want.details != nil {
				if diff := cmp.Diff(*tt.want.details, artskoder); diff != "" {
					t.Errorf("GuarddutySalary() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
