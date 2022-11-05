package minwintid

import (
	"github.com/google/go-cmp/cmp"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/shopspring/decimal"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func Test_formatTimesheet(t *testing.T) {
	type args struct {
		days []Dag
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
				days: []Dag{
					{
						Dato:       "2022-08-02T00:00:00",
						SkjemaTid:  7,
						SkjemaNavn: "Heltid 0800-1500 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid: "2022-08-02T07:45:10",
								Retning:      "Inn",
								Type:         "B1",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-08-02T16:00:01",
								Retning:      "Ut",
								Type:         "B2",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-08-02T16:00:00",
								Retning:      "Inn fra fravær",
								Type:         "B4",
								FravarKode:   940,
							},
							{
								StemplingTid: "2022-08-02T14:31:01",
								Retning:      "Ut på fravær",
								Type:         "B5",
								FravarKode:   940,
							},
							{
								StemplingTid: "2022-08-02T14:31:00",
								Retning:      "Inn",
								Type:         "B1",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-08-02T14:30:11",
								Retning:      "Ut",
								Type:         "B2",
								FravarKode:   0,
							},
						},
						Stillinger: []Stilling{
							{
								Koststed:  "855210",
								Formal:    "000000",
								Aktivitet: "000000",
								RATEK001:  500_000,
							},
						},
					},
				},
			},
			want: map[string]models.TimeSheet{
				"2022-08-02": {
					Date:         time.Date(2022, 8, 2, 0, 0, 0, 0, time.UTC),
					WorkingHours: 7,
					WorkingDay:   "Virkedag",
					FormName:     "Heltid 0800-1500 (2018)",
					Salary:       decimal.NewFromInt(500_000),
					Formal:       "000000",
					Koststed:     "855210",
					Aktivitet:    "000000",
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 8, 2, 7, 45, 10, 0, time.UTC),
							Out: time.Date(2022, 8, 2, 14, 30, 11, 0, time.UTC),
						},
						{
							In:  time.Date(2022, 8, 2, 14, 31, 0, 0, time.UTC),
							Out: time.Date(2022, 8, 2, 16, 0, 1, 0, time.UTC),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "helg med utrykning (liten og stor BV begrunnelse)",
			args: args{
				days: []Dag{
					{
						Dato:       "2022-09-17T00:00:00",
						SkjemaTid:  0,
						SkjemaNavn: "BV Lørdag IKT",
						Godkjent:   5,
						Virkedag:   "Lørdag",
						Stemplinger: []Stempling{
							{
								StemplingTid: "2022-09-17T20:30:00",
								Retning:      "Inn",
								Type:         "B1",
								FravarKode:   0,
							},
							{
								StemplingTid:       "2022-09-17T22:29:59",
								Retning:            "Overtid                 ",
								Type:               "B6",
								FravarKode:         0,
								OvertidBegrunnelse: "BV",
							},
							{
								StemplingTid: "2022-09-17T22:30:00",
								Retning:      "Ut",
								Type:         "B2",
								FravarKode:   0,
							},
						},
						Stillinger: []Stilling{
							{
								Koststed:  "855210",
								Formal:    "000000",
								Aktivitet: "000000",
								RATEK001:  500_000,
							},
						},
					},
					{
						Dato:       "2022-09-24T00:00:00",
						SkjemaTid:  0,
						SkjemaNavn: "BV Lørdag IKT",
						Godkjent:   5,
						Virkedag:   "Lørdag",
						Stemplinger: []Stempling{
							{
								StemplingTid: "2022-09-24T20:30:00",
								Retning:      "Inn",
								Type:         "B1",
								FravarKode:   0,
							},
							{
								StemplingTid:       "2022-09-24T22:29:59",
								Retning:            "Overtid                 ",
								Type:               "B6",
								FravarKode:         0,
								OvertidBegrunnelse: "bv",
							},
							{
								StemplingTid: "2022-09-24T22:30:00",
								Retning:      "Ut",
								Type:         "B2",
								FravarKode:   0,
							},
						},
						Stillinger: []Stilling{
							{
								Koststed:  "855210",
								Formal:    "000000",
								Aktivitet: "000000",
								RATEK001:  500_000,
							},
						},
					},
				},
			},
			want: map[string]models.TimeSheet{
				"2022-09-17": {
					Date:         time.Date(2022, 9, 17, 0, 0, 0, 0, time.UTC),
					WorkingHours: 0,
					WorkingDay:   "Lørdag",
					FormName:     "BV Lørdag IKT",
					Salary:       decimal.NewFromInt(500_000),
					Formal:       "000000",
					Koststed:     "855210",
					Aktivitet:    "000000",
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 9, 17, 20, 30, 0, 0, time.UTC),
							Out: time.Date(2022, 9, 17, 22, 30, 0, 0, time.UTC),
							OtG: true,
						},
					},
				},
				"2022-09-24": {
					Date:         time.Date(2022, 9, 24, 0, 0, 0, 0, time.UTC),
					WorkingHours: 0,
					WorkingDay:   "Lørdag",
					FormName:     "BV Lørdag IKT",
					Salary:       decimal.NewFromInt(500_000),
					Formal:       "000000",
					Koststed:     "855210",
					Aktivitet:    "000000",
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 9, 24, 20, 30, 0, 0, time.UTC),
							Out: time.Date(2022, 9, 24, 22, 30, 0, 0, time.UTC),
							OtG: true,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Heldags Kurs/Seminar",
			args: args{
				days: []Dag{
					{
						Dato:       "2022-05-03T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "Heltid 0800-1545 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid: "2022-05-03T08:00:00",
								Retning:      "Inn",
								Type:         "B1",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-05-03T08:00:01",
								Retning:      "Ut på fravær",
								Type:         "B5",
								FravarKode:   740,
							},
						},
						Stillinger: []Stilling{
							{
								Koststed:  "855210",
								Formal:    "000000",
								Aktivitet: "000000",
								RATEK001:  500_000,
							},
						},
					},
				},
			},
			want: map[string]models.TimeSheet{
				"2022-05-03": {
					Date:         time.Date(2022, 5, 3, 0, 0, 0, 0, time.UTC),
					WorkingHours: 7.75,
					WorkingDay:   "Virkedag",
					FormName:     "Heltid 0800-1545 (2018)",
					Salary:       decimal.NewFromInt(500_000),
					Formal:       "000000",
					Koststed:     "855210",
					Aktivitet:    "000000",
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 5, 3, 8, 0, 0, 0, time.UTC),
							Out: time.Date(2022, 5, 3, 15, 45, 0, 0, time.UTC),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Heldags fravær",
			args: args{
				days: []Dag{
					{
						Dato:       "2022-10-17T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "Heltid 0800-1545 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid: "2022-10-17T08:00:00",
								Retning:      "Inn",
								Type:         "B1",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-10-17T08:00:01",
								Retning:      "Ut på fravær",
								Type:         "B5",
								FravarKode:   630,
							},
							{
								StemplingTid: "2022-10-17T15:45:00",
								Retning:      "Inn fra fravær",
								Type:         "B4",
								FravarKode:   630,
							},
							{
								StemplingTid: "2022-10-17T15:45:01",
								Retning:      "Ut",
								Type:         "B2",
								FravarKode:   0,
							},
						},
						Stillinger: []Stilling{
							{
								Koststed:  "855210",
								Formal:    "000000",
								Aktivitet: "000000",
								RATEK001:  500_000,
							},
						},
					},
				},
			},
			want: map[string]models.TimeSheet{
				"2022-10-17": {
					Date:         time.Date(2022, 10, 17, 0, 0, 0, 0, time.UTC),
					WorkingHours: 7.75,
					WorkingDay:   "Virkedag",
					FormName:     "Heltid 0800-1545 (2018)",
					Salary:       decimal.NewFromInt(500_000),
					Formal:       "000000",
					Koststed:     "855210",
					Aktivitet:    "000000",
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 10, 17, 8, 0, 0, 0, time.UTC),
							Out: time.Date(2022, 10, 17, 15, 45, 0, 0, time.UTC),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Inn fra fravær",
			args: args{
				days: []Dag{
					{
						Dato:       "2022-10-20T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "Heltid 0800-1545 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid: "2022-10-20T11:12:10",
								Retning:      "Inn fra fravær",
								Type:         "B4",
								FravarKode:   630,
							},
							{
								StemplingTid: "2022-10-20T16:00:00",
								Retning:      "Ut",
								Type:         "B2",
								FravarKode:   0,
							},
						},
						Stillinger: []Stilling{
							{
								Koststed:  "855210",
								Formal:    "000000",
								Aktivitet: "000000",
								RATEK001:  500_000,
							},
						},
					},
				},
			},
			want: map[string]models.TimeSheet{
				"2022-10-20": {
					Date:         time.Date(2022, 10, 20, 0, 0, 0, 0, time.UTC),
					WorkingHours: 7.75,
					WorkingDay:   "Virkedag",
					FormName:     "Heltid 0800-1545 (2018)",
					Salary:       decimal.NewFromInt(500_000),
					Formal:       "000000",
					Koststed:     "855210",
					Aktivitet:    "000000",
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 10, 20, 11, 12, 10, 0, time.UTC),
							Out: time.Date(2022, 10, 20, 16, 0, 0, 0, time.UTC),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "To overtid på natten",
			args: args{
				days: []Dag{
					{
						Dato:       "2022-09-15T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid:       "2022-09-15T00:34:21",
								Retning:            "Inn",
								Type:               "B1",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-09-15T00:34:24",
								Retning:            "Overtid                 ",
								Type:               "B6",
								FravarKode:         0,
								OvertidBegrunnelse: "BV - IKT-478705 DVH",
							},
							{
								StemplingTid:       "2022-09-15T01:34:42",
								Retning:            "Ut",
								Type:               "B2",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid: "2022-09-15T03:10:00",
								Retning:      "Inn",
								Type:         "B1",
								FravarKode:   0,
							},
							{
								StemplingTid:       "2022-09-15T03:31:00",
								Retning:            "Overtid                 ",
								Type:               "B6",
								FravarKode:         0,
								OvertidBegrunnelse: "BV",
							},
							{
								StemplingTid: "2022-09-15T04:32:00",
								Retning:      "Ut",
								Type:         "B2",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-09-15T08:04:00",
								Retning:      "Inn",
								Type:         "B1",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-09-15T16:26:00",
								Retning:      "Ut",
								Type:         "B2",
								FravarKode:   0,
							},
						},
						Stillinger: []Stilling{
							{
								Koststed:  "855210",
								Formal:    "000000",
								Aktivitet: "000000",
								RATEK001:  500_000,
							},
						},
					},
				},
			},
			want: map[string]models.TimeSheet{
				"2022-09-15": {
					Date:         time.Date(2022, 9, 15, 0, 0, 0, 0, time.UTC),
					WorkingHours: 7.75,
					WorkingDay:   "Virkedag",
					FormName:     "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
					Salary:       decimal.NewFromInt(500_000),
					Formal:       "000000",
					Koststed:     "855210",
					Aktivitet:    "000000",
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 9, 15, 0, 34, 21, 0, time.UTC),
							Out: time.Date(2022, 9, 15, 1, 34, 42, 0, time.UTC),
							OtG: true,
						},
						{
							In:  time.Date(2022, 9, 15, 3, 10, 0, 0, time.UTC),
							Out: time.Date(2022, 9, 15, 4, 32, 0, 0, time.UTC),
							OtG: true,
						},
						{
							In:  time.Date(2022, 9, 15, 8, 4, 0, 0, time.UTC),
							Out: time.Date(2022, 9, 15, 16, 26, 0, 0, time.UTC),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Overtid over midnatt, ikke vakt dagen etterpå",
			args: args{
				days: []Dag{
					{
						Dato:       "2022-09-15T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid: "2022-09-15T08:04:00",
								Retning:      "Inn",
								Type:         "B1",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-09-15T16:26:00",
								Retning:      "Ut",
								Type:         "B2",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-09-15T23:10:00",
								Retning:      "Inn",
								Type:         "B1",
								FravarKode:   0,
							},
							{
								StemplingTid:       "2022-09-15T23:31:00",
								Retning:            "Overtid                 ",
								Type:               "B6",
								FravarKode:         0,
								OvertidBegrunnelse: "BV",
							},
							{
								StemplingTid: "2022-09-16T00:32:00",
								Retning:      "Ut",
								Type:         "B2",
								FravarKode:   0,
							},
						},
						Stillinger: []Stilling{
							{
								Koststed:  "855210",
								Formal:    "000000",
								Aktivitet: "000000",
								RATEK001:  500_000,
							},
						},
					},
				},
			},
			want: map[string]models.TimeSheet{
				"2022-09-15": {
					Date:         time.Date(2022, 9, 15, 0, 0, 0, 0, time.UTC),
					WorkingHours: 7.75,
					WorkingDay:   "Virkedag",
					FormName:     "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
					Salary:       decimal.NewFromInt(500_000),
					Formal:       "000000",
					Koststed:     "855210",
					Aktivitet:    "000000",
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 9, 15, 8, 4, 0, 0, time.UTC),
							Out: time.Date(2022, 9, 15, 16, 26, 0, 0, time.UTC),
						},
						{
							In:  time.Date(2022, 9, 15, 23, 10, 0, 0, time.UTC),
							Out: time.Date(2022, 9, 16, 0, 0, 0, 0, time.UTC),
							OtG: true,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Overtid over midnatt, med vakt påfølgende dag",
			args: args{
				days: []Dag{
					{
						Dato:       "2022-09-15T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid: "2022-09-15T08:04:00",
								Retning:      "Inn",
								Type:         "B1",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-09-15T16:26:00",
								Retning:      "Ut",
								Type:         "B2",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-09-15T23:10:00",
								Retning:      "Inn",
								Type:         "B1",
								FravarKode:   0,
							},
							{
								StemplingTid:       "2022-09-15T23:31:00",
								Retning:            "Overtid                 ",
								Type:               "B6",
								FravarKode:         0,
								OvertidBegrunnelse: "BV",
							},
							{
								StemplingTid: "2022-09-16T00:32:00",
								Retning:      "Ut",
								Type:         "B2",
								FravarKode:   0,
							},
						},
						Stillinger: []Stilling{
							{
								Koststed:  "855210",
								Formal:    "000000",
								Aktivitet: "000000",
								RATEK001:  500_000,
							},
						},
					},
					{
						Dato:       "2022-09-16T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid: "2022-09-16T08:04:00",
								Retning:      "Inn",
								Type:         "B1",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-09-16T15:41:00",
								Retning:      "Ut",
								Type:         "B2",
								FravarKode:   0,
							},
						},
						Stillinger: []Stilling{
							{
								Koststed:  "855210",
								Formal:    "000000",
								Aktivitet: "000000",
								RATEK001:  500_000,
							},
						},
					},
				},
			},
			want: map[string]models.TimeSheet{
				"2022-09-15": {
					Date:         time.Date(2022, 9, 15, 0, 0, 0, 0, time.UTC),
					WorkingHours: 7.75,
					WorkingDay:   "Virkedag",
					FormName:     "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
					Salary:       decimal.NewFromInt(500_000),
					Formal:       "000000",
					Koststed:     "855210",
					Aktivitet:    "000000",
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 9, 15, 8, 4, 0, 0, time.UTC),
							Out: time.Date(2022, 9, 15, 16, 26, 0, 0, time.UTC),
						},
						{
							In:  time.Date(2022, 9, 15, 23, 10, 0, 0, time.UTC),
							Out: time.Date(2022, 9, 16, 0, 00, 0, 0, time.UTC),
							OtG: true,
						},
					},
				},
				"2022-09-16": {
					Date:         time.Date(2022, 9, 16, 0, 0, 0, 0, time.UTC),
					WorkingHours: 7.75,
					WorkingDay:   "Virkedag",
					FormName:     "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
					Salary:       decimal.NewFromInt(500_000),
					Formal:       "000000",
					Koststed:     "855210",
					Aktivitet:    "000000",
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 9, 16, 0, 0, 0, 0, time.UTC),
							Out: time.Date(2022, 9, 16, 0, 32, 0, 0, time.UTC),
							OtG: true,
						},
						{
							In:  time.Date(2022, 9, 16, 8, 4, 0, 0, time.UTC),
							Out: time.Date(2022, 9, 16, 15, 41, 0, 0, time.UTC),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Utrykning på natten, og påfølgende kveld over midnatt, med jobb påfølgende dag",
			args: args{
				days: []Dag{
					{
						Dato:       "2022-10-25T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid:       "2022-10-25T00:34:21",
								Retning:            "Inn",
								Type:               "B1",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-25T00:34:24",
								Retning:            "Overtid                 ",
								Type:               "B6",
								FravarKode:         0,
								OvertidBegrunnelse: "BV - IKT-478705 DVH",
							},
							{
								StemplingTid:       "2022-10-25T01:34:42",
								Retning:            "Ut",
								Type:               "B2",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-25T06:34:45",
								Retning:            "Inn",
								Type:               "B1",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-25T07:18:43",
								Retning:            "Ut",
								Type:               "B2",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-25T08:47:49",
								Retning:            "Inn",
								Type:               "B1",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-25T15:48:30",
								Retning:            "Ut",
								Type:               "B2",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-25T23:31:37",
								Retning:            "Inn",
								Type:               "B1",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-26T00:45:34",
								Retning:            "Overtid                 ",
								Type:               "B6",
								FravarKode:         1,
								OvertidBegrunnelse: "BV - Feilsøking ifbm høy load på CICSP460, IKT-479284 DVH, IKT-479282 KUHR",
							},
							{
								StemplingTid:       "2022-10-26T00:45:35",
								Retning:            "Ut",
								Type:               "B2",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
						},
						Stillinger: []Stilling{
							{
								Koststed:  "855210",
								Formal:    "000000",
								Aktivitet: "000000",
								RATEK001:  500_000,
							},
						},
					},
					{
						Dato:       "2022-10-26T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid:       "2022-10-26T08:00:00",
								Retning:            "Inn",
								Type:               "B1",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-26T15:45:00",
								Retning:            "Ut",
								Type:               "B2",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
						},
						Stillinger: []Stilling{
							{
								Koststed:  "855210",
								Formal:    "000000",
								Aktivitet: "000000",
								RATEK001:  500_000,
							},
						},
					},
				},
			},
			want: map[string]models.TimeSheet{
				"2022-10-25": {
					Date:         time.Date(2022, 10, 25, 0, 0, 0, 0, time.UTC),
					WorkingHours: 7.75,
					WorkingDay:   "Virkedag",
					FormName:     "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
					Salary:       decimal.NewFromInt(500_000),
					Formal:       "000000",
					Koststed:     "855210",
					Aktivitet:    "000000",
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 10, 25, 0, 34, 21, 0, time.UTC),
							Out: time.Date(2022, 10, 25, 1, 34, 42, 0, time.UTC),
							OtG: true,
						},
						{
							In:  time.Date(2022, 10, 25, 6, 34, 45, 0, time.UTC),
							Out: time.Date(2022, 10, 25, 7, 18, 43, 0, time.UTC),
						},
						{
							In:  time.Date(2022, 10, 25, 8, 47, 49, 0, time.UTC),
							Out: time.Date(2022, 10, 25, 15, 48, 30, 0, time.UTC),
						},
						{
							In:  time.Date(2022, 10, 25, 23, 31, 37, 0, time.UTC),
							Out: time.Date(2022, 10, 26, 0, 0, 0, 0, time.UTC),
							OtG: true,
						},
					},
				},
				"2022-10-26": {
					Date:         time.Date(2022, 10, 26, 0, 0, 0, 0, time.UTC),
					WorkingHours: 7.75,
					WorkingDay:   "Virkedag",
					FormName:     "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
					Salary:       decimal.NewFromInt(500_000),
					Formal:       "000000",
					Koststed:     "855210",
					Aktivitet:    "000000",
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 10, 26, 0, 0, 0, 0, time.UTC),
							Out: time.Date(2022, 10, 26, 0, 45, 35, 0, time.UTC),
							OtG: true,
						},
						{
							In:  time.Date(2022, 10, 26, 8, 0, 0, 0, time.UTC),
							Out: time.Date(2022, 10, 26, 15, 45, 0, 0, time.UTC),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Kun ut på fravær",
			args: args{
				days: []Dag{
					{
						Dato:       "2022-01-20T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid: "2022-01-20T08:09:00",
								Retning:      "Inn",
								Type:         "B1",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-01-20T14:34:00",
								Retning:      "Ut på fravær",
								Type:         "B5",
								FravarKode:   470,
							},
						},
						Stillinger: []Stilling{
							{
								Koststed:  "855210",
								Formal:    "000000",
								Aktivitet: "000000",
								RATEK001:  500_000,
							},
						},
					},
				},
			},
			want: map[string]models.TimeSheet{
				"2022-01-20": {
					Date:         time.Date(2022, 1, 20, 0, 0, 0, 0, time.UTC),
					WorkingHours: 7.75,
					WorkingDay:   "Virkedag",
					FormName:     "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
					Salary:       decimal.NewFromInt(500_000),
					Formal:       "000000",
					Koststed:     "855210",
					Aktivitet:    "000000",
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 1, 20, 8, 9, 0, 0, time.UTC),
							Out: time.Date(2022, 1, 20, 14, 34, 0, 0, time.UTC),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Ut på frævar, så inn igjen",
			args: args{
				days: []Dag{
					{
						Dato:       "2022-01-24T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid: "2022-01-24T08:27:00",
								Retning:      "Inn",
								Type:         "B1",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-01-24T10:01:00",
								Retning:      "Ut på fravær",
								Type:         "B5",
								FravarKode:   180,
							},
							{
								StemplingTid: "2022-01-24T11:27:00",
								Retning:      "Inn",
								Type:         "B1",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-01-24T15:45:00",
								Retning:      "Ut",
								Type:         "B2",
								FravarKode:   0,
							},
						},
						Stillinger: []Stilling{
							{
								Koststed:  "855210",
								Formal:    "000000",
								Aktivitet: "000000",
								RATEK001:  500_000,
							},
						},
					},
				},
			},
			want: map[string]models.TimeSheet{
				"2022-01-24": {
					Date:         time.Date(2022, 1, 24, 0, 0, 0, 0, time.UTC),
					WorkingHours: 7.75,
					WorkingDay:   "Virkedag",
					FormName:     "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
					Salary:       decimal.NewFromInt(500_000),
					Formal:       "000000",
					Koststed:     "855210",
					Aktivitet:    "000000",
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 1, 24, 8, 27, 0, 0, time.UTC),
							Out: time.Date(2022, 1, 24, 10, 01, 0, 0, time.UTC),
						},
						{
							In:  time.Date(2022, 1, 24, 11, 27, 0, 0, time.UTC),
							Out: time.Date(2022, 1, 24, 15, 45, 0, 0, time.UTC),
						},
					},
				},
			},
			wantErr: false,
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

