package service

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/navikt/vaktor-lonn/pkg/models"
	gensql "github.com/navikt/vaktor-lonn/pkg/sql/gen"
	"github.com/navikt/vaktor-lonn/pkg/testhelpers"
)

func Test_formatTimesheet(t *testing.T) {
	type args struct {
		days []models.MWTDag
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]models.TimeSheet
		wantErr bool
	}{
		{
			name: "arbeidsdag med litt fravær",
			args: args{
				days: []models.MWTDag{
					testhelpers.MWTDag("2022-08-02", 7, "Heltid 0800-1500 (2018)", 3, "Virkedag", 500_000,
						testhelpers.MWTStempling("2022-08-02T07:45:10", "Inn", "B1", 0),
						testhelpers.MWTStempling("2022-08-02T14:30:11", "Ut", "B2", 0),
						testhelpers.MWTStempling("2022-08-02T14:31:00", "Inn", "B1", 0),
						testhelpers.MWTStempling("2022-08-02T14:31:01", "Ut på fravær", "B5", 940),
						testhelpers.MWTStempling("2022-08-02T16:00:00", "Inn fra fravær", "B4", 940),
						testhelpers.MWTStempling("2022-08-02T16:00:01", "Ut", "B2", 0),
					),
				},
			},
			want: map[string]models.TimeSheet{
				"2022-08-02": testhelpers.WorkDay("2022-08-02", 7, "Virkedag", "Heltid 0800-1500 (2018)", testhelpers.Salary500k,
					testhelpers.Clocking("2022-08-02T07:45:10", "2022-08-02T14:30:11"),
					testhelpers.Clocking("2022-08-02T14:31:00", "2022-08-02T14:31:01"),
					testhelpers.Clocking("2022-08-02T16:00:00", "2022-08-02T16:00:01"),
				),
			},
		},
		{
			name: "arbeidsdag med to fravær",
			args: args{
				days: []models.MWTDag{
					testhelpers.MWTDag("2023-06-07", 7, "Heltid 0800-1500 (2018)", 3, "Virkedag", 500_000,
						testhelpers.MWTStempling("2023-06-07T08:29:46", "Inn", "B1", 0),
						testhelpers.MWTStempling("2023-06-07T11:43:10", "Ut på fravær", "B5", 180),
						testhelpers.MWTStempling("2023-06-07T12:39:09", "Inn fra fravær", "B4", 180),
						testhelpers.MWTStempling("2023-06-07T12:40:00", "Ut på fravær", "B5", 920),
						testhelpers.MWTStempling("2023-06-07T13:30:00", "Inn fra fravær", "B4", 920),
						testhelpers.MWTStempling("2023-06-07T15:11:43", "Ut", "B2", 0),
					),
				},
			},
			want: map[string]models.TimeSheet{
				"2023-06-07": testhelpers.WorkDay("2023-06-07", 7, "Virkedag", "Heltid 0800-1500 (2018)", testhelpers.Salary500k,
					testhelpers.Clocking("2023-06-07T08:29:46", "2023-06-07T11:43:10"),
					testhelpers.Clocking("2023-06-07T12:39:09", "2023-06-07T12:40:00"),
					testhelpers.Clocking("2023-06-07T13:30:00", "2023-06-07T15:11:43"),
				),
			},
		},
		{
			name: "kveld med utrykning (glemt BV begrunnelse)",
			args: args{
				days: []models.MWTDag{
					testhelpers.MWTDag("2023-02-14", 7.45, "Heltid 0800-1545 (2018)", 5, "Virkedag", 500_000,
						testhelpers.MWTStempling("2023-02-14T08:00:00", "Inn", "B1", 0),
						testhelpers.MWTStempling("2023-02-14T15:45:00", "Ut", "B2", 0),
						testhelpers.MWTStempling("2023-02-14T20:30:00", "Inn", "B1", 0),
						testhelpers.MWTStemplingBV("2023-02-14T22:29:59", "Overtid", "B6", 0, ""),
						testhelpers.MWTStempling("2023-02-14T22:30:00", "Ut", "B2", 0),
					),
				},
			},
			want: map[string]models.TimeSheet{
				"2023-02-14": testhelpers.WorkDay("2023-02-14", 7.45, "Virkedag", "Heltid 0800-1545 (2018)", testhelpers.Salary500k,
					testhelpers.Clocking("2023-02-14T08:00:00", "2023-02-14T15:45:00"),
					testhelpers.Clocking("2023-02-14T20:30:00", "2023-02-14T22:30:00"),
				),
			},
		},
		{
			name: "helg med utrykning (liten og stor BV begrunnelse)",
			args: args{
				days: []models.MWTDag{
					testhelpers.MWTDag("2023-02-04", 0, "BV Lørdag IKT", 5, "Lørdag", 500_000,
						testhelpers.MWTStempling("2023-02-04T20:30:00", "Inn", "B1", 0),
						testhelpers.MWTStemplingBV("2023-02-04T22:29:59", "Overtid", "B6", 0, "BV"),
						testhelpers.MWTStempling("2023-02-04T22:30:00", "Ut", "B2", 0),
					),
					testhelpers.MWTDag("2023-02-11", 0, "BV Lørdag IKT", 5, "Lørdag", 500_000,
						testhelpers.MWTStempling("2023-02-11T20:30:00", "Inn", "B1", 0),
						testhelpers.MWTStemplingBV("2023-02-11T22:29:59", "Overtid", "B6", 0, "bv"),
						testhelpers.MWTStempling("2023-02-11T22:30:00", "Ut", "B2", 0),
					),
				},
			},
			want: map[string]models.TimeSheet{
				"2023-02-04": {
					Date:       testhelpers.D("2023-02-04"),
					WorkingDay: "Lørdag",
					FormName:   "BV Lørdag IKT",
					Salary:     testhelpers.Salary500k,
					Clockings:  []models.Clocking{testhelpers.ClockingOtG("2023-02-04T20:30:00", "2023-02-04T22:30:00")},
				},
				"2023-02-11": {
					Date:       testhelpers.D("2023-02-11"),
					WorkingDay: "Lørdag",
					FormName:   "BV Lørdag IKT",
					Salary:     testhelpers.Salary500k,
					Clockings:  []models.Clocking{testhelpers.ClockingOtG("2023-02-11T20:30:00", "2023-02-11T22:30:00")},
				},
			},
		},
		{
			name: "helg med utrykning (før krav om BV begrunnelse)",
			args: args{
				days: []models.MWTDag{
					testhelpers.MWTDag("2022-09-17", 0, "BV Lørdag IKT", 5, "Lørdag", 500_000,
						testhelpers.MWTStempling("2022-09-17T20:30:00", "Inn", "B1", 0),
						testhelpers.MWTStemplingBV("2022-09-17T22:29:59", "Overtid", "B6", 0, ""),
						testhelpers.MWTStempling("2022-09-17T22:30:00", "Ut", "B2", 0),
					),
					testhelpers.MWTDag("2022-09-24", 0, "BV Lørdag IKT", 5, "Lørdag", 500_000,
						testhelpers.MWTStempling("2022-09-24T20:30:00", "Inn", "B1", 0),
						testhelpers.MWTStemplingBV("2022-09-24T22:29:59", "Overtid", "B6", 0, ""),
						testhelpers.MWTStempling("2022-09-24T22:30:00", "Ut", "B2", 0),
					),
				},
			},
			want: map[string]models.TimeSheet{
				"2022-09-17": {
					Date:       testhelpers.D("2022-09-17"),
					WorkingDay: "Lørdag",
					FormName:   "BV Lørdag IKT",
					Salary:     testhelpers.Salary500k,
					Clockings:  []models.Clocking{testhelpers.ClockingOtG("2022-09-17T20:30:00", "2022-09-17T22:30:00")},
				},
				"2022-09-24": {
					Date:       testhelpers.D("2022-09-24"),
					WorkingDay: "Lørdag",
					FormName:   "BV Lørdag IKT",
					Salary:     testhelpers.Salary500k,
					Clockings:  []models.Clocking{testhelpers.ClockingOtG("2022-09-24T20:30:00", "2022-09-24T22:30:00")},
				},
			},
		},
		{
			name: "Heldags Kurs/Seminar",
			args: args{
				days: []models.MWTDag{
					testhelpers.MWTDag("2022-05-03", 7.75, "Heltid 0800-1545 (2018)", 3, "Virkedag", 500_000,
						testhelpers.MWTStempling("2022-05-03T08:00:00", "Inn", "B1", 0),
						testhelpers.MWTStempling("2022-05-03T08:00:01", "Ut på fravær", "B5", 740),
					),
				},
			},
			want: map[string]models.TimeSheet{
				"2022-05-03": testhelpers.WorkDay("2022-05-03", 7.75, "Virkedag", "Heltid 0800-1545 (2018)", testhelpers.Salary500k,
					testhelpers.Clocking("2022-05-03T08:00:00", "2022-05-03T15:45:00"),
				),
			},
		},
		{
			name: "Heldags fravær",
			args: args{
				days: []models.MWTDag{
					testhelpers.MWTDag("2022-10-17", 7.75, "Heltid 0800-1545 (2018)", 3, "Virkedag", 500_000,
						testhelpers.MWTStempling("2022-10-17T08:00:00", "Inn", "B1", 0),
						testhelpers.MWTStempling("2022-10-17T08:00:01", "Ut på fravær", "B5", 630),
						testhelpers.MWTStempling("2022-10-17T15:45:00", "Inn fra fravær", "B4", 630),
						testhelpers.MWTStempling("2022-10-17T15:45:01", "Ut", "B2", 0),
					),
				},
			},
			want: map[string]models.TimeSheet{
				"2022-10-17": testhelpers.WorkDay("2022-10-17", 7.75, "Virkedag", "Heltid 0800-1545 (2018)", testhelpers.Salary500k,
					testhelpers.Clocking("2022-10-17T08:00:00", "2022-10-17T15:45:00"),
				),
			},
		},
		{
			name: "Inn fra fravær",
			args: args{
				days: []models.MWTDag{
					testhelpers.MWTDag("2022-10-20", 7.75, "Heltid 0800-1545 (2018)", 3, "Virkedag", 500_000,
						testhelpers.MWTStempling("2022-10-20T11:12:10", "Inn fra fravær", "B4", 630),
						testhelpers.MWTStempling("2022-10-20T16:00:00", "Ut", "B2", 0),
					),
				},
			},
			want: map[string]models.TimeSheet{
				"2022-10-20": testhelpers.WorkDay("2022-10-20", 7.75, "Virkedag", "Heltid 0800-1545 (2018)", testhelpers.Salary500k,
					testhelpers.Clocking("2022-10-20T11:12:10", "2022-10-20T16:00:00"),
				),
			},
		},
		{
			name: "To overtid på natten",
			args: args{
				days: []models.MWTDag{
					testhelpers.MWTDag("2022-09-15", 7.75, "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", 3, "Virkedag", 500_000,
						testhelpers.MWTStempling("2022-09-15T00:34:21", "Inn", "B1", 0),
						testhelpers.MWTStemplingBV("2022-09-15T00:34:24", "Overtid", "B6", 0, "BV - IKT-478705 DVH"),
						testhelpers.MWTStempling("2022-09-15T01:34:42", "Ut", "B2", 0),
						testhelpers.MWTStempling("2022-09-15T03:10:00", "Inn", "B1", 0),
						testhelpers.MWTStemplingBV("2022-09-15T03:31:00", "Overtid", "B6", 0, "BV"),
						testhelpers.MWTStempling("2022-09-15T04:32:00", "Ut", "B2", 0),
						testhelpers.MWTStempling("2022-09-15T08:04:00", "Inn", "B1", 0),
						testhelpers.MWTStempling("2022-09-15T16:26:00", "Ut", "B2", 0),
					),
				},
			},
			want: map[string]models.TimeSheet{
				"2022-09-15": testhelpers.WorkDay("2022-09-15", 7.75, "Virkedag", "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary500k,
					testhelpers.ClockingOtG("2022-09-15T00:34:21", "2022-09-15T01:34:42"),
					testhelpers.ClockingOtG("2022-09-15T03:10:00", "2022-09-15T04:32:00"),
					testhelpers.Clocking("2022-09-15T08:04:00", "2022-09-15T16:26:00"),
				),
			},
		},
		{
			name: "Overtid over midnatt, ikke vakt dagen etterpå",
			args: args{
				days: []models.MWTDag{
					testhelpers.MWTDag("2022-09-15", 7.75, "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", 3, "Virkedag", 500_000,
						testhelpers.MWTStempling("2022-09-15T08:04:00", "Inn", "B1", 0),
						testhelpers.MWTStempling("2022-09-15T16:26:00", "Ut", "B2", 0),
						testhelpers.MWTStempling("2022-09-15T23:10:00", "Inn", "B1", 0),
						testhelpers.MWTStemplingBV("2022-09-15T23:31:00", "Overtid", "B6", 0, "BV"),
						testhelpers.MWTStempling("2022-09-16T00:32:00", "Ut", "B2", 0),
					),
				},
			},
			want: map[string]models.TimeSheet{
				"2022-09-15": testhelpers.WorkDay("2022-09-15", 7.75, "Virkedag", "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary500k,
					testhelpers.Clocking("2022-09-15T08:04:00", "2022-09-15T16:26:00"),
					testhelpers.ClockingOtG("2022-09-15T23:10:00", "2022-09-16T00:00:00"),
				),
			},
		},
		{
			name: "Overtid over midnatt, med vakt påfølgende dag",
			args: args{
				days: []models.MWTDag{
					testhelpers.MWTDag("2022-09-15", 7.75, "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", 3, "Virkedag", 500_000,
						testhelpers.MWTStempling("2022-09-15T08:04:00", "Inn", "B1", 0),
						testhelpers.MWTStempling("2022-09-15T16:26:00", "Ut", "B2", 0),
						testhelpers.MWTStempling("2022-09-15T23:10:00", "Inn", "B1", 0),
						testhelpers.MWTStemplingBV("2022-09-15T23:31:00", "Overtid", "B6", 0, "BV"),
						testhelpers.MWTStempling("2022-09-16T00:32:00", "Ut", "B2", 0),
					),
					testhelpers.MWTDag("2022-09-16", 7.75, "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", 3, "Virkedag", 500_000,
						testhelpers.MWTStempling("2022-09-16T08:04:00", "Inn", "B1", 0),
						testhelpers.MWTStempling("2022-09-16T15:41:00", "Ut", "B2", 0),
					),
				},
			},
			want: map[string]models.TimeSheet{
				"2022-09-15": testhelpers.WorkDay("2022-09-15", 7.75, "Virkedag", "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary500k,
					testhelpers.Clocking("2022-09-15T08:04:00", "2022-09-15T16:26:00"),
					testhelpers.ClockingOtG("2022-09-15T23:10:00", "2022-09-16T00:00:00"),
				),
				"2022-09-16": testhelpers.WorkDay("2022-09-16", 7.75, "Virkedag", "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary500k,
					testhelpers.ClockingOtG("2022-09-16T00:00:00", "2022-09-16T00:32:00"),
					testhelpers.Clocking("2022-09-16T08:04:00", "2022-09-16T15:41:00"),
				),
			},
		},
		{
			name: "Utrykning på natten, og påfølgende kveld over midnatt, med jobb påfølgende dag",
			args: args{
				days: []models.MWTDag{
					testhelpers.MWTDag("2022-10-25", 7.75, "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", 3, "Virkedag", 500_000,
						testhelpers.MWTStempling("2022-10-25T00:34:21", "Inn", "B1", 0),
						testhelpers.MWTStemplingBV("2022-10-25T00:34:24", "Overtid", "B6", 0, "BV - IKT-478705 DVH"),
						testhelpers.MWTStempling("2022-10-25T01:34:42", "Ut", "B2", 0),
						testhelpers.MWTStempling("2022-10-25T06:34:45", "Inn", "B1", 0),
						testhelpers.MWTStempling("2022-10-25T07:18:43", "Ut", "B2", 0),
						testhelpers.MWTStempling("2022-10-25T08:47:49", "Inn", "B1", 0),
						testhelpers.MWTStempling("2022-10-25T15:48:30", "Ut", "B2", 0),
						testhelpers.MWTStempling("2022-10-25T23:31:37", "Inn", "B1", 0),
						testhelpers.MWTStemplingBV("2022-10-26T00:45:34", "Overtid", "B6", 1, "BV - Feilsøking ifbm høy load på CICSP460, IKT-479284 DVH, IKT-479282 KUHR"),
						testhelpers.MWTStempling("2022-10-26T00:45:35", "Ut", "B2", 0),
					),
					testhelpers.MWTDag("2022-10-26", 7.75, "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", 3, "Virkedag", 500_000,
						testhelpers.MWTStempling("2022-10-26T08:00:00", "Inn", "B1", 0),
						testhelpers.MWTStempling("2022-10-26T15:45:00", "Ut", "B2", 0),
					),
				},
			},
			want: map[string]models.TimeSheet{
				"2022-10-25": testhelpers.WorkDay("2022-10-25", 7.75, "Virkedag", "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary500k,
					testhelpers.ClockingOtG("2022-10-25T00:34:21", "2022-10-25T01:34:42"),
					testhelpers.Clocking("2022-10-25T06:34:45", "2022-10-25T07:18:43"),
					testhelpers.Clocking("2022-10-25T08:47:49", "2022-10-25T15:48:30"),
					testhelpers.ClockingOtG("2022-10-25T23:31:37", "2022-10-26T00:00:00"),
				),
				"2022-10-26": testhelpers.WorkDay("2022-10-26", 7.75, "Virkedag", "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary500k,
					testhelpers.ClockingOtG("2022-10-26T00:00:00", "2022-10-26T00:45:35"),
					testhelpers.Clocking("2022-10-26T08:00:00", "2022-10-26T15:45:00"),
				),
			},
		},
		{
			name: "Lang dag, overtid på kvelden, og overtid over midnatt, med kort vakt påfølgende dag",
			args: args{
				days: []models.MWTDag{
					testhelpers.MWTDag("2022-10-18", 7.75, "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", 3, "Virkedag", 500_000,
						testhelpers.MWTStempling("2022-10-18T08:30:00", "Inn", "B1", 0),
						testhelpers.MWTStempling("2022-10-18T17:00:00", "Ut", "B2", 0),
						testhelpers.MWTStempling("2022-10-18T20:00:00", "Inn", "B1", 0),
						testhelpers.MWTStemplingBV("2022-10-18T20:59:59", "Overtid", "B6", 0, "BV"),
						testhelpers.MWTStempling("2022-10-18T21:00:00", "Ut", "B2", 0),
						testhelpers.MWTStempling("2022-10-18T23:30:00", "Inn", "B1", 0),
						testhelpers.MWTStemplingBV("2022-10-19T00:29:59", "Overtid", "B6", 1, "BV"),
						testhelpers.MWTStempling("2022-10-19T00:30:00", "Ut", "B2", 0),
					),
					testhelpers.MWTDag("2022-10-19", 7.75, "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", 3, "Virkedag", 500_000,
						testhelpers.MWTStempling("2022-10-19T08:00:00", "Inn", "B1", 0),
						testhelpers.MWTStempling("2022-10-19T17:00:00", "Ut", "B2", 0),
					),
				},
			},
			want: map[string]models.TimeSheet{
				"2022-10-18": testhelpers.WorkDay("2022-10-18", 7.75, "Virkedag", "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary500k,
					testhelpers.Clocking("2022-10-18T08:30:00", "2022-10-18T17:00:00"),
					testhelpers.ClockingOtG("2022-10-18T20:00:00", "2022-10-18T21:00:00"),
					testhelpers.ClockingOtG("2022-10-18T23:30:00", "2022-10-19T00:00:00"),
				),
				"2022-10-19": testhelpers.WorkDay("2022-10-19", 7.75, "Virkedag", "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary500k,
					testhelpers.ClockingOtG("2022-10-19T00:00:00", "2022-10-19T00:30:00"),
					testhelpers.Clocking("2022-10-19T08:00:00", "2022-10-19T17:00:00"),
				),
			},
		},
		{
			name: "Kun ut på fravær",
			args: args{
				days: []models.MWTDag{
					testhelpers.MWTDag("2022-01-20", 7.75, "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", 3, "Virkedag", 500_000,
						testhelpers.MWTStempling("2022-01-20T08:09:00", "Inn", "B1", 0),
						testhelpers.MWTStempling("2022-01-20T14:34:00", "Ut på fravær", "B5", 470),
					),
				},
			},
			want: map[string]models.TimeSheet{
				"2022-01-20": testhelpers.WorkDay("2022-01-20", 7.75, "Virkedag", "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary500k,
					testhelpers.Clocking("2022-01-20T08:09:00", "2022-01-20T14:34:00"),
				),
			},
		},
		{
			name: "Ut på frævar, så inn igjen",
			args: args{
				days: []models.MWTDag{
					testhelpers.MWTDag("2022-01-24", 7.75, "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", 3, "Virkedag", 500_000,
						testhelpers.MWTStempling("2022-01-24T08:27:00", "Inn", "B1", 0),
						testhelpers.MWTStempling("2022-01-24T10:01:00", "Ut på fravær", "B5", 180),
						testhelpers.MWTStempling("2022-01-24T11:27:00", "Inn", "B1", 0),
						testhelpers.MWTStempling("2022-01-24T15:45:00", "Ut", "B2", 0),
					),
				},
			},
			want: map[string]models.TimeSheet{
				"2022-01-24": testhelpers.WorkDay("2022-01-24", 7.75, "Virkedag", "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary500k,
					testhelpers.Clocking("2022-01-24T08:27:00", "2022-01-24T10:01:00"),
					testhelpers.Clocking("2022-01-24T11:27:00", "2022-01-24T15:45:00"),
				),
			},
		},
		{
			name: "En tilfeldig døgnkontinuerlig vaktuke",
			args: args{
				days: []models.MWTDag{
					testhelpers.MWTDag("2022-10-05", 7.75, "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", 2, "Virkedag", 725_000,
						testhelpers.MWTStempling("2022-10-05T07:21:42", "Inn", "B1", 0),
						testhelpers.MWTStempling("2022-10-05T15:24:14", "Ut", "B2", 0),
					),
					testhelpers.MWTDag("2022-10-06", 7.75, "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", 2, "Virkedag", 725_000,
						testhelpers.MWTStempling("2022-10-06T07:13:24", "Inn", "B1", 0),
						testhelpers.MWTStempling("2022-10-06T15:03:51", "Ut", "B2", 0),
					),
					testhelpers.MWTDag("2022-10-07", 7.75, "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", 2, "Virkedag", 725_000,
						testhelpers.MWTStempling("2022-10-07T07:18:52", "Inn", "B1", 0),
						testhelpers.MWTStempling("2022-10-07T15:06:59", "Ut", "B2", 0),
					),
					testhelpers.MWTDag("2022-10-08", 0, "BV Lørdag IKT", 2, "Lørdag", 725_000),
					testhelpers.MWTDag("2022-10-09", 0, "BV Søndag IKT", 2, "Søndag", 725_000),
					testhelpers.MWTDag("2022-10-10", 7.75, "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", 2, "Virkedag", 725_000,
						testhelpers.MWTStempling("2022-10-10T07:18:32", "Inn", "B1", 0),
						testhelpers.MWTStempling("2022-10-10T15:25:00", "Ut", "B2", 0),
					),
					testhelpers.MWTDag("2022-10-11", 7.75, "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", 2, "Virkedag", 725_000,
						testhelpers.MWTStempling("2022-10-11T07:09:58", "Inn", "B1", 0),
						testhelpers.MWTStempling("2022-10-11T15:23:41", "Ut", "B2", 0),
					),
					testhelpers.MWTDag("2022-10-12", 7.75, "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", 2, "Virkedag", 725_000,
						testhelpers.MWTStempling("2022-10-12T08:00:00", "Inn", "B1", 0),
						testhelpers.MWTStempling("2022-10-12T09:00:00", "Ut", "B2", 0),
					),
				},
			},
			want: map[string]models.TimeSheet{
				"2022-10-05": testhelpers.WorkDay("2022-10-05", 7.75, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary725k,
					testhelpers.Clocking("2022-10-05T07:21:42", "2022-10-05T15:24:14"),
				),
				"2022-10-06": testhelpers.WorkDay("2022-10-06", 7.75, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary725k,
					testhelpers.Clocking("2022-10-06T07:13:24", "2022-10-06T15:03:51"),
				),
				"2022-10-07": testhelpers.WorkDay("2022-10-07", 7.75, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary725k,
					testhelpers.Clocking("2022-10-07T07:18:52", "2022-10-07T15:06:59"),
				),
				"2022-10-08": testhelpers.OffDay("2022-10-08", "Lørdag", "BV Lørdag IKT", testhelpers.Salary725k),
				"2022-10-09": testhelpers.OffDay("2022-10-09", "Søndag", "BV Søndag IKT", testhelpers.Salary725k),
				"2022-10-10": testhelpers.WorkDay("2022-10-10", 7.75, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary725k,
					testhelpers.Clocking("2022-10-10T07:18:32", "2022-10-10T15:25:00"),
				),
				"2022-10-11": testhelpers.WorkDay("2022-10-11", 7.75, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary725k,
					testhelpers.Clocking("2022-10-11T07:09:58", "2022-10-11T15:23:41"),
				),
				"2022-10-12": testhelpers.WorkDay("2022-10-12", 7.75, "Virkedag", "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)", testhelpers.Salary725k,
					testhelpers.Clocking("2022-10-12T08:00:00", "2022-10-12T09:00:00"),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatTimesheet(tt.args.days)
			if (err != nil) != tt.wantErr {
				t.Errorf("formatTimesheet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("formatTimesheet() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_calculateSalary(t *testing.T) {
	type args struct {
		beredskapsvakt gensql.Beredskapsvakt
	}
	type want struct {
		payroll *models.Payroll
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Dybdetest av en tilfeldig vakt",
			args: args{
				beredskapsvakt: gensql.Beredskapsvakt{
					Ident:       "a123456",
					Plan:        readPlanFile(t, "Dybdetest av en tilfeldig vakt"),
					PeriodBegin: testhelpers.At("2022-10-05T12:00"),
					PeriodEnd:   testhelpers.At("2022-10-12T12:00"),
				},
			},
			want: want{
				payroll: &models.Payroll{
					ID:           uuid.MustParse("b4ac8e53-9d64-4557-8ef8-d00774ab9c06"),
					ApproverID:   "M654321",
					ApproverName: "Kalpana, Bran",
					Artskoder: models.Artskoder{
						Morgen: models.Artskode{Sum: testhelpers.Dec(6245.84), Hours: 30},
						Kveld:  models.Artskode{Sum: testhelpers.Dec(4163.89), Hours: 20},
						Dag:    models.Artskode{Sum: testhelpers.Dec(4685.52), Hours: 31},
						Helg:   models.Artskode{Sum: testhelpers.Dec(10435.74), Hours: 48},
						Skift:  models.Artskode{Sum: testhelpers.DecI(160), Hours: 20},
					},
					Stillingskode: "258",
				},
			},
		},

		{
			name: "Vanlig ukesvakt med litt overtid",
			args: args{
				beredskapsvakt: gensql.Beredskapsvakt{
					Ident:       "a123456",
					Plan:        readPlanFile(t, "Vanlig ukesvakt med litt overtid"),
					PeriodBegin: testhelpers.At("2022-10-12T12:00"),
					PeriodEnd:   testhelpers.At("2022-10-09T12:00"),
				},
			},
			want: want{
				payroll: &models.Payroll{
					ID:           uuid.MustParse("b4ac8e53-9d64-4557-8ef8-d00774ab9c06"),
					ApproverID:   "M654321",
					ApproverName: "Kalpana, Bran",
					Artskoder: models.Artskoder{
						Morgen:    models.Artskode{Sum: testhelpers.Dec(7212.97), Hours: 30},
						Kveld:     models.Artskode{Sum: testhelpers.Dec(4808.65), Hours: 20},
						Dag:       models.Artskode{Sum: testhelpers.Dec(4909.08), Hours: 28},
						Helg:      models.Artskode{Sum: testhelpers.Dec(11983.16), Hours: 48},
						Skift:     models.Artskode{Sum: testhelpers.DecI(152), Hours: 19},
						Utrykning: models.Artskode{Sum: testhelpers.DecI(0), Hours: 0},
					},
					Stillingskode: "258",
				},
			},
		},

		{
			name: "Vakt skal deles ved månedsskifte",
			args: args{
				beredskapsvakt: gensql.Beredskapsvakt{
					Ident:       "a123456",
					Plan:        readPlanFile(t, "Vakt skal deles ved månedsskifte"),
					PeriodBegin: testhelpers.At("2022-10-26T12:00"),
					PeriodEnd:   testhelpers.At("2022-11-01T00:00"),
				},
			},
			want: want{
				payroll: &models.Payroll{
					ID:           uuid.MustParse("b4ac8e53-9d64-4557-8ef8-d00774ab9c06"),
					ApproverID:   "M654321",
					ApproverName: "Kalpana, Bran",
					Artskoder: models.Artskoder{
						Morgen: models.Artskode{Sum: testhelpers.Dec(3884.11), Hours: 18},
						Kveld:  models.Artskode{Sum: testhelpers.Dec(3452.54), Hours: 16},
						Dag:    models.Artskode{Sum: testhelpers.Dec(3293.59), Hours: 21},
						Helg:   models.Artskode{Sum: testhelpers.Dec(11032.61), Hours: 49}, // stilte klokken en time tilbake denne vakten
						Skift:  models.Artskode{Sum: testhelpers.DecI(120), Hours: 15},
					},
					Stillingskode: "258",
				},
			},
		},

		{
			name: "Helg med overtid ikke merket med bv",
			args: args{
				beredskapsvakt: gensql.Beredskapsvakt{
					Ident:       "a123456",
					Plan:        readPlanFile(t, "Helg med overtid ikke merket med bv"),
					PeriodBegin: testhelpers.D("2022-10-15"),
					PeriodEnd:   testhelpers.D("2022-10-16"),
				},
			},
			want: want{
				payroll: &models.Payroll{
					ID:           uuid.MustParse("b4ac8e53-9d64-4557-8ef8-d00774ab9c06"),
					ApproverID:   "M654321",
					ApproverName: "Kalpana, Bran",
					Artskoder: models.Artskoder{
						Morgen:    models.Artskode{Sum: testhelpers.DecI(0), Hours: 0},
						Kveld:     models.Artskode{Sum: testhelpers.DecI(0), Hours: 0},
						Dag:       models.Artskode{Sum: testhelpers.DecI(0), Hours: 0},
						Helg:      models.Artskode{Sum: testhelpers.Dec(8586.31), Hours: 48},
						Utrykning: models.Artskode{Sum: testhelpers.DecI(168), Hours: 2},
					},
					Stillingskode: "265",
				},
			},
		},

		{
			name: "Ukesvakt med helg og overtid ikke merket bv",
			args: args{
				beredskapsvakt: gensql.Beredskapsvakt{
					Ident:       "a123456",
					Plan:        readPlanFile(t, "Ukesvakt med helg og overtid ikke merket bv"),
					PeriodBegin: testhelpers.D("2022-10-10"),
					PeriodEnd:   testhelpers.D("2022-10-17"),
				},
			},
			want: want{
				payroll: &models.Payroll{
					ID:           uuid.MustParse("b4ac8e53-9d64-4557-8ef8-d00774ab9c06"),
					ApproverID:   "M654321",
					ApproverName: "Kalpana, Bran",
					Artskoder: models.Artskoder{
						Morgen:    models.Artskode{Sum: testhelpers.Dec(5089.95), Hours: 30},
						Kveld:     models.Artskode{Sum: testhelpers.Dec(4071.96), Hours: 24},
						Dag:       models.Artskode{Sum: testhelpers.Dec(4278.7), Hours: 35},
						Helg:      models.Artskode{Sum: testhelpers.Dec(8586.31), Hours: 48},
						Skift:     models.Artskode{Sum: testhelpers.DecI(184), Hours: 23},
						Utrykning: models.Artskode{Sum: testhelpers.DecI(168), Hours: 2},
					},
					Stillingskode: "265",
				},
			},
		},

		{
			name: "Vakt på nyttårsaften",
			args: args{
				beredskapsvakt: gensql.Beredskapsvakt{
					Ident:       "a123456",
					Plan:        readPlanFile(t, "Vakt på nyttårsaften"),
					PeriodBegin: testhelpers.D("2022-12-31"),
					PeriodEnd:   testhelpers.D("2023-01-01"),
				},
			},
			want: want{
				payroll: &models.Payroll{
					ID:           uuid.MustParse("b4ac8e53-9d64-4557-8ef8-d00774ab9c06"),
					ApproverID:   "M654321",
					ApproverName: "Kalpana, Bran",
					Artskoder: models.Artskoder{
						Helg: models.Artskode{Sum: testhelpers.Dec(3583.79), Hours: 24},
					},
					Stillingskode: "265",
				},
			},
		},

		{
			name: "Overtid utenom beredskapsvakt",
			args: args{
				beredskapsvakt: gensql.Beredskapsvakt{
					Ident:       "a123456",
					Plan:        readPlanFile(t, "Overtid utenom beredskapsvakt"),
					PeriodBegin: testhelpers.D("2023-06-17"),
					PeriodEnd:   testhelpers.D("2023-06-18"),
				},
			},
			want: want{
				payroll: &models.Payroll{
					ID:           uuid.MustParse("b4ac8e53-9d64-4557-8ef8-d00774ab9c06"),
					ApproverID:   "M654321",
					ApproverName: "Kalpana, Bran",
					Artskoder: models.Artskoder{
						Helg:      models.Artskode{Sum: testhelpers.Dec(520.6), Hours: 12},
						Utrykning: models.Artskode{Sum: testhelpers.DecI(504), Hours: 6},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := os.Open(fmt.Sprintf("testdata/%s.json", tt.name))
			if err != nil {
				t.Errorf("failed to open file: %v", err)
				return
			}

			var body []byte
			if body, err = io.ReadAll(file); err != nil {
				t.Errorf("failed to read file: %v", err)
				return
			}

			var response models.MWTRespons
			if err := json.Unmarshal(body, &response); err != nil {
				t.Errorf("failed while unmarshling: %v", err)
				return
			}

			got, _, err := calculateSalary(tt.args.beredskapsvakt, response)
			if err != nil {
				t.Errorf("calculateSalary() returned an error: %v", err)
				return
			}

			if diff := cmp.Diff(tt.want.payroll, got); diff != "" {
				t.Errorf("calculateSalary() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func readPlanFile(t *testing.T, name string) json.RawMessage {
	t.Helper()
	body, err := os.ReadFile(fmt.Sprintf("testdata/plan/%s.json", name))
	if err != nil {
		t.Fatalf("failed to read plan file: %v", err)
	}
	return json.RawMessage(body)
}

func Test_createPerfectClocking(t *testing.T) {
	type args struct {
		tid  float64
		date time.Time
	}
	tests := []struct {
		name string
		args args
		want models.Clocking
	}{
		{
			name: "Vintertid",
			args: args{
				tid:  7.75,
				date: testhelpers.D("2022-11-06"),
			},
			want: testhelpers.Clocking("2022-11-06T08:00", "2022-11-06T15:45"),
		},
		{
			name: "Sommertid",
			args: args{
				tid:  7,
				date: testhelpers.D("2022-11-06"),
			},
			want: testhelpers.Clocking("2022-11-06T08:00", "2022-11-06T15:00"),
		},
		{
			name: "Normaltid",
			args: args{
				tid:  7.5,
				date: testhelpers.D("2022-11-06"),
			},
			want: testhelpers.Clocking("2022-11-06T08:00", "2022-11-06T15:30"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createPerfectClocking(tt.args.tid, tt.args.date)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("createPerfectClocking() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_isTimesheetApproved(t *testing.T) {
	type args struct {
		days []models.MWTDag
	}
	tests := []struct {
		name  string
		args  args
		error bool
	}{
		{
			name: "Ingen godkjente dager",
			args: args{
				days: []models.MWTDag{{Godkjent: 0}},
			},
			error: true,
		},
		{
			name: "Godkjent av vakthaver",
			args: args{
				days: []models.MWTDag{{Godkjent: 1}},
			},
			error: true,
		},
		{
			name: "Godkjent av personalleder",
			args: args{
				days: []models.MWTDag{{Godkjent: 2}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := isTimesheetApproved(tt.args.days); (err != nil) != tt.error {
				t.Errorf("isTimesheetApproved() = %v, want error %v", err, tt.error)
			}
		})
	}
}
