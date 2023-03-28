package minwintid

import (
	"encoding/json"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/navikt/vaktor-lonn/pkg/models"
	gensql "github.com/navikt/vaktor-lonn/pkg/sql/gen"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
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
			name: "Lang dag, overtid på kvelden, og overtid over midnatt, med kort vakt påfølgende dag",
			args: args{
				days: []Dag{
					{
						Dato:       "2022-10-18T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid: "2022-10-18T08:30:00",
								Retning:      "Inn",
								Type:         "B1",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-10-18T17:00:00",
								Retning:      "Ut",
								Type:         "B2",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-10-18T20:00:00",
								Retning:      "Inn",
								Type:         "B1",
								FravarKode:   0,
							},
							{
								StemplingTid:       "2022-10-18T20:59:59",
								Retning:            "Overtid                 ",
								Type:               "B6",
								FravarKode:         0,
								OvertidBegrunnelse: "BV",
							},
							{
								StemplingTid: "2022-10-18T21:00:00",
								Retning:      "Ut",
								Type:         "B2",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-10-18T23:30:00",
								Retning:      "Inn",
								Type:         "B1",
								FravarKode:   0,
							},
							{
								StemplingTid:       "2022-10-19T00:29:59",
								Retning:            "Overtid                 ",
								Type:               "B6",
								FravarKode:         1,
								OvertidBegrunnelse: "BV",
							},
							{
								StemplingTid: "2022-10-19T00:30:00",
								Retning:      "Ut",
								Type:         "B2",
								FravarKode:   0,
							},
						},
						Stillinger: []Stilling{
							{
								RATEK001:  500_000,
							},
						},
					},
					{
						Dato:       "2022-10-19T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid: "2022-10-19T08:00:00",
								Retning:      "Inn",
								Type:         "B1",
								FravarKode:   0,
							},
							{
								StemplingTid: "2022-10-19T17:00:00",
								Retning:      "Ut",
								Type:         "B2",
								FravarKode:   0,
							},
						},
						Stillinger: []Stilling{
							{
								RATEK001:  500_000,
							},
						},
					},
				},
			},
			want: map[string]models.TimeSheet{
				"2022-10-18": {
					Date:         time.Date(2022, 10, 18, 0, 0, 0, 0, time.UTC),
					WorkingHours: 7.75,
					WorkingDay:   "Virkedag",
					FormName:     "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
					Salary:       decimal.NewFromInt(500_000),
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 10, 18, 8, 30, 0, 0, time.UTC),
							Out: time.Date(2022, 10, 18, 17, 0, 0, 0, time.UTC),
						},
						{
							In:  time.Date(2022, 10, 18, 20, 0, 0, 0, time.UTC),
							Out: time.Date(2022, 10, 18, 21, 0, 0, 0, time.UTC),
							OtG: true,
						},
						{
							In:  time.Date(2022, 10, 18, 23, 30, 0, 0, time.UTC),
							Out: time.Date(2022, 10, 19, 0, 0, 0, 0, time.UTC),
							OtG: true,
						},
					},
				},
				"2022-10-19": {
					Date:         time.Date(2022, 10, 19, 0, 0, 0, 0, time.UTC),
					WorkingHours: 7.75,
					WorkingDay:   "Virkedag",
					FormName:     "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
					Salary:       decimal.NewFromInt(500_000),
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 10, 19, 0, 0, 0, 0, time.UTC),
							Out: time.Date(2022, 10, 19, 0, 30, 0, 0, time.UTC),
							OtG: true,
						},
						{
							In:  time.Date(2022, 10, 19, 8, 0, 0, 0, time.UTC),
							Out: time.Date(2022, 10, 19, 17, 0, 0, 0, time.UTC),
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
		{
			name: "En tilfeldig døgnkontinuerlig vaktuke",
			args: args{
				days: []Dag{
					{
						Dato:       "2022-10-05T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   2,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid:       "2022-10-05T07:21:42",
								Retning:            "Inn",
								Type:               "B1",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-05T15:24:14",
								Retning:            "Ut",
								Type:               "B2",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
						},
						Stillinger: []Stilling{
							{
								RATEK001:  725000,
							},
						},
					},
					{
						Dato:       "2022-10-06T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   2,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid:       "2022-10-06T07:13:24",
								Retning:            "Inn",
								Type:               "B1",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-06T15:03:51",
								Retning:            "Ut",
								Type:               "B2",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
						},
						Stillinger: []Stilling{
							{
								RATEK001:  725000,
							},
						},
					},
					{
						Dato:       "2022-10-07T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   2,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid:       "2022-10-07T07:18:52",
								Retning:            "Inn",
								Type:               "B1",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-07T15:06:59",
								Retning:            "Ut",
								Type:               "B2",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
						},
						Stillinger: []Stilling{
							{
								RATEK001:  725000,
							},
						},
					},
					{
						Dato:        "2022-10-08T00:00:00",
						SkjemaTid:   0,
						SkjemaNavn:  "BV Lørdag IKT",
						Godkjent:    2,
						Virkedag:    "Lørdag",
						Stemplinger: nil,
						Stillinger: []Stilling{
							{
								RATEK001:  725000,
							},
						},
					},
					{
						Dato:        "2022-10-09T00:00:00",
						SkjemaTid:   0,
						SkjemaNavn:  "BV Søndag IKT",
						Godkjent:    2,
						Virkedag:    "Søndag",
						Stemplinger: nil,
						Stillinger: []Stilling{
							{
								RATEK001:  725000,
							},
						},
					},
					{
						Dato:       "2022-10-10T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   2,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid:       "2022-10-10T07:18:32",
								Retning:            "Inn",
								Type:               "B1",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-10T15:25:00",
								Retning:            "Ut",
								Type:               "B2",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
						},
						Stillinger: []Stilling{
							{
								RATEK001:  725000,
							},
						},
					},
					{
						Dato:       "2022-10-11T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   2,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid:       "2022-10-11T07:09:58",
								Retning:            "Inn",
								Type:               "B1",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-11T15:23:41",
								Retning:            "Ut",
								Type:               "B2",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
						},
						Stillinger: []Stilling{
							{
								RATEK001:  725000,
							},
						},
					},
					{
						Dato:       "2022-10-12T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   2,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid:       "2022-10-12T08:00:00",
								Retning:            "Inn",
								Type:               "B1",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-12T09:00:00",
								Retning:            "Ut",
								Type:               "B2",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
						},
						Stillinger: []Stilling{
							{
								RATEK001:  725000,
							},
						},
					},
				},
			},
			want: map[string]models.TimeSheet{
				"2022-10-05": {
					Date:         time.Date(2022, 10, 5, 0, 0, 0, 0, time.UTC),
					WorkingHours: 7.75,
					WorkingDay:   "Virkedag",
					FormName:     "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
					Salary:       decimal.NewFromInt(725000),
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 10, 05, 07, 21, 42, 0, time.UTC),
							Out: time.Date(2022, 10, 05, 15, 24, 14, 0, time.UTC),
						},
					},
				},
				"2022-10-06": {
					Date:         time.Date(2022, 10, 6, 0, 0, 0, 0, time.UTC),
					WorkingHours: 7.75,
					WorkingDay:   "Virkedag",
					FormName:     "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
					Salary:       decimal.NewFromInt(725000),
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 10, 06, 07, 13, 24, 0, time.UTC),
							Out: time.Date(2022, 10, 06, 15, 03, 51, 0, time.UTC),
						},
					},
				},
				"2022-10-07": {
					Date:         time.Date(2022, 10, 7, 0, 0, 0, 0, time.UTC),
					WorkingHours: 7.75,
					WorkingDay:   "Virkedag",
					FormName:     "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
					Salary:       decimal.NewFromInt(725000),
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 10, 07, 07, 18, 52, 0, time.UTC),
							Out: time.Date(2022, 10, 07, 15, 06, 59, 0, time.UTC),
						},
					},
				},
				"2022-10-08": {
					Date:       time.Date(2022, 10, 8, 0, 0, 0, 0, time.UTC),
					WorkingDay: "Lørdag",
					FormName:   "BV Lørdag IKT",
					Salary:     decimal.NewFromInt(725000),
					Clockings:  []models.Clocking{},
				},
				"2022-10-09": {
					Date:       time.Date(2022, 10, 9, 0, 0, 0, 0, time.UTC),
					WorkingDay: "Søndag",
					FormName:   "BV Søndag IKT",
					Salary:     decimal.NewFromInt(725000),
					Clockings:  []models.Clocking{},
				},
				"2022-10-10": {
					Date:         time.Date(2022, 10, 10, 0, 0, 0, 0, time.UTC),
					WorkingHours: 7.75,
					WorkingDay:   "Virkedag",
					FormName:     "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
					Salary:       decimal.NewFromInt(725000),
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 10, 10, 07, 18, 32, 0, time.UTC),
							Out: time.Date(2022, 10, 10, 15, 25, 0, 0, time.UTC),
						},
					},
				},
				"2022-10-11": {
					Date:         time.Date(2022, 10, 11, 0, 0, 0, 0, time.UTC),
					WorkingHours: 7.75,
					WorkingDay:   "Virkedag",
					FormName:     "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
					Salary:       decimal.NewFromInt(725000),
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 10, 11, 7, 9, 58, 0, time.UTC),
							Out: time.Date(2022, 10, 11, 15, 23, 41, 0, time.UTC),
						},
					},
				},
				"2022-10-12": {
					Date:         time.Date(2022, 10, 12, 0, 0, 0, 0, time.UTC),
					WorkingHours: 7.75,
					WorkingDay:   "Virkedag",
					FormName:     "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
					Salary:       decimal.NewFromInt(725000),
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 10, 12, 8, 0, 0, 0, time.UTC),
							Out: time.Date(2022, 10, 12, 9, 0, 0, 0, time.UTC),
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
			"Vaktor.dager": "[{\"dato\":\"2022-07-15T00:00:00\",\"skjema_tid\":7,\"skjema_navn\":\"Heltid 0800-1500 (2018)\",\"godkjent\":3,\"ansatt_dato_godkjent_av\":\"a123456\",\"godkjent_dato\":\"2022-08-01T13:17:21\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-07-15T15:00:00\",\"Navn\":\"Inn fra fravær\",\"Type\":\"B4\",\"Fravar_kode\":210,\"Fravar_kode_navn\":\"06 Ferie\"},{\"Stempling_Tid\":\"2022-07-15T15:00:01\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"}],\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"500000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":500000,\"RATE_I143\":0,\"RATE_B100\":500000,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]}]"
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
			"Vaktor.dager": "[{\"dato\":\"2022-08-02T00:00:00\",\"skjema_tid\":7,\"skjema_navn\":\"Heltid 0800-1500 (2018)\",\"godkjent\":5,\"ansatt_dato_godkjent_av\":\"a123456\",\"godkjent_dato\":\"2022-09-01T10:32:41\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-08-02T07:45:10\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"},{\"Stempling_Tid\":\"2022-08-02T16:00:01\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"},{\"Stempling_Tid\":\"2022-08-02T16:00:00\",\"Navn\":\"Inn fra fravær\",\"Type\":\"B4\",\"Fravar_kode\":940,\"Fravar_kode_navn\":\"36 Annet fravær med lønn\"},{\"Stempling_Tid\":\"2022-08-02T14:31:01\",\"Navn\":\"Ut på fravær\",\"Type\":\"B5\",\"Fravar_kode\":940,\"Fravar_kode_navn\":\"36 Annet fravær med lønn\"},{\"Stempling_Tid\":\"2022-08-02T14:31:00\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"},{\"Stempling_Tid\":\"2022-08-02T14:30:11\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"}],\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"500000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":500000,\"RATE_I143\":0,\"RATE_B100\":500000,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]}]"
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
			"Vaktor.dager": "[{\"dato\":\"2022-09-24T00:00:00\",\"skjema_tid\":0,\"skjema_navn\":\"BV Lørdag IKT\",\"godkjent\":5,\"ansatt_dato_godkjent_av\":\"a123456\",\"godkjent_dato\":\"2022-10-04T10:50:25\",\"virkedag\":\"Lørdag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-09-24T20:30:00\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"},{\"Stempling_Tid\":\"2022-09-24T22:29:59\",\"Navn\":\"Overtid                 \",\"Type\":\"B6\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"},{\"Stempling_Tid\":\"2022-09-24T22:30:00\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"}],\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"500000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":500000,\"RATE_I143\":0,\"RATE_B100\":500000,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]}]"
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
			"Vaktor.dager": "[{\"dato\":\"2022-05-03T00:00:00\",\"skjema_tid\":7.75,\"skjema_navn\":\"Heltid 0800-1545 (2018)\",\"godkjent\":3,\"ansatt_dato_godkjent_av\":\"a123456\",\"godkjent_dato\":\"2022-06-01T09:49:19\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-05-03T08:00:00\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"},{\"Stempling_Tid\":\"2022-05-03T08:00:01\",\"Navn\":\"Ut på fravær\",\"Type\":\"B5\",\"Fravar_kode\":740,\"Fravar_kode_navn\":\"02 Kurs\/Seminar\"}],\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"500000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":500000,\"RATE_I143\":0,\"RATE_B100\":500000,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]}]"
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
			"Vaktor.dager": "[{\"dato\":\"2022-09-15T00:00:00\",\"skjema_tid\":7.75,\"skjema_navn\":\"NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":3,\"ansatt_dato_godkjent_av\":\"a123456\",\"godkjent_dato\":\"2022-10-03T12:28:15\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-09-15T08:04:08\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"},{\"Stempling_Tid\":\"2022-09-15T16:26:15\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"},{\"Stempling_Tid\":\"2022-09-15T23:10:27\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"},{\"Stempling_Tid\":\"2022-09-15T23:31:51\",\"Navn\":\"Overtid                 \",\"Type\":\"B6\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"},{\"Stempling_Tid\":\"2022-09-16T00:32:24\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\"}],\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"80\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":500000,\"RATE_I143\":0,\"RATE_B100\":500000,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]}]"
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
								RATEK001:  500_000,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "En tilfeldig døgnkontinuerlig vaktuke",
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
			"Vaktor.dager": "[{\"dato\":\"2022-10-05T00:00:00\",\"skjema_tid\":7.75,\"skjema_navn\":\"BV 0800-1545 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":2,\"ansatt_dato_godkjent_av\":\"a123456\",\"godkjent_dato\":\"2022-11-02T10:24:21\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-10-05T07:21:42\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-05T15:24:14\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null}],\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"725000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":725000,\"RATE_I143\":0,\"RATE_B100\":725000,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-12T00:00:00\",\"skjema_tid\":7.75,\"skjema_navn\":\"BV 0800-1545 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":2,\"ansatt_dato_godkjent_av\":\"a123456\",\"godkjent_dato\":\"2022-11-05T08:36:43\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-10-12T08:00:00\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-12T09:00:00\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null}],\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"725000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":725000,\"RATE_I143\":0,\"RATE_B100\":725000,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-11T00:00:00\",\"skjema_tid\":7.75,\"skjema_navn\":\"BV 0800-1545 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":2,\"ansatt_dato_godkjent_av\":\"a123456\",\"godkjent_dato\":\"2022-11-02T10:24:21\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-10-11T07:09:58\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-11T15:23:41\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null}],\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"725000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":725000,\"RATE_I143\":0,\"RATE_B100\":725000,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-09T00:00:00\",\"skjema_tid\":0,\"skjema_navn\":\"BV Søndag IKT\",\"godkjent\":2,\"ansatt_dato_godkjent_av\":\"a123456\",\"godkjent_dato\":\"2022-11-02T10:24:21\",\"virkedag\":\"Søndag\",\"Stemplinger\":null,\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"725000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":725000,\"RATE_I143\":0,\"RATE_B100\":725000,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-07T00:00:00\",\"skjema_tid\":7.75,\"skjema_navn\":\"BV 0800-1545 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":2,\"ansatt_dato_godkjent_av\":\"a123456\",\"godkjent_dato\":\"2022-11-02T10:24:21\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-10-07T07:18:52\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-07T15:06:59\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null}],\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"725000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":725000,\"RATE_I143\":0,\"RATE_B100\":725000,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-06T00:00:00\",\"skjema_tid\":7.75,\"skjema_navn\":\"BV 0800-1545 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":2,\"ansatt_dato_godkjent_av\":\"a123456\",\"godkjent_dato\":\"2022-11-02T10:24:21\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-10-06T07:13:24\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-06T15:03:51\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null}],\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"725000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":725000,\"RATE_I143\":0,\"RATE_B100\":725000,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-10T00:00:00\",\"skjema_tid\":7.75,\"skjema_navn\":\"BV 0800-1545 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":2,\"ansatt_dato_godkjent_av\":\"a123456\",\"godkjent_dato\":\"2022-11-02T10:24:21\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-10-10T07:18:32\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-10T15:25:00\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null}],\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"725000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":725000,\"RATE_I143\":0,\"RATE_B100\":725000,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-08T00:00:00\",\"skjema_tid\":0,\"skjema_navn\":\"BV Lørdag IKT\",\"godkjent\":2,\"ansatt_dato_godkjent_av\":\"a123456\",\"godkjent_dato\":\"2022-11-02T10:24:21\",\"virkedag\":\"Lørdag\",\"Stemplinger\":null,\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"725000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":725000,\"RATE_I143\":0,\"RATE_B100\":725000,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]}]"
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
						Dato:       "2022-10-05T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   2,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid:       "2022-10-05T07:21:42",
								Retning:            "Inn",
								Type:               "B1",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-05T15:24:14",
								Retning:            "Ut",
								Type:               "B2",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
						},
						Stillinger: []Stilling{
							{
								RATEK001:  725000,
							},
						},
					},
					{
						Dato:       "2022-10-06T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   2,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid:       "2022-10-06T07:13:24",
								Retning:            "Inn",
								Type:               "B1",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-06T15:03:51",
								Retning:            "Ut",
								Type:               "B2",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
						},
						Stillinger: []Stilling{
							{
								RATEK001:  725000,
							},
						},
					},
					{
						Dato:       "2022-10-07T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   2,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid:       "2022-10-07T07:18:52",
								Retning:            "Inn",
								Type:               "B1",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-07T15:06:59",
								Retning:            "Ut",
								Type:               "B2",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
						},
						Stillinger: []Stilling{
							{
								RATEK001:  725000,
							},
						},
					},
					{
						Dato:        "2022-10-08T00:00:00",
						SkjemaTid:   0,
						SkjemaNavn:  "BV Lørdag IKT",
						Godkjent:    2,
						Virkedag:    "Lørdag",
						Stemplinger: nil,
						Stillinger: []Stilling{
							{
								RATEK001:  725000,
							},
						},
					},
					{
						Dato:        "2022-10-09T00:00:00",
						SkjemaTid:   0,
						SkjemaNavn:  "BV Søndag IKT",
						Godkjent:    2,
						Virkedag:    "Søndag",
						Stemplinger: nil,
						Stillinger: []Stilling{
							{
								RATEK001:  725000,
							},
						},
					},
					{
						Dato:       "2022-10-10T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   2,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid:       "2022-10-10T07:18:32",
								Retning:            "Inn",
								Type:               "B1",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-10T15:25:00",
								Retning:            "Ut",
								Type:               "B2",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
						},
						Stillinger: []Stilling{
							{
								RATEK001:  725000,
							},
						},
					},
					{
						Dato:       "2022-10-11T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   2,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid:       "2022-10-11T07:09:58",
								Retning:            "Inn",
								Type:               "B1",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-11T15:23:41",
								Retning:            "Ut",
								Type:               "B2",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
						},
						Stillinger: []Stilling{
							{
								RATEK001:  725000,
							},
						},
					},
					{
						Dato:       "2022-10-12T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   2,
						Virkedag:   "Virkedag",
						Stemplinger: []Stempling{
							{
								StemplingTid:       "2022-10-12T08:00:00",
								Retning:            "Inn",
								Type:               "B1",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-12T09:00:00",
								Retning:            "Ut",
								Type:               "B2",
								FravarKode:         0,
								OvertidBegrunnelse: "",
							},
						},
						Stillinger: []Stilling{
							{
								RATEK001:  725000,
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
			var response Response
			err := json.Unmarshal([]byte(tt.args.body), &response)
			if err != nil {
				t.Errorf("failed while unmarshling: %v", err)
				return
			}

			got, err := decodeMinWinTid(response)
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

func Test_calculateSalary(t *testing.T) {
	log, err := zap.NewDevelopment([]zap.Option{}...)
	if err != nil {
		t.Errorf("can't create zap.Logger: %v", err)
		return
	}

	type args struct {
		beredskapsvakt gensql.Beredskapsvakt
		body           string
	}
	type want struct {
		payroll models.Payroll
		ok      bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Dybde test av en tilfeldig vakt",
			args: args{
				beredskapsvakt: gensql.Beredskapsvakt{
					Ident:       "a123456",
					Plan:        json.RawMessage(`{"id":"b4ac8e53-9d64-4557-8ef8-d00774ab9c06","user_id":"E123456","start_timestamp":"2022-10-05T12:00:00Z","end_timestamp":"2022-10-12T12:00:00Z","schedule":{"2022-10-05":[{"start_timestamp":"2022-10-05T12:00:00Z","end_timestamp":"2022-10-06T00:00:00Z"}],"2022-10-06":[{"start_timestamp":"2022-10-06T00:00:00Z","end_timestamp":"2022-10-07T00:00:00Z"}],"2022-10-07":[{"start_timestamp":"2022-10-07T00:00:00Z","end_timestamp":"2022-10-08T00:00:00Z"}],"2022-10-08":[{"start_timestamp":"2022-10-08T00:00:00Z","end_timestamp":"2022-10-09T00:00:00Z"}],"2022-10-09":[{"start_timestamp":"2022-10-09T00:00:00Z","end_timestamp":"2022-10-10T00:00:00Z"}],"2022-10-10":[{"start_timestamp":"2022-10-10T00:00:00Z","end_timestamp":"2022-10-11T00:00:00Z"}],"2022-10-11":[{"start_timestamp":"2022-10-11T00:00:00Z","end_timestamp":"2022-10-12T00:00:00Z"}],"2022-10-12":[{"start_timestamp":"2022-10-12T00:00:00Z","end_timestamp":"2022-10-12T12:00:00Z"}]}}`),
					PeriodBegin: time.Date(2022, 10, 5, 12, 0, 0, 0, time.UTC),
					PeriodEnd:   time.Date(2022, 10, 12, 12, 0, 0, 0, time.UTC),
				},
				body: `{
	"Vaktor.Vaktor_TiddataResponse": {
		"Vaktor.Vaktor_TiddataResult": [{
			"Vaktor.nav_id": "123456",
			"Vaktor.resource_id": "E123456",
			"Vaktor.leder_resource_id": "654321",
			"Vaktor.leder_nav_id": "M654321",
			"Vaktor.leder_navn": "Kalpana, Bran",
			"Vaktor.leder_epost": "Bran.Kalpana@nav.no",
			"Vaktor.dager": "[{\"dato\":\"2022-10-05T00:00:00\",\"skjema_tid\":7.75,\"skjema_navn\":\"BV 0800-1545 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":4,\"ansatt_dato_godkjent_av\":\"M654321\",\"godkjent_dato\":\"2022-11-02T08:47:09\",\"virkedag\":\"Virkedag\",\"Stemplinger\":null,\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"79\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":814900,\"RATE_I143\":0,\"RATE_B100\":814900,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-12T00:00:00\",\"skjema_tid\":7.75,\"skjema_navn\":\"BV 0800-1545 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":4,\"ansatt_dato_godkjent_av\":\"M654321\",\"godkjent_dato\":\"2022-11-02T08:47:09\",\"virkedag\":\"Virkedag\",\"Stemplinger\":null,\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"79\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":814900,\"RATE_I143\":0,\"RATE_B100\":814900,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-11T00:00:00\",\"skjema_tid\":7.75,\"skjema_navn\":\"BV 0800-1545 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":4,\"ansatt_dato_godkjent_av\":\"M654321\",\"godkjent_dato\":\"2022-11-02T08:47:09\",\"virkedag\":\"Virkedag\",\"Stemplinger\":null,\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"79\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":814900,\"RATE_I143\":0,\"RATE_B100\":814900,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-10T00:00:00\",\"skjema_tid\":7.75,\"skjema_navn\":\"BV 0800-1545 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":4,\"ansatt_dato_godkjent_av\":\"M654321\",\"godkjent_dato\":\"2022-11-02T08:47:09\",\"virkedag\":\"Virkedag\",\"Stemplinger\":null,\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"79\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":814900,\"RATE_I143\":0,\"RATE_B100\":814900,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-09T00:00:00\",\"skjema_tid\":0,\"skjema_navn\":\"BV Søndag IKT\",\"godkjent\":4,\"ansatt_dato_godkjent_av\":\"M654321\",\"godkjent_dato\":\"2022-11-02T08:47:09\",\"virkedag\":\"Søndag\",\"Stemplinger\":null,\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"79\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":814900,\"RATE_I143\":0,\"RATE_B100\":814900,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-08T00:00:00\",\"skjema_tid\":0,\"skjema_navn\":\"BV Lørdag IKT\",\"godkjent\":4,\"ansatt_dato_godkjent_av\":\"M654321\",\"godkjent_dato\":\"2022-11-02T08:47:09\",\"virkedag\":\"Lørdag\",\"Stemplinger\":null,\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"79\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":814900,\"RATE_I143\":0,\"RATE_B100\":814900,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-06T00:00:00\",\"skjema_tid\":7.75,\"skjema_navn\":\"BV 0800-1545 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":4,\"ansatt_dato_godkjent_av\":\"M654321\",\"godkjent_dato\":\"2022-11-02T08:47:09\",\"virkedag\":\"Virkedag\",\"Stemplinger\":null,\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"79\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":814900,\"RATE_I143\":0,\"RATE_B100\":814900,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-07T00:00:00\",\"skjema_tid\":7.75,\"skjema_navn\":\"BV 0800-1545 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":4,\"ansatt_dato_godkjent_av\":\"M654321\",\"godkjent_dato\":\"2022-11-02T08:47:09\",\"virkedag\":\"Virkedag\",\"Stemplinger\":null,\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"79\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":814900,\"RATE_I143\":0,\"RATE_B100\":814900,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"user_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]}]"
		}]
	}
}`,
			},
			want: want{
				payroll: models.Payroll{
					ID:           uuid.MustParse("b4ac8e53-9d64-4557-8ef8-d00774ab9c06"),
					ApproverID:   "M654321",
					ApproverName: "Kalpana, Bran",
					Artskoder: models.Artskoder{
						Morgen: models.Artskode{
							Sum:   decimal.NewFromFloat(6035.84),
							Hours: 30,
						},
						Kveld: models.Artskode{
							Sum:   decimal.NewFromFloat(4023.89),
							Hours: 20,
						},
						Dag: models.Artskode{
							Sum:   decimal.NewFromFloat(4561.52),
							Hours: 31,
						},
						Helg: models.Artskode{
							Sum:   decimal.NewFromFloat(10001.34),
							Hours: 48,
						},
						Skift: models.Artskode{
							Sum:   decimal.NewFromFloat(100),
							Hours: 20,
						},
					},
				},
				ok: false,
			},
		},

		{
			name: "vanlig ukesvakt med litt overtid",
			args: args{
				beredskapsvakt: gensql.Beredskapsvakt{
					Ident:       "a123456",
					Plan:        json.RawMessage(`{"id":"b4ac8e53-9d64-4557-8ef8-d00774ab9c06","user_id":"E123456","start_timestamp":"2022-10-12T12:00:00Z","end_timestamp":"2022-10-19T12:00:00Z","schedule":{"2022-10-12":[{"start_timestamp":"2022-10-12T12:00:00Z","end_timestamp":"2022-10-13T00:00:00Z"}],"2022-10-13":[{"start_timestamp":"2022-10-13T00:00:00Z","end_timestamp":"2022-10-14T00:00:00Z"}],"2022-10-14":[{"start_timestamp":"2022-10-14T00:00:00Z","end_timestamp":"2022-10-15T00:00:00Z"}],"2022-10-15":[{"start_timestamp":"2022-10-15T00:00:00Z","end_timestamp":"2022-10-16T00:00:00Z"}],"2022-10-16":[{"start_timestamp":"2022-10-16T00:00:00Z","end_timestamp":"2022-10-17T00:00:00Z"}],"2022-10-17":[{"start_timestamp":"2022-10-17T00:00:00Z","end_timestamp":"2022-10-18T00:00:00Z"}],"2022-10-18":[{"start_timestamp":"2022-10-18T00:00:00Z","end_timestamp":"2022-10-19T00:00:00Z"}],"2022-10-19":[{"start_timestamp":"2022-10-19T00:00:00Z","end_timestamp":"2022-10-19T12:00:00Z"}]}}`),
					PeriodBegin: time.Date(2022, 10, 12, 12, 0, 0, 0, time.UTC),
					PeriodEnd:   time.Date(2022, 10, 9, 12, 0, 0, 0, time.UTC),
				},
				body: `{
	"Vaktor.Vaktor_TiddataResponse": {
		"Vaktor.Vaktor_TiddataResult": [{
			"Vaktor.nav_id": "123456",
			"Vaktor.resource_id": "E123456",
			"Vaktor.leder_resource_id": "654321",
			"Vaktor.leder_nav_id": "M654321",
			"Vaktor.leder_navn": "Kalpana, Bran",
			"Vaktor.leder_epost": "Bran.Kalpana@nav.no",
			"Vaktor.dager": "[{\"dato\":\"2022-10-12T00:00:00\",\"skjema_tid\":7.75,\"skjema_navn\":\"BV 0800-1545 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":2,\"ansatt_dato_godkjent_av\":\"b140650\",\"godkjent_dato\":\"2022-11-03T17:12:18\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-10-12T08:00:00\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-12T16:00:00\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null}],\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"964000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"AK_UNIO\",\"TILL_LONN\":\"0\",\"RATE_K001\":964000,\"RATE_I143\":0,\"RATE_B100\":964000,\"RATE_K170\":35,\"RATE_K171\":15,\"RATE_K172\":25,\"RATE_K160\":25,\"RATE_K161\":65,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-19T00:00:00\",\"skjema_tid\":7.75,\"skjema_navn\":\"BV 0800-1545 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":2,\"ansatt_dato_godkjent_av\":\"b140650\",\"godkjent_dato\":\"2022-11-03T17:12:18\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-10-19T08:00:00\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-19T17:00:00\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null}],\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"964000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"AK_UNIO\",\"TILL_LONN\":\"0\",\"RATE_K001\":964000,\"RATE_I143\":0,\"RATE_B100\":964000,\"RATE_K170\":35,\"RATE_K171\":15,\"RATE_K172\":25,\"RATE_K160\":25,\"RATE_K161\":65,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-18T00:00:00\",\"skjema_tid\":7.75,\"skjema_navn\":\"BV 0800-1545 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":2,\"ansatt_dato_godkjent_av\":\"b140650\",\"godkjent_dato\":\"2022-11-03T17:12:18\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-10-18T08:30:00\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-19T00:29:59\",\"Navn\":\"Overtid                 \",\"Type\":\"B6\",\"Fravar_kode\":1,\"Fravar_kode_navn\":\"Inne\",\"Overtid_Begrunnelse\":\"Oppringt vakt, feilsøk, BV\"},{\"Stempling_Tid\":\"2022-10-18T20:59:59\",\"Navn\":\"Overtid                 \",\"Type\":\"B6\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":\"Endring i prod, BV\"},{\"Stempling_Tid\":\"2022-10-19T00:30:00\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-18T23:30:00\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-18T21:00:00\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-18T20:00:00\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-18T17:00:00\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null}],\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"964000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"AK_UNIO\",\"TILL_LONN\":\"0\",\"RATE_K001\":964000,\"RATE_I143\":0,\"RATE_B100\":964000,\"RATE_K170\":35,\"RATE_K171\":15,\"RATE_K172\":25,\"RATE_K160\":25,\"RATE_K161\":65,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-17T00:00:00\",\"skjema_tid\":7.75,\"skjema_navn\":\"BV 0800-1545 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":2,\"ansatt_dato_godkjent_av\":\"b140650\",\"godkjent_dato\":\"2022-11-03T17:12:18\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-10-17T08:45:00\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-17T16:30:00\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null}],\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"964000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"AK_UNIO\",\"TILL_LONN\":\"0\",\"RATE_K001\":964000,\"RATE_I143\":0,\"RATE_B100\":964000,\"RATE_K170\":35,\"RATE_K171\":15,\"RATE_K172\":25,\"RATE_K160\":25,\"RATE_K161\":65,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-16T00:00:00\",\"skjema_tid\":0,\"skjema_navn\":\"BV Søndag IKT\",\"godkjent\":2,\"ansatt_dato_godkjent_av\":\"b140650\",\"godkjent_dato\":\"2022-11-03T17:12:18\",\"virkedag\":\"Søndag\",\"Stemplinger\":null,\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"964000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"AK_UNIO\",\"TILL_LONN\":\"0\",\"RATE_K001\":964000,\"RATE_I143\":0,\"RATE_B100\":964000,\"RATE_K170\":35,\"RATE_K171\":15,\"RATE_K172\":25,\"RATE_K160\":25,\"RATE_K161\":65,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-15T00:00:00\",\"skjema_tid\":0,\"skjema_navn\":\"BV Lørdag IKT\",\"godkjent\":2,\"ansatt_dato_godkjent_av\":\"b140650\",\"godkjent_dato\":\"2022-11-03T17:12:18\",\"virkedag\":\"Lørdag\",\"Stemplinger\":null,\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"964000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"AK_UNIO\",\"TILL_LONN\":\"0\",\"RATE_K001\":964000,\"RATE_I143\":0,\"RATE_B100\":964000,\"RATE_K170\":35,\"RATE_K171\":15,\"RATE_K172\":25,\"RATE_K160\":25,\"RATE_K161\":65,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-14T00:00:00\",\"skjema_tid\":7.75,\"skjema_navn\":\"BV 0800-1545 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":2,\"ansatt_dato_godkjent_av\":\"b140650\",\"godkjent_dato\":\"2022-11-03T17:12:18\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-10-14T08:00:00\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-14T14:00:00\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null}],\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"964000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"AK_UNIO\",\"TILL_LONN\":\"0\",\"RATE_K001\":964000,\"RATE_I143\":0,\"RATE_B100\":964000,\"RATE_K170\":35,\"RATE_K171\":15,\"RATE_K172\":25,\"RATE_K160\":25,\"RATE_K161\":65,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-13T00:00:00\",\"skjema_tid\":7.75,\"skjema_navn\":\"BV 0800-1545 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":2,\"ansatt_dato_godkjent_av\":\"b140650\",\"godkjent_dato\":\"2022-11-03T17:12:18\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-10-13T07:45:00\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-13T18:00:00\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null}],\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"964000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"AK_UNIO\",\"TILL_LONN\":\"0\",\"RATE_K001\":964000,\"RATE_I143\":0,\"RATE_B100\":964000,\"RATE_K170\":35,\"RATE_K171\":15,\"RATE_K172\":25,\"RATE_K160\":25,\"RATE_K161\":65,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]}]"
		}]
	}
}`,
			},
			want: want{
				payroll: models.Payroll{
					ID:           uuid.MustParse("b4ac8e53-9d64-4557-8ef8-d00774ab9c06"),
					ApproverID:   "M654321",
					ApproverName: "Kalpana, Bran",
					Artskoder: models.Artskoder{
						Morgen: models.Artskode{
							Sum:   decimal.NewFromFloat(7002.97),
							Hours: 30,
						},
						Kveld: models.Artskode{
							Sum:   decimal.NewFromFloat(4435.22),
							Hours: 19,
						},
						Dag: models.Artskode{
							Sum:   decimal.NewFromFloat(4797.08),
							Hours: 28,
						},
						Helg: models.Artskode{
							Sum:   decimal.NewFromFloat(11548.76),
							Hours: 48,
						},
						Skift: models.Artskode{
							Sum:   decimal.NewFromFloat(95),
							Hours: 19,
						},
						Utrykning: models.Artskode{
							Sum:   decimal.NewFromFloat(75),
							Hours: 3,
						},
					},
				},
				ok: false,
			},
		},

		{
			name: "delt vakt i månedsskifte",
			args: args{
				beredskapsvakt: gensql.Beredskapsvakt{
					Ident:       "a123456",
					Plan:        json.RawMessage(`{"id":"b4ac8e53-9d64-4557-8ef8-d00774ab9c06","user_id":"E123456","start_timestamp":"2022-10-26T12:00:00Z","end_timestamp":"2022-11-01T00:00:00Z","schedule":{"2022-10-26":[{"start_timestamp":"2022-10-26T12:00:00Z","end_timestamp":"2022-10-27T00:00:00Z"}],"2022-10-27":[{"start_timestamp":"2022-10-27T00:00:00Z","end_timestamp":"2022-10-28T00:00:00Z"}],"2022-10-28":[{"start_timestamp":"2022-10-28T00:00:00Z","end_timestamp":"2022-10-29T00:00:00Z"}],"2022-10-29":[{"start_timestamp":"2022-10-29T00:00:00Z","end_timestamp":"2022-10-30T00:00:00Z"}],"2022-10-30":[{"start_timestamp":"2022-10-30T00:00:00Z","end_timestamp":"2022-10-31T00:00:00Z"}],"2022-10-31":[{"start_timestamp":"2022-10-31T00:00:00Z","end_timestamp":"2022-11-01T00:00:00Z"}]}}`),
					PeriodBegin: time.Date(2022, 10, 26, 12, 0, 0, 0, time.UTC),
					PeriodEnd:   time.Date(2022, 11, 1, 0, 0, 0, 0, time.UTC),
				},
				body: `{
	"Vaktor.Vaktor_TiddataResponse": {
		"Vaktor.Vaktor_TiddataResult": [{
			"Vaktor.nav_id": "123456",
			"Vaktor.resource_id": "E123456",
			"Vaktor.leder_resource_id": "654321",
			"Vaktor.leder_nav_id": "M654321",
			"Vaktor.leder_navn": "Kalpana, Bran",
			"Vaktor.leder_epost": "Bran.Kalpana@nav.no",
			"Vaktor.dager": "[{\"dato\":\"2022-10-26T00:00:00\",\"skjema_tid\":7.75,\"skjema_navn\":\"BV 0800-1545 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":2,\"ansatt_dato_godkjent_av\":\"b140650\",\"godkjent_dato\":\"2022-11-03T17:11:48\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-10-26T07:01:58\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-26T14:59:32\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null}],\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"850000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"AK_UNIO\",\"TILL_LONN\":\"0\",\"RATE_K001\":850000,\"RATE_I143\":0,\"RATE_B100\":850000,\"RATE_K170\":35,\"RATE_K171\":15,\"RATE_K172\":25,\"RATE_K160\":25,\"RATE_K161\":65,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-31T00:00:00\",\"skjema_tid\":7.75,\"skjema_navn\":\"BV 0800-1545 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":2,\"ansatt_dato_godkjent_av\":\"b140650\",\"godkjent_dato\":\"2022-11-03T17:11:48\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-10-31T06:55:03\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-31T14:56:21\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null}],\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"850000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"AK_UNIO\",\"TILL_LONN\":\"0\",\"RATE_K001\":850000,\"RATE_I143\":0,\"RATE_B100\":850000,\"RATE_K170\":35,\"RATE_K171\":15,\"RATE_K172\":25,\"RATE_K160\":25,\"RATE_K161\":65,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-30T00:00:00\",\"skjema_tid\":0,\"skjema_navn\":\"BV Søndag IKT\",\"godkjent\":2,\"ansatt_dato_godkjent_av\":\"b140650\",\"godkjent_dato\":\"2022-11-03T17:11:48\",\"virkedag\":\"Søndag\",\"Stemplinger\":null,\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"850000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"AK_UNIO\",\"TILL_LONN\":\"0\",\"RATE_K001\":850000,\"RATE_I143\":0,\"RATE_B100\":850000,\"RATE_K170\":35,\"RATE_K171\":15,\"RATE_K172\":25,\"RATE_K160\":25,\"RATE_K161\":65,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-29T00:00:00\",\"skjema_tid\":0,\"skjema_navn\":\"BV Lørdag IKT\",\"godkjent\":2,\"ansatt_dato_godkjent_av\":\"b140650\",\"godkjent_dato\":\"2022-11-03T17:11:48\",\"virkedag\":\"Lørdag\",\"Stemplinger\":null,\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"850000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"AK_UNIO\",\"TILL_LONN\":\"0\",\"RATE_K001\":850000,\"RATE_I143\":0,\"RATE_B100\":850000,\"RATE_K170\":35,\"RATE_K171\":15,\"RATE_K172\":25,\"RATE_K160\":25,\"RATE_K161\":65,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-28T00:00:00\",\"skjema_tid\":7.75,\"skjema_navn\":\"BV 0800-1545 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":2,\"ansatt_dato_godkjent_av\":\"b140650\",\"godkjent_dato\":\"2022-11-03T17:11:48\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-10-28T07:02:29\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-28T15:52:28\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null}],\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"850000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"AK_UNIO\",\"TILL_LONN\":\"0\",\"RATE_K001\":850000,\"RATE_I143\":0,\"RATE_B100\":850000,\"RATE_K170\":35,\"RATE_K171\":15,\"RATE_K172\":25,\"RATE_K160\":25,\"RATE_K161\":65,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-27T00:00:00\",\"skjema_tid\":7.75,\"skjema_navn\":\"BV 0800-1545 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":2,\"ansatt_dato_godkjent_av\":\"b140650\",\"godkjent_dato\":\"2022-11-03T17:11:48\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-10-27T07:16:24\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-27T16:04:18\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null}],\"Stillinger\":[{\"post_id\":\"258\",\"parttime_pct\":100,\"post_code\":\"1364\",\"post_description\":\"Seniorrådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"850000\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"AK_UNIO\",\"TILL_LONN\":\"0\",\"RATE_K001\":850000,\"RATE_I143\":0,\"RATE_B100\":850000,\"RATE_K170\":35,\"RATE_K171\":15,\"RATE_K172\":25,\"RATE_K160\":25,\"RATE_K161\":65,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]}]"
		}]
	}
}`,
			},
			want: want{
				payroll: models.Payroll{
					ID:           uuid.MustParse("b4ac8e53-9d64-4557-8ef8-d00774ab9c06"),
					ApproverID:   "M654321",
					ApproverName: "Kalpana, Bran",
					Artskoder: models.Artskoder{
						Morgen: models.Artskode{
							Sum:   decimal.NewFromFloat(3758.11),
							Hours: 18,
						},
						Kveld: models.Artskode{
							Sum:   decimal.NewFromFloat(3340.54),
							Hours: 16,
						},
						Dag: models.Artskode{
							Sum:   decimal.NewFromFloat(3209.59),
							Hours: 21,
						},
						Helg: models.Artskode{
							Sum:   decimal.NewFromFloat(10587.41),
							Hours: 49, // stilte klokken en time tilbake denne vakten
						},
						Skift: models.Artskode{
							Sum:   decimal.NewFromFloat(75),
							Hours: 15,
						},
					},
				},
				ok: false,
			},
		},

		{
			name: "helg med overtid (ikke merket bv)",
			args: args{
				beredskapsvakt: gensql.Beredskapsvakt{
					Ident:       "a123456",
					Plan:        json.RawMessage(`{"id":"b4ac8e53-9d64-4557-8ef8-d00774ab9c06","user_id":"E123456","start_timestamp":"2022-10-15T00:00:00Z","end_timestamp":"2022-10-16T00:00:00Z","schedule":{"2022-10-15":[{"start_timestamp":"2022-10-15T00:00:00Z","end_timestamp":"2022-10-16T00:00:00Z"}],"2022-10-16":[{"start_timestamp":"2022-10-16T00:00:00Z","end_timestamp":"2022-10-17T00:00:00Z"}]}}`),
					PeriodBegin: time.Date(2022, 10, 15, 0, 0, 0, 0, time.UTC),
					PeriodEnd:   time.Date(2022, 10, 16, 0, 0, 0, 0, time.UTC),
				},
				body: `{
	"Vaktor.Vaktor_TiddataResponse": {
		"Vaktor.Vaktor_TiddataResult": [{
			"Vaktor.nav_id": "123456",
			"Vaktor.resource_id": "E123456",
			"Vaktor.leder_resource_id": "654321",
			"Vaktor.leder_nav_id": "M654321",
			"Vaktor.leder_navn": "Kalpana, Bran",
			"Vaktor.leder_epost": "Bran.Kalpana@nav.no",
			"Vaktor.dager": "[{\"dato\":\"2022-10-16T00:00:00\",\"skjema_tid\":0,\"skjema_navn\":\"BV Søndag IKT\",\"godkjent\":3,\"ansatt_dato_godkjent_av\":\"M654321\",\"godkjent_dato\":\"2022-11-02T08:46:53\",\"virkedag\":\"Søndag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-10-16T15:59:56\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-16T17:48:38\",\"Navn\":\"Overtid                 \",\"Type\":\"B6\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":\"Statussjekk helg + populering av database for bussines objects.\"},{\"Stempling_Tid\":\"2022-10-16T17:48:40\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-16T20:51:58\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-16T21:02:06\",\"Navn\":\"Overtid                 \",\"Type\":\"B6\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":\"Duplisering av Business Objects-base https://jira.adeo.no/browse/IKT-475117\"},{\"Stempling_Tid\":\"2022-10-16T21:02:09\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null}],\"Stillinger\":[{\"post_id\":\"265\",\"parttime_pct\":100,\"post_code\":\"1434\",\"post_description\":\"Rådgiver\",\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"68\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":636700,\"RATE_I143\":0,\"RATE_B100\":636700,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-15T00:00:00\",\"skjema_tid\":0,\"skjema_navn\":\"BV Lørdag IKT\",\"godkjent\":3,\"ansatt_dato_godkjent_av\":\"M654321\",\"godkjent_dato\":\"2022-11-02T08:46:53\",\"virkedag\":\"Lørdag\",\"Stemplinger\":null,\"Stillinger\":[{\"post_id\":\"265\",\"parttime_pct\":100,\"post_code\":\"1434\",\"post_description\":\"Rådgiver\",\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"68\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":636700,\"RATE_I143\":0,\"RATE_B100\":636700,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]}]"
		}]
	}
}`,
			},
			want: want{
				payroll: models.Payroll{
					ID:           uuid.MustParse("b4ac8e53-9d64-4557-8ef8-d00774ab9c06"),
					ApproverID:   "M654321",
					ApproverName: "Kalpana, Bran",
					Artskoder: models.Artskoder{
						Morgen: models.Artskode{
							Sum:   decimal.NewFromFloat(0),
							Hours: 0,
						},
						Kveld: models.Artskode{
							Sum:   decimal.NewFromFloat(0),
							Hours: 0,
						},
						Dag: models.Artskode{
							Sum:   decimal.NewFromFloat(0),
							Hours: 0,
						},
						Helg: models.Artskode{
							Sum:   decimal.NewFromFloat(7820.58),
							Hours: 46,
						},
						Utrykning: models.Artskode{
							Sum:   decimal.NewFromFloat(0),
							Hours: 0,
						},
					},
				},
				ok: false,
			},
		},

		{
			name: "ukesvakt med helg i overtid (ikke merket bv)",
			args: args{
				beredskapsvakt: gensql.Beredskapsvakt{
					Ident:       "a123456",
					Plan:        json.RawMessage(`{"id":"b4ac8e53-9d64-4557-8ef8-d00774ab9c06","user_id":"E123456","start_timestamp":"2022-10-15T00:00:00Z","end_timestamp":"2022-10-16T00:00:00Z","schedule":{"2022-10-10":[{"start_timestamp":"2022-10-10T12:00:00Z","end_timestamp":"2022-10-11T00:00:00Z"}],"2022-10-11":[{"start_timestamp":"2022-10-11T00:00:00Z","end_timestamp":"2022-10-12T00:00:00Z"}],"2022-10-12":[{"start_timestamp":"2022-10-12T00:00:00Z","end_timestamp":"2022-10-13T00:00:00Z"}],"2022-10-13":[{"start_timestamp":"2022-10-13T00:00:00Z","end_timestamp":"2022-10-14T00:00:00Z"}],"2022-10-14":[{"start_timestamp":"2022-10-14T00:00:00Z","end_timestamp":"2022-10-15T00:00:00Z"}],"2022-10-15":[{"start_timestamp":"2022-10-15T00:00:00Z","end_timestamp":"2022-10-16T00:00:00Z"}],"2022-10-16":[{"start_timestamp":"2022-10-16T00:00:00Z","end_timestamp":"2022-10-17T00:00:00Z"}],"2022-10-17":[{"start_timestamp":"2022-10-17T00:00:00Z","end_timestamp":"2022-10-18T12:00:00Z"}]}}`),
					PeriodBegin: time.Date(2022, 10, 10, 0, 0, 0, 0, time.UTC),
					PeriodEnd:   time.Date(2022, 10, 17, 0, 0, 0, 0, time.UTC),
				},
				body: `{
	"Vaktor.Vaktor_TiddataResponse": {
		"Vaktor.Vaktor_TiddataResult": [{
			"Vaktor.nav_id": "123456",
			"Vaktor.resource_id": "E123456",
			"Vaktor.leder_resource_id": "654321",
			"Vaktor.leder_nav_id": "M654321",
			"Vaktor.leder_navn": "Kalpana, Bran",
			"Vaktor.leder_epost": "Bran.Kalpana@nav.no",
			"Vaktor.dager": "[{\"dato\":\"2022-10-10T00:00:00\",\"skjema_tid\":7.5,\"skjema_navn\":\"BV Heltid 0800-1530 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":3,\"ansatt_dato_godkjent_av\":\"M654321\",\"godkjent_dato\":\"2022-11-02T08:46:53\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-10-10T06:52:42\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-10T15:53:09\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-10T12:20:56\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":1,\"Fravar_kode_navn\":\"Inne\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-10T12:06:46\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-10T09:16:47\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":1,\"Fravar_kode_navn\":\"Inne\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-10T09:01:34\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null}],\"Stillinger\":[{\"post_id\":\"265\",\"parttime_pct\":100,\"post_code\":\"1434\",\"post_description\":\"Rådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"68\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":636700,\"RATE_I143\":0,\"RATE_B100\":636700,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-17T00:00:00\",\"skjema_tid\":7.5,\"skjema_navn\":\"BV Heltid 0800-1530 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":3,\"ansatt_dato_godkjent_av\":\"M654321\",\"godkjent_dato\":\"2022-11-02T08:46:53\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-10-17T10:08:09\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":1,\"Fravar_kode_navn\":\"Inne\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-17T16:05:04\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-17T13:05:03\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":1,\"Fravar_kode_navn\":\"Inne\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-17T12:51:38\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null}],\"Stillinger\":[{\"post_id\":\"265\",\"parttime_pct\":100,\"post_code\":\"1434\",\"post_description\":\"Rådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"68\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":636700,\"RATE_I143\":0,\"RATE_B100\":636700,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-16T00:00:00\",\"skjema_tid\":0,\"skjema_navn\":\"BV Søndag IKT\",\"godkjent\":3,\"ansatt_dato_godkjent_av\":\"M654321\",\"godkjent_dato\":\"2022-11-02T08:46:53\",\"virkedag\":\"Søndag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-10-16T15:59:56\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-16T21:02:09\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-16T21:02:06\",\"Navn\":\"Overtid                 \",\"Type\":\"B6\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":\"Duplisering av Business Objects-base https:\/\/jira.adeo.no\/browse\/IKT-475117\"},{\"Stempling_Tid\":\"2022-10-16T20:51:58\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-16T17:48:40\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-16T17:48:38\",\"Navn\":\"Overtid                 \",\"Type\":\"B6\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":\"Statussjekk helg + populering av database for bussines objects.\"}],\"Stillinger\":[{\"post_id\":\"265\",\"parttime_pct\":100,\"post_code\":\"1434\",\"post_description\":\"Rådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"68\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":636700,\"RATE_I143\":0,\"RATE_B100\":636700,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-15T00:00:00\",\"skjema_tid\":0,\"skjema_navn\":\"BV Lørdag IKT\",\"godkjent\":3,\"ansatt_dato_godkjent_av\":\"M654321\",\"godkjent_dato\":\"2022-11-02T08:46:53\",\"virkedag\":\"Lørdag\",\"Stemplinger\":null,\"Stillinger\":[{\"post_id\":\"265\",\"parttime_pct\":100,\"post_code\":\"1434\",\"post_description\":\"Rådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"68\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":636700,\"RATE_I143\":0,\"RATE_B100\":636700,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-14T00:00:00\",\"skjema_tid\":7.5,\"skjema_navn\":\"BV Heltid 0800-1530 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":3,\"ansatt_dato_godkjent_av\":\"M654321\",\"godkjent_dato\":\"2022-11-02T08:46:53\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-10-14T08:09:47\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-14T16:19:46\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null}],\"Stillinger\":[{\"post_id\":\"265\",\"parttime_pct\":100,\"post_code\":\"1434\",\"post_description\":\"Rådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"68\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":636700,\"RATE_I143\":0,\"RATE_B100\":636700,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-13T00:00:00\",\"skjema_tid\":7.5,\"skjema_navn\":\"BV Heltid 0800-1530 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":3,\"ansatt_dato_godkjent_av\":\"M654321\",\"godkjent_dato\":\"2022-11-02T08:46:53\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-10-13T09:35:02\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":1,\"Fravar_kode_navn\":\"Inne\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-13T16:35:03\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null}],\"Stillinger\":[{\"post_id\":\"265\",\"parttime_pct\":100,\"post_code\":\"1434\",\"post_description\":\"Rådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"68\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":636700,\"RATE_I143\":0,\"RATE_B100\":636700,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-12T00:00:00\",\"skjema_tid\":7.5,\"skjema_navn\":\"BV Heltid 0800-1530 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":3,\"ansatt_dato_godkjent_av\":\"M654321\",\"godkjent_dato\":\"2022-11-02T08:46:53\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-10-12T07:50:17\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-12T14:38:11\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null}],\"Stillinger\":[{\"post_id\":\"265\",\"parttime_pct\":100,\"post_code\":\"1434\",\"post_description\":\"Rådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"68\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":636700,\"RATE_I143\":0,\"RATE_B100\":636700,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]},{\"dato\":\"2022-10-11T00:00:00\",\"skjema_tid\":7.5,\"skjema_navn\":\"BV Heltid 0800-1530 m\/Beredskapsvakt, start vakt kl 1600 (2018)\",\"godkjent\":3,\"ansatt_dato_godkjent_av\":\"M654321\",\"godkjent_dato\":\"2022-11-02T08:46:53\",\"virkedag\":\"Virkedag\",\"Stemplinger\":[{\"Stempling_Tid\":\"2022-10-11T07:54:09\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-11T16:11:56\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-11T11:12:15\",\"Navn\":\"Inn\",\"Type\":\"B1\",\"Fravar_kode\":1,\"Fravar_kode_navn\":\"Inne\",\"Overtid_Begrunnelse\":null},{\"Stempling_Tid\":\"2022-10-11T10:50:29\",\"Navn\":\"Ut\",\"Type\":\"B2\",\"Fravar_kode\":0,\"Fravar_kode_navn\":\"Ute\",\"Overtid_Begrunnelse\":null}],\"Stillinger\":[{\"post_id\":\"265\",\"parttime_pct\":100,\"post_code\":\"1434\",\"post_description\":\"Rådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"68\",\"aga\":\"060501180000\",\"statskonto\":\"060501110000\",\"HTA\":\"\",\"TILL_LONN\":\"0\",\"RATE_K001\":636700,\"RATE_I143\":0,\"RATE_B100\":636700,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_KSTED\":[{\"role_id\":\"\",\"domain_info\":\"\"},{\"role_id\":\"\",\"domain_info\":\"\"}],\"BDM_FORMAL\":null}]}]"
		}]
	}
}`,
			},
			want: want{
				payroll: models.Payroll{
					ID:           uuid.MustParse("b4ac8e53-9d64-4557-8ef8-d00774ab9c06"),
					ApproverID:   "M654321",
					ApproverName: "Kalpana, Bran",
					Artskoder: models.Artskoder{
						Morgen: models.Artskode{
							Sum:   decimal.NewFromFloat(4879.95),
							Hours: 30,
						},
						Kveld: models.Artskode{
							Sum:   decimal.NewFromFloat(3903.96),
							Hours: 24,
						},
						Dag: models.Artskode{
							Sum:   decimal.NewFromFloat(4138.7),
							Hours: 35,
						},
						Helg: models.Artskode{
							Sum:   decimal.NewFromFloat(7820.58),
							Hours: 46,
						},
						Skift: models.Artskode{
							Sum:   decimal.NewFromFloat(115),
							Hours: 23,
						},
						Utrykning: models.Artskode{
							Sum:   decimal.NewFromFloat(0),
							Hours: 0,
						},
					},
				},
				ok: false,
			},
		},

		{
			name: "vakt på nyttårsaften",
			args: args{
				beredskapsvakt: gensql.Beredskapsvakt{
					Ident:       "a123456",
					Plan:        json.RawMessage(`{"id":"b4ac8e53-9d64-4557-8ef8-d00774ab9c06","user_id":"E123456","start_timestamp":"2022-12-31T00:00:00Z","end_timestamp":"2023-01-01T00:00:00Z","schedule":{"2022-12-31":[{"start_timestamp":"2022-12-31T00:00:00Z","end_timestamp":"2023-01-01T00:00:00Z"}]}}`),
					PeriodBegin: time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC),
					PeriodEnd:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				body: `{
	"Vaktor.Vaktor_TiddataResponse": {
		"Vaktor.Vaktor_TiddataResult": [{
			"Vaktor.nav_id": "123456",
			"Vaktor.resource_id": "E123456",
			"Vaktor.leder_resource_id": "654321",
			"Vaktor.leder_nav_id": "M654321",
			"Vaktor.leder_navn": "Kalpana, Bran",
			"Vaktor.leder_epost": "Bran.Kalpana@nav.no",
			"Vaktor.dager": "[{\"dato\":\"2022-12-31T00:00:00\",\"skjema_tid\":0,\"skjema_navn\":\"BV Lørdag IKT\",\"godkjent\":2,\"ansatt_dato_godkjent_av\":\"f102546\",\"godkjent_dato\":\"2023-01-03T14:06:39\",\"virkedag\":\"Lørdag\",\"Stemplinger\":null,\"Stillinger\":[{\"post_id\":\"265\",\"parttime_pct\":100,\"post_code\":\"1434\",\"post_description\":\"Rådgiver\",\"parttime_pct\":100,\"koststed\":\"000000\",\"formal\":\"000000\",\"aktivitet\":\"000000\",\"scale_id\":\"68\",\"aga\":\"000000\",\"statskonto\":\"000000\",\"HTA\":\"LO_YS\",\"TILL_LONN\":\"0\",\"RATE_K001\":500000,\"RATE_I143\":0,\"RATE_B100\":500000,\"RATE_K170\":35,\"RATE_K171\":10,\"RATE_K172\":20,\"RATE_K160\":15,\"RATE_K161\":55,\"RATE_G014\":33.33,\"BDM_FORMAL\":null}]}]"
		}]
	}
}`,
			},
			want: want{
				payroll: models.Payroll{
					ID:           uuid.MustParse("b4ac8e53-9d64-4557-8ef8-d00774ab9c06"),
					ApproverID:   "M654321",
					ApproverName: "Kalpana, Bran",
					Artskoder: models.Artskoder{
						Helg: models.Artskode{
							Sum:   decimal.NewFromFloat(3366.59),
							Hours: 24,
						},
					},
				},
				ok: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var response Response
			err := json.Unmarshal([]byte(tt.args.body), &response)
			if err != nil {
				t.Errorf("failed while unmarshling: %v", err)
				return
			}

			got, ok := calculateSalary(log, tt.args.beredskapsvakt, response)
			if (!ok) != tt.want.ok {
				t.Errorf("calculateSalary() ok = %v, want.ok %v", ok, tt.want.ok)
				return
			}
			if diff := cmp.Diff(tt.want.payroll, got); diff != "" {
				t.Errorf("calculateSalary() mismatch (-want +got):\n%s", diff)
			}
		})
	}
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
				date: time.Date(2022, 11, 6, 0, 0, 0, 0, time.UTC),
			},
			want: models.Clocking{
				In:  time.Date(2022, 11, 6, 8, 0, 0, 0, time.UTC),
				Out: time.Date(2022, 11, 6, 15, 45, 0, 0, time.UTC),
			},
		},
		{
			name: "Sommertid",
			args: args{
				tid:  7,
				date: time.Date(2022, 11, 6, 0, 0, 0, 0, time.UTC),
			},
			want: models.Clocking{
				In:  time.Date(2022, 11, 6, 8, 0, 0, 0, time.UTC),
				Out: time.Date(2022, 11, 6, 15, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Normaltid",
			args: args{
				tid:  7.5,
				date: time.Date(2022, 11, 6, 0, 0, 0, 0, time.UTC),
			},
			want: models.Clocking{
				In:  time.Date(2022, 11, 6, 8, 0, 0, 0, time.UTC),
				Out: time.Date(2022, 11, 6, 15, 30, 0, 0, time.UTC),
			},
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
		days []Dag
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Ingen godkjente dager",
			args: args{
				days: []Dag{
					{
						Godkjent: 0,
					},
				},
			},
			want: false,
		},
		{
			name: "Godkjent av vakthaver",
			args: args{
				days: []Dag{
					{
						Godkjent: 1,
					},
				},
			},
			want: false,
		},
		{
			name: "Godkjent av personalleder",
			args: args{
				days: []Dag{
					{
						Godkjent: 2,
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTimesheetApproved(tt.args.days); got != tt.want {
				t.Errorf("isTimesheetApproved() = %v, want %v", got, tt.want)
			}
		})
	}
}