func Test_decodeMinWinTid(t *testing.T) {
	type args struct {
		body string
	}
	tests := []struct {
		name    string
		args    args
		want    TiddataResult
		wantErr bool
	}{
		{
			name: "Ingen dager",
			args: args{
				body: `{
	"Vaktor.Vaktor_TiddataResponse": {
		"Vaktor.Vaktor_TiddataResult": [{
			"Vaktor.nav_id": "123456",
			"Vaktor.resource_id": "E123456",
			"Vaktor.leder_resource_id": "654321",
			"Vaktor.leder_nav_id": "M654321",
			"Vaktor.leder_navn": "Kalpana, Bran",
			"Vaktor.leder_epost": "Bran.Kalpana@nav.no",
			"Vaktor.dager": "[]"
		}]
	}
}`,
			},
			want: TiddataResult{
				VaktorNavId:      "123456",
				VaktorResourceId: "E123456",
				VaktorLederNavId: "M654321",
				VaktorLederNavn:  "Kalpana, Bran",
				VaktorDager:      "",
				Dager:            []Dag{},
			},
			wantErr: false,
		},
		{
			name: "feriedag",
			args: args{
				body: `{
	"Vaktor.Vaktor_TiddataResponse": {
		"Vaktor.Vaktor_TiddataResult": [{
			"Vaktor.nav_id": "123456",
			"Vaktor.resource_id": "E123456",
			"Vaktor.leder_resource_id": "654321",
			"Vaktor.leder_nav_id": "M654321",
			"Vaktor.leder_navn": "Kalpana, Bran",
			"Vaktor.leder_epost": "Bran.Kalpana@nav.no",
			"Vaktor.dager": "[{\"dato\":\"2022-07-15T00:00:00\",\"skjema_tid\":7,\"skjema_navn\":\"Heltid 0800-1500 (2018)\",\"godkjent\":3,\"ansatt_dato_godkjent_av\":\"m158267\",\"godkjent_dato\":\"2022-08-01T13:17:21\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-07-15T15:00:00\",\"Navn\":\"Inn fra fravær\",\"Type\":\"B4\",\"Fravar_kode\":210,\"Fravar_kode_navn\":\"06 Ferie\"},{\"Stempling_Tid\":\"2022-07-15T15:00:01\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"}],\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"855210\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"500000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"AK_UNIO\",\"TILL_LONN\":\"0\",\"RATE_K001\":500000,\"RATE_I143\":0,\"RATE_B100\":500000,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"BDM855210\",\"user_id\":\"117355\",\"domain_info\":\"ADEO\\\\E117355\"},{\"role_id\":\"BDM855210\",\"user_id\":\"159373\",\"domain_info\":\"ADEO\\\\S159373\"},{\"role_id\":\"BDM855210\",\"user_id\":\"153469\",\"domain_info\":\"ADEO\\\\J153469\"},{\"role_id\":\"BDM855210\",\"user_id\":\"153411\",\"domain_info\":\"ADEO\\\\B153411\"},{\"role_id\":\"BDM855210\",\"user_id\":\"158267\",\"domain_info\":\"ADEO\\\\M158267\"},{\"role_id\":\"BDM855210\",\"user_id\":\"158266\",\"domain_info\":\"ADEO\\\\B158266\"},{\"role_id\":\"BDM855210\",\"user_id\":\"142793\",\"domain_info\":\"ADEO\\\\M142793\"},{\"role_id\":\"BDM855210\",\"user_id\":\"166502\",\"domain_info\":\"ADEO\\\\L166502\"},{\"role_id\":\"BDM855210\",\"user_id\":\"166501\",\"domain_info\":\"ADEO\\\\H166501\"},{\"role_id\":\"BDM855210\",\"user_id\":\"166743\",\"domain_info\":\"ADEO\\\\H166743\"},{\"role_id\":\"BDM855210\",\"user_id\":\"116514\",\"domain_info\":\"ADEO\\\\A116514\"}],\"BDM_FORMAL\":null}]}]"
		}]
	}
}`,
			},
			want: TiddataResult{
				VaktorNavId:      "123456",
				VaktorResourceId: "E123456",
				VaktorLederNavId: "M654321",
				VaktorLederNavn:  "Kalpana, Bran",
				VaktorDager:      "",
				Dager: []Dag{
					{
						Dato:       "2022-07-15T00:00:00",
						SkjemaTid:  7,
						SkjemaNavn: "Heltid 0800-1500 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid: "2022-07-15T15:00:00",
								Retning:      "Inn fra fravær",
								Type:         "B4",
								FravarKode:   210,
							},
							{
								StemplingTid: "2022-07-15T15:00:01",
								Retning:      "Ut",
								Type:         "B2",
								FravarKode:   0,
							},
						},
						Stillinger: []Stilling{
							{
								Koststed:  "855210",
								Formal:    "000000",
								Aktivitet: "000000",
								RATEK001:  500_000,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "En dag med litt fravær",
			args: args{
				body: `{
	"Vaktor.Vaktor_TiddataResponse": {
		"Vaktor.Vaktor_TiddataResult": [{
			"Vaktor.nav_id": "123456",
			"Vaktor.resource_id": "E123456",
			"Vaktor.leder_resource_id": "654321",
			"Vaktor.leder_nav_id": "M654321",
			"Vaktor.leder_navn": "Kalpana, Bran",
			"Vaktor.leder_epost": "Bran.Kalpana@nav.no",
			"Vaktor.dager": "[{\"dato\":\"2022-08-02T00:00:00\",\"skjema_tid\":7,\"skjema_navn\":\"Heltid 0800-1500 (2018)\",\"godkjent\":5,\"ansatt_dato_godkjent_av\":\"m158267\",\"godkjent_dato\":\"2022-09-01T10:32:41\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-08-02T07:45:10\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"},{\"Stempling_Tid\":\"2022-08-02T16:00:01\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"},{\"Stempling_Tid\":\"2022-08-02T16:00:00\",\"Navn\":\"Inn fra fravær\",\"Type\":\"B4\",\"Fravar_kode\":940,\"Fravar_kode_navn\":\"36 Annet fravær med lønn\"},{\"Stempling_Tid\":\"2022-08-02T14:31:01\",\"Navn\":\"Ut på fravær\",\"Type\":\"B5\",\"Fravar_kode\":940,\"Fravar_kode_navn\":\"36 Annet fravær med lønn\"},{\"Stempling_Tid\":\"2022-08-02T14:31:00\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"},{\"Stempling_Tid\":\"2022-08-02T14:30:11\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"}],\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"855210\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"500000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"AK_UNIO\",\"TILL_LONN\":\"0\",\"RATE_K001\":500000,\"RATE_I143\":0,\"RATE_B100\":500000,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"BDM855210\",\"user_id\":\"117355\",\"domain_info\":\"ADEO\\\\E117355\"},{\"role_id\":\"BDM855210\",\"user_id\":\"159373\",\"domain_info\":\"ADEO\\\\S159373\"},{\"role_id\":\"BDM855210\",\"user_id\":\"153469\",\"domain_info\":\"ADEO\\\\J153469\"},{\"role_id\":\"BDM855210\",\"user_id\":\"153411\",\"domain_info\":\"ADEO\\\\B153411\"},{\"role_id\":\"BDM855210\",\"user_id\":\"158267\",\"domain_info\":\"ADEO\\\\M158267\"},{\"role_id\":\"BDM855210\",\"user_id\":\"158266\",\"domain_info\":\"ADEO\\\\B158266\"},{\"role_id\":\"BDM855210\",\"user_id\":\"142793\",\"domain_info\":\"ADEO\\\\M142793\"},{\"role_id\":\"BDM855210\",\"user_id\":\"166502\",\"domain_info\":\"ADEO\\\\L166502\"},{\"role_id\":\"BDM855210\",\"user_id\":\"166501\",\"domain_info\":\"ADEO\\\\H166501\"},{\"role_id\":\"BDM855210\",\"user_id\":\"166743\",\"domain_info\":\"ADEO\\\\H166743\"},{\"role_id\":\"BDM855210\",\"user_id\":\"116514\",\"domain_info\":\"ADEO\\\\A116514\"}],\"BDM_FORMAL\":null}]}]"
		}]
	}
}`,
			},
			want: TiddataResult{
				VaktorNavId:      "123456",
				VaktorResourceId: "E123456",
				VaktorLederNavId: "M654321",
				VaktorLederNavn:  "Kalpana, Bran",
				VaktorDager:      "",
				Dager: []Dag{
					{
						Dato:       "2022-08-02T00:00:00",
						SkjemaTid:  7,
						SkjemaNavn: "Heltid 0800-1500 (2018)",
						Godkjent:   5,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid: "2022-08-02T07:45:10",
								Retning:      "Inn",
								Type:         "B1",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-08-02T16:00:01",
								Retning:      "Ut",
								Type:         "B2",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-08-02T16:00:00",
								Retning:      "Inn fra fravær",
								Type:         "B4",
								FravarKode:   940,
							},
							{
								StemplingTid: "2022-08-02T14:31:01",
								Retning:      "Ut på fravær",
								Type:         "B5",
								FravarKode:   940,
							},
							{
								StemplingTid: "2022-08-02T14:31:00",
								Retning:      "Inn",
								Type:         "B1",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-08-02T14:30:11",
								Retning:      "Ut",
								Type:         "B2",
								FravarKode:   0,
							},
						},
						Stillinger: []Stilling{
							{
								Koststed:  "855210",
								Formal:    "000000",
								Aktivitet: "000000",
								RATEK001:  500_000,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "helg med utrykning",
			args: args{
				body: `{
	"Vaktor.Vaktor_TiddataResponse": {
		"Vaktor.Vaktor_TiddataResult": [{
			"Vaktor.nav_id": "123456",
			"Vaktor.resource_id": "E123456",
			"Vaktor.leder_resource_id": "654321",
			"Vaktor.leder_nav_id": "M654321",
			"Vaktor.leder_navn": "Kalpana, Bran",
			"Vaktor.leder_epost": "Bran.Kalpana@nav.no",
			"Vaktor.dager": "[{\"dato\":\"2022-09-24T00:00:00\",\"skjema_tid\":0,\"skjema_navn\":\"BV Lørdag IKT\",\"godkjent\":5,\"ansatt_dato_godkjent_av\":\"m158267\",\"godkjent_dato\":\"2022-10-04T10:50:25\",\"virkedag\":\"Lørdag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-09-24T20:30:00\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"},{\"Stempling_Tid\":\"2022-09-24T22:29:59\",\"Navn\":\"Overtid                 \",\"Type\":\"B6\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"},{\"Stempling_Tid\":\"2022-09-24T22:30:00\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"}],\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"855210\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"500000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"AK_UNIO\",\"TILL_LONN\":\"0\",\"RATE_K001\":500000,\"RATE_I143\":0,\"RATE_B100\":500000,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"BDM855210\",\"user_id\":\"117355\",\"domain_info\":\"ADEO\\\\E117355\"},{\"role_id\":\"BDM855210\",\"user_id\":\"159373\",\"domain_info\":\"ADEO\\\\S159373\"},{\"role_id\":\"BDM855210\",\"user_id\":\"153469\",\"domain_info\":\"ADEO\\\\J153469\"},{\"role_id\":\"BDM855210\",\"user_id\":\"153411\",\"domain_info\":\"ADEO\\\\B153411\"},{\"role_id\":\"BDM855210\",\"user_id\":\"158267\",\"domain_info\":\"ADEO\\\\M158267\"},{\"role_id\":\"BDM855210\",\"user_id\":\"158266\",\"domain_info\":\"ADEO\\\\B158266\"},{\"role_id\":\"BDM855210\",\"user_id\":\"142793\",\"domain_info\":\"ADEO\\\\M142793\"},{\"role_id\":\"BDM855210\",\"user_id\":\"166502\",\"domain_info\":\"ADEO\\\\L166502\"},{\"role_id\":\"BDM855210\",\"user_id\":\"166501\",\"domain_info\":\"ADEO\\\\H166501\"},{\"role_id\":\"BDM855210\",\"user_id\":\"166743\",\"domain_info\":\"ADEO\\\\H166743\"},{\"role_id\":\"BDM855210\",\"user_id\":\"116514\",\"domain_info\":\"ADEO\\\\A116514\"}],\"BDM_FORMAL\":null}]}]"
		}]
	}
}`,
			},
			want: TiddataResult{
				VaktorNavId:      "123456",
				VaktorResourceId: "E123456",
				VaktorLederNavId: "M654321",
				VaktorLederNavn:  "Kalpana, Bran",
				VaktorDager:      "",
				Dager: []Dag{
					{
						Dato:       "2022-09-24T00:00:00",
						SkjemaTid:  0,
						SkjemaNavn: "BV Lørdag IKT",
						Godkjent:   5,
						Virkedag:   "Lørdag",
						Stemplinger: []Stempling{
							{
								StemplingTid: "2022-09-24T20:30:00",
								Retning:      "Inn",
								Type:         "B1",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-09-24T22:29:59",
								Retning:      "Overtid                 ",
								Type:         "B6",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-09-24T22:30:00",
								Retning:      "Ut",
								Type:         "B2",
								FravarKode:   0,
							},
						},
						Stillinger: []Stilling{
							{
								Koststed:  "855210",
								Formal:    "000000",
								Aktivitet: "000000",
								RATEK001:  500_000,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Heldags Kurs/Seminar",
			args: args{
				body: `{
	"Vaktor.Vaktor_TiddataResponse": {
		"Vaktor.Vaktor_TiddataResult": [{
			"Vaktor.nav_id": "123456",
			"Vaktor.resource_id": "E123456",
			"Vaktor.leder_resource_id": "654321",
			"Vaktor.leder_nav_id": "M654321",
			"Vaktor.leder_navn": "Kalpana, Bran",
			"Vaktor.leder_epost": "Bran.Kalpana@nav.no",
			"Vaktor.dager": "[{\"dato\":\"2022-05-03T00:00:00\",\"skjema_tid\":7.75,\"skjema_navn\":\"Heltid 0800-1545 (2018)\",\"godkjent\":3,\"ansatt_dato_godkjent_av\":\"m158267\",\"godkjent_dato\":\"2022-06-01T09:49:19\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-05-03T08:00:00\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"},{\"Stempling_Tid\":\"2022-05-03T08:00:01\",\"Navn\":\"Ut på fravær\",\"Type\":\"B5\",\"Fravar_kode\":740,\"Fravar_kode_navn\":\"02 Kurs\/Seminar\"}],\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"855210\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"500000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"AK_UNIO\",\"TILL_LONN\":\"0\",\"RATE_K001\":500000,\"RATE_I143\":0,\"RATE_B100\":500000,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"BDM855210\",\"user_id\":\"117355\",\"domain_info\":\"ADEO\\\\E117355\"},{\"role_id\":\"BDM855210\",\"user_id\":\"159373\",\"domain_info\":\"ADEO\\\\S159373\"},{\"role_id\":\"BDM855210\",\"user_id\":\"153469\",\"domain_info\":\"ADEO\\\\J153469\"},{\"role_id\":\"BDM855210\",\"user_id\":\"153411\",\"domain_info\":\"ADEO\\\\B153411\"},{\"role_id\":\"BDM855210\",\"user_id\":\"158267\",\"domain_info\":\"ADEO\\\\M158267\"},{\"role_id\":\"BDM855210\",\"user_id\":\"158266\",\"domain_info\":\"ADEO\\\\B158266\"},{\"role_id\":\"BDM855210\",\"user_id\":\"142793\",\"domain_info\":\"ADEO\\\\M142793\"},{\"role_id\":\"BDM855210\",\"user_id\":\"166502\",\"domain_info\":\"ADEO\\\\L166502\"},{\"role_id\":\"BDM855210\",\"user_id\":\"166501\",\"domain_info\":\"ADEO\\\\H166501\"},{\"role_id\":\"BDM855210\",\"user_id\":\"166743\",\"domain_info\":\"ADEO\\\\H166743\"},{\"role_id\":\"BDM855210\",\"user_id\":\"116514\",\"domain_info\":\"ADEO\\\\A116514\"}],\"BDM_FORMAL\":null}]}]"
		}]
	}
}`,
			},
			want: TiddataResult{
				VaktorNavId:      "123456",
				VaktorResourceId: "E123456",
				VaktorLederNavId: "M654321",
				VaktorLederNavn:  "Kalpana, Bran",
				VaktorDager:      "",
				Dager: []Dag{
					{
						Dato:       "2022-05-03T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "Heltid 0800-1545 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid: "2022-05-03T08:00:00",
								Retning:      "Inn",
								Type:         "B1",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-05-03T08:00:01",
								Retning:      "Ut på fravær",
								Type:         "B5",
								FravarKode:   740,
							},
						},
						Stillinger: []Stilling{
							{
								Koststed:  "855210",
								Formal:    "000000",
								Aktivitet: "000000",
								RATEK001:  500_000,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Heldags Kurs/Seminar",
			args: args{
				body: `{
	"Vaktor.Vaktor_TiddataResponse": {
		"Vaktor.Vaktor_TiddataResult": [{
			"Vaktor.nav_id": "123456",
			"Vaktor.resource_id": "E123456",
			"Vaktor.leder_resource_id": "654321",
			"Vaktor.leder_nav_id": "M654321",
			"Vaktor.leder_navn": "Kalpana, Bran",
			"Vaktor.leder_epost": "Bran.Kalpana@nav.no",
			"Vaktor.dager": "[{\"dato\":\"2022-09-15T00:00:00\",\"skjema_tid\":7.75,\"skjema_navn\":\"NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":3,\"ansatt_dato_godkjent_av\":\"n155953\",\"godkjent_dato\":\"2022-10-03T12:28:15\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-09-15T08:04:08\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"},{\"Stempling_Tid\":\"2022-09-15T16:26:15\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"},{\"Stempling_Tid\":\"2022-09-15T23:10:27\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"},{\"Stempling_Tid\":\"2022-09-15T23:31:51\",\"Navn\":\"Overtid                 \",\"Type\":\"B6\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"},{\"Stempling_Tid\":\"2022-09-16T00:32:24\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"}],\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"koststed\":\"855210\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"80\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"LO_YS\",\"TILL_LONN\":\"0\",\"RATE_K001\":500000,\"RATE_I143\":0,\"RATE_B100\":500000,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"BDM855120\",\"user_id\":\"140650\",\"domain_info\":\"ADEO\\\\B140650\"},{\"role_id\":\"BDM855120\",\"user_id\":\"155953\",\"domain_info\":\"ADEO\\\\N155953\"}],\"BDM_FORMAL\":null}]}]"
		}]
	}
}`,
			},
			want: TiddataResult{
				VaktorNavId:      "123456",
				VaktorResourceId: "E123456",
				VaktorLederNavId: "M654321",
				VaktorLederNavn:  "Kalpana, Bran",
				VaktorDager:      "",
				Dager: []Dag{
					{
						Dato:       "2022-09-15T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid: "2022-09-15T08:04:08",
								Retning:      "Inn",
								Type:         "B1",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-09-15T16:26:15",
								Retning:      "Ut",
								Type:         "B2",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-09-15T23:10:27",
								Retning:      "Inn",
								Type:         "B1",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-09-15T23:31:51",
								Retning:      "Overtid                 ",
								Type:         "B6",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-09-16T00:32:24",
								Retning:      "Ut",
								Type:         "B2",
								FravarKode:   0,
							},
						},
						Stillinger: []Stilling{
							{
								Koststed:  "855210",
								Formal:    "000000",
								Aktivitet: "000000",
								RATEK001:  500_000,
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				Body: io.NopCloser(strings.NewReader(tt.args.body)),
			}

			got, err := decodeMinWinTid(resp)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeMinWinTid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("decodeMinWinTid() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
