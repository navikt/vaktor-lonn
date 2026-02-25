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
	"github.com/shopspring/decimal"
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
					{
						Dato:       "2022-08-02T00:00:00",
						SkjemaTid:  7,
						SkjemaNavn: "Heltid 0800-1500 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []models.MWTStempling{
							{
								StemplingTid: "2022-08-02T07:45:10",
								Retning:      "Inn",
								Type:         "B1",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2022-08-02T14:30:11",
								Retning:      "Ut",
								Type:         "B2",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2022-08-02T14:31:00",
								Retning:      "Inn",
								Type:         "B1",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2022-08-02T14:31:01",
								Retning:      "Ut på fravær",
								Type:         "B5",
								Fravarkode:   940,
							},
							{
								StemplingTid: "2022-08-02T16:00:00",
								Retning:      "Inn fra fravær",
								Type:         "B4",
								Fravarkode:   940,
							},
							{
								StemplingTid: "2022-08-02T16:00:01",
								Retning:      "Ut",
								Type:         "B2",
								Fravarkode:   0,
							},
						},
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 500_000,
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
							Out: time.Date(2022, 8, 2, 14, 31, 1, 0, time.UTC),
						},
						{
							In:  time.Date(2022, 8, 2, 16, 0, 0, 0, time.UTC),
							Out: time.Date(2022, 8, 2, 16, 0, 1, 0, time.UTC),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "arbeidsdag med to fravær",
			args: args{
				days: []models.MWTDag{
					{
						Dato:       "2023-06-07T00:00:00",
						SkjemaTid:  7,
						SkjemaNavn: "Heltid 0800-1500 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []models.MWTStempling{
							{
								StemplingTid: "2023-06-07T08:29:46",
								Retning:      "Inn",
								Type:         "B1",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2023-06-07T11:43:10",
								Retning:      "Ut på fravær",
								Type:         "B5",
								Fravarkode:   180,
							},
							{
								StemplingTid: "2023-06-07T12:39:09",
								Retning:      "Inn fra fravær",
								Type:         "B4",
								Fravarkode:   180,
							},
							{
								StemplingTid: "2023-06-07T12:40:00",
								Retning:      "Ut på fravær",
								Type:         "B5",
								Fravarkode:   920,
							},
							{
								StemplingTid: "2023-06-07T13:30:00",
								Retning:      "Inn fra fravær",
								Type:         "B4",
								Fravarkode:   920,
							},
							{
								StemplingTid: "2023-06-07T15:11:43",
								Retning:      "Ut",
								Type:         "B2",
								Fravarkode:   0,
							},
						},
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 500_000,
							},
						},
					},
				},
			},
			want: map[string]models.TimeSheet{
				"2023-06-07": {
					Date:         time.Date(2023, 6, 7, 0, 0, 0, 0, time.UTC),
					WorkingHours: 7,
					WorkingDay:   "Virkedag",
					FormName:     "Heltid 0800-1500 (2018)",
					Salary:       decimal.NewFromInt(500_000),
					Clockings: []models.Clocking{
						{
							In:  time.Date(2023, 6, 7, 8, 29, 46, 0, time.UTC),
							Out: time.Date(2023, 6, 7, 11, 43, 10, 0, time.UTC),
						},
						{
							In:  time.Date(2023, 6, 7, 12, 39, 9, 0, time.UTC),
							Out: time.Date(2023, 6, 7, 12, 40, 0, 0, time.UTC),
						},
						{
							In:  time.Date(2023, 6, 7, 13, 30, 0, 0, time.UTC),
							Out: time.Date(2023, 6, 7, 15, 11, 43, 0, time.UTC),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "kveld med utrykning (glemt BV begrunnelse)",
			args: args{
				days: []models.MWTDag{
					{
						Dato:       "2023-02-14T00:00:00",
						SkjemaTid:  7.45,
						SkjemaNavn: "Heltid 0800-1545 (2018)",
						Godkjent:   5,
						Virkedag:   "Virkedag",
						Stemplinger: []models.MWTStempling{
							{
								StemplingTid: "2023-02-14T08:00:00",
								Retning:      "Inn",
								Type:         "B1",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2023-02-14T15:45:00",
								Retning:      "Ut",
								Type:         "B2",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2023-02-14T20:30:00",
								Retning:      "Inn",
								Type:         "B1",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2023-02-14T22:29:59",
								Retning:      "Overtid",
								Type:         "B6",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2023-02-14T22:30:00",
								Retning:      "Ut",
								Type:         "B2",
								Fravarkode:   0,
							},
						},
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 500_000,
							},
						},
					},
				},
			},
			want: map[string]models.TimeSheet{
				"2023-02-14": {
					Date:         time.Date(2023, 2, 14, 0, 0, 0, 0, time.UTC),
					WorkingHours: 7.45,
					WorkingDay:   "Virkedag",
					FormName:     "Heltid 0800-1545 (2018)",
					Salary:       decimal.NewFromInt(500_000),
					Clockings: []models.Clocking{
						{
							In:  time.Date(2023, 2, 14, 8, 0, 0, 0, time.UTC),
							Out: time.Date(2023, 2, 14, 15, 45, 0, 0, time.UTC),
						},
						{
							In:  time.Date(2023, 2, 14, 20, 30, 0, 0, time.UTC),
							Out: time.Date(2023, 2, 14, 22, 30, 0, 0, time.UTC),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "helg med utrykning (liten og stor BV begrunnelse)",
			args: args{
				days: []models.MWTDag{
					{
						Dato:       "2023-02-04T00:00:00",
						SkjemaTid:  0,
						SkjemaNavn: "BV Lørdag IKT",
						Godkjent:   5,
						Virkedag:   "Lørdag",
						Stemplinger: []models.MWTStempling{
							{
								StemplingTid: "2023-02-04T20:30:00",
								Retning:      "Inn",
								Type:         "B1",
								Fravarkode:   0,
							},
							{
								StemplingTid:       "2023-02-04T22:29:59",
								Retning:            "Overtid",
								Type:               "B6",
								Fravarkode:         0,
								OvertidBegrunnelse: "BV",
							},
							{
								StemplingTid: "2023-02-04T22:30:00",
								Retning:      "Ut",
								Type:         "B2",
								Fravarkode:   0,
							},
						},
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 500_000,
							},
						},
					},
					{
						Dato:       "2023-02-11T00:00:00",
						SkjemaTid:  0,
						SkjemaNavn: "BV Lørdag IKT",
						Godkjent:   5,
						Virkedag:   "Lørdag",
						Stemplinger: []models.MWTStempling{
							{
								StemplingTid: "2023-02-11T20:30:00",
								Retning:      "Inn",
								Type:         "B1",
								Fravarkode:   0,
							},
							{
								StemplingTid:       "2023-02-11T22:29:59",
								Retning:            "Overtid",
								Type:               "B6",
								Fravarkode:         0,
								OvertidBegrunnelse: "bv",
							},
							{
								StemplingTid: "2023-02-11T22:30:00",
								Retning:      "Ut",
								Type:         "B2",
								Fravarkode:   0,
							},
						},
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 500_000,
							},
						},
					},
				},
			},
			want: map[string]models.TimeSheet{
				"2023-02-04": {
					Date:         time.Date(2023, 2, 4, 0, 0, 0, 0, time.UTC),
					WorkingHours: 0,
					WorkingDay:   "Lørdag",
					FormName:     "BV Lørdag IKT",
					Salary:       decimal.NewFromInt(500_000),
					Clockings: []models.Clocking{
						{
							In:  time.Date(2023, 2, 4, 20, 30, 0, 0, time.UTC),
							Out: time.Date(2023, 2, 4, 22, 30, 0, 0, time.UTC),
							OtG: true,
						},
					},
				},
				"2023-02-11": {
					Date:         time.Date(2023, 2, 11, 0, 0, 0, 0, time.UTC),
					WorkingHours: 0,
					WorkingDay:   "Lørdag",
					FormName:     "BV Lørdag IKT",
					Salary:       decimal.NewFromInt(500_000),
					Clockings: []models.Clocking{
						{
							In:  time.Date(2023, 2, 11, 20, 30, 0, 0, time.UTC),
							Out: time.Date(2023, 2, 11, 22, 30, 0, 0, time.UTC),
							OtG: true,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "helg med utrykning (før krav om BV begrunnelse)",
			args: args{
				days: []models.MWTDag{
					{
						Dato:       "2022-09-17T00:00:00",
						SkjemaTid:  0,
						SkjemaNavn: "BV Lørdag IKT",
						Godkjent:   5,
						Virkedag:   "Lørdag",
						Stemplinger: []models.MWTStempling{
							{
								StemplingTid: "2022-09-17T20:30:00",
								Retning:      "Inn",
								Type:         "B1",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2022-09-17T22:29:59",
								Retning:      "Overtid",
								Type:         "B6",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2022-09-17T22:30:00",
								Retning:      "Ut",
								Type:         "B2",
								Fravarkode:   0,
							},
						},
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 500_000,
							},
						},
					},
					{
						Dato:       "2022-09-24T00:00:00",
						SkjemaTid:  0,
						SkjemaNavn: "BV Lørdag IKT",
						Godkjent:   5,
						Virkedag:   "Lørdag",
						Stemplinger: []models.MWTStempling{
							{
								StemplingTid: "2022-09-24T20:30:00",
								Retning:      "Inn",
								Type:         "B1",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2022-09-24T22:29:59",
								Retning:      "Overtid",
								Type:         "B6",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2022-09-24T22:30:00",
								Retning:      "Ut",
								Type:         "B2",
								Fravarkode:   0,
							},
						},
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 500_000,
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
				days: []models.MWTDag{
					{
						Dato:       "2022-05-03T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "Heltid 0800-1545 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []models.MWTStempling{
							{
								StemplingTid: "2022-05-03T08:00:00",
								Retning:      "Inn",
								Type:         "B1",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2022-05-03T08:00:01",
								Retning:      "Ut på fravær",
								Type:         "B5",
								Fravarkode:   740,
							},
						},
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 500_000,
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
				days: []models.MWTDag{
					{
						Dato:       "2022-10-17T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "Heltid 0800-1545 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []models.MWTStempling{
							{
								StemplingTid: "2022-10-17T08:00:00",
								Retning:      "Inn",
								Type:         "B1",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2022-10-17T08:00:01",
								Retning:      "Ut på fravær",
								Type:         "B5",
								Fravarkode:   630,
							},
							{
								StemplingTid: "2022-10-17T15:45:00",
								Retning:      "Inn fra fravær",
								Type:         "B4",
								Fravarkode:   630,
							},
							{
								StemplingTid: "2022-10-17T15:45:01",
								Retning:      "Ut",
								Type:         "B2",
								Fravarkode:   0,
							},
						},
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 500_000,
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
				days: []models.MWTDag{
					{
						Dato:       "2022-10-20T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "Heltid 0800-1545 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []models.MWTStempling{
							{
								StemplingTid: "2022-10-20T11:12:10",
								Retning:      "Inn fra fravær",
								Type:         "B4",
								Fravarkode:   630,
							},
							{
								StemplingTid: "2022-10-20T16:00:00",
								Retning:      "Ut",
								Type:         "B2",
								Fravarkode:   0,
							},
						},
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 500_000,
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
				days: []models.MWTDag{
					{
						Dato:       "2022-09-15T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []models.MWTStempling{
							{
								StemplingTid:       "2022-09-15T00:34:21",
								Retning:            "Inn",
								Type:               "B1",
								Fravarkode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-09-15T00:34:24",
								Retning:            "Overtid",
								Type:               "B6",
								Fravarkode:         0,
								OvertidBegrunnelse: "BV - IKT-478705 DVH",
							},
							{
								StemplingTid:       "2022-09-15T01:34:42",
								Retning:            "Ut",
								Type:               "B2",
								Fravarkode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid: "2022-09-15T03:10:00",
								Retning:      "Inn",
								Type:         "B1",
								Fravarkode:   0,
							},
							{
								StemplingTid:       "2022-09-15T03:31:00",
								Retning:            "Overtid",
								Type:               "B6",
								Fravarkode:         0,
								OvertidBegrunnelse: "BV",
							},
							{
								StemplingTid: "2022-09-15T04:32:00",
								Retning:      "Ut",
								Type:         "B2",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2022-09-15T08:04:00",
								Retning:      "Inn",
								Type:         "B1",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2022-09-15T16:26:00",
								Retning:      "Ut",
								Type:         "B2",
								Fravarkode:   0,
							},
						},
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 500_000,
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
				days: []models.MWTDag{
					{
						Dato:       "2022-09-15T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []models.MWTStempling{
							{
								StemplingTid: "2022-09-15T08:04:00",
								Retning:      "Inn",
								Type:         "B1",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2022-09-15T16:26:00",
								Retning:      "Ut",
								Type:         "B2",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2022-09-15T23:10:00",
								Retning:      "Inn",
								Type:         "B1",
								Fravarkode:   0,
							},
							{
								StemplingTid:       "2022-09-15T23:31:00",
								Retning:            "Overtid",
								Type:               "B6",
								Fravarkode:         0,
								OvertidBegrunnelse: "BV",
							},
							{
								StemplingTid: "2022-09-16T00:32:00",
								Retning:      "Ut",
								Type:         "B2",
								Fravarkode:   0,
							},
						},
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 500_000,
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
				days: []models.MWTDag{
					{
						Dato:       "2022-09-15T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []models.MWTStempling{
							{
								StemplingTid: "2022-09-15T08:04:00",
								Retning:      "Inn",
								Type:         "B1",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2022-09-15T16:26:00",
								Retning:      "Ut",
								Type:         "B2",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2022-09-15T23:10:00",
								Retning:      "Inn",
								Type:         "B1",
								Fravarkode:   0,
							},
							{
								StemplingTid:       "2022-09-15T23:31:00",
								Retning:            "Overtid",
								Type:               "B6",
								Fravarkode:         0,
								OvertidBegrunnelse: "BV",
							},
							{
								StemplingTid: "2022-09-16T00:32:00",
								Retning:      "Ut",
								Type:         "B2",
								Fravarkode:   0,
							},
						},
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 500_000,
							},
						},
					},
					{
						Dato:       "2022-09-16T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []models.MWTStempling{
							{
								StemplingTid: "2022-09-16T08:04:00",
								Retning:      "Inn",
								Type:         "B1",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2022-09-16T15:41:00",
								Retning:      "Ut",
								Type:         "B2",
								Fravarkode:   0,
							},
						},
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 500_000,
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
				days: []models.MWTDag{
					{
						Dato:       "2022-10-25T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []models.MWTStempling{
							{
								StemplingTid:       "2022-10-25T00:34:21",
								Retning:            "Inn",
								Type:               "B1",
								Fravarkode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-25T00:34:24",
								Retning:            "Overtid",
								Type:               "B6",
								Fravarkode:         0,
								OvertidBegrunnelse: "BV - IKT-478705 DVH",
							},
							{
								StemplingTid:       "2022-10-25T01:34:42",
								Retning:            "Ut",
								Type:               "B2",
								Fravarkode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-25T06:34:45",
								Retning:            "Inn",
								Type:               "B1",
								Fravarkode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-25T07:18:43",
								Retning:            "Ut",
								Type:               "B2",
								Fravarkode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-25T08:47:49",
								Retning:            "Inn",
								Type:               "B1",
								Fravarkode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-25T15:48:30",
								Retning:            "Ut",
								Type:               "B2",
								Fravarkode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-25T23:31:37",
								Retning:            "Inn",
								Type:               "B1",
								Fravarkode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-26T00:45:34",
								Retning:            "Overtid",
								Type:               "B6",
								Fravarkode:         1,
								OvertidBegrunnelse: "BV - Feilsøking ifbm høy load på CICSP460, IKT-479284 DVH, IKT-479282 KUHR",
							},
							{
								StemplingTid:       "2022-10-26T00:45:35",
								Retning:            "Ut",
								Type:               "B2",
								Fravarkode:         0,
								OvertidBegrunnelse: "",
							},
						},
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 500_000,
							},
						},
					},
					{
						Dato:       "2022-10-26T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []models.MWTStempling{
							{
								StemplingTid:       "2022-10-26T08:00:00",
								Retning:            "Inn",
								Type:               "B1",
								Fravarkode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-26T15:45:00",
								Retning:            "Ut",
								Type:               "B2",
								Fravarkode:         0,
								OvertidBegrunnelse: "",
							},
						},
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 500_000,
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
				days: []models.MWTDag{
					{
						Dato:       "2022-10-18T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []models.MWTStempling{
							{
								StemplingTid: "2022-10-18T08:30:00",
								Retning:      "Inn",
								Type:         "B1",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2022-10-18T17:00:00",
								Retning:      "Ut",
								Type:         "B2",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2022-10-18T20:00:00",
								Retning:      "Inn",
								Type:         "B1",
								Fravarkode:   0,
							},
							{
								StemplingTid:       "2022-10-18T20:59:59",
								Retning:            "Overtid",
								Type:               "B6",
								Fravarkode:         0,
								OvertidBegrunnelse: "BV",
							},
							{
								StemplingTid: "2022-10-18T21:00:00",
								Retning:      "Ut",
								Type:         "B2",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2022-10-18T23:30:00",
								Retning:      "Inn",
								Type:         "B1",
								Fravarkode:   0,
							},
							{
								StemplingTid:       "2022-10-19T00:29:59",
								Retning:            "Overtid",
								Type:               "B6",
								Fravarkode:         1,
								OvertidBegrunnelse: "BV",
							},
							{
								StemplingTid: "2022-10-19T00:30:00",
								Retning:      "Ut",
								Type:         "B2",
								Fravarkode:   0,
							},
						},
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 500_000,
							},
						},
					},
					{
						Dato:       "2022-10-19T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []models.MWTStempling{
							{
								StemplingTid: "2022-10-19T08:00:00",
								Retning:      "Inn",
								Type:         "B1",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2022-10-19T17:00:00",
								Retning:      "Ut",
								Type:         "B2",
								Fravarkode:   0,
							},
						},
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 500_000,
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
				days: []models.MWTDag{
					{
						Dato:       "2022-01-20T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []models.MWTStempling{
							{
								StemplingTid: "2022-01-20T08:09:00",
								Retning:      "Inn",
								Type:         "B1",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2022-01-20T14:34:00",
								Retning:      "Ut på fravær",
								Type:         "B5",
								Fravarkode:   470,
							},
						},
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 500_000,
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
				days: []models.MWTDag{
					{
						Dato:       "2022-01-24T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "NY BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   3,
						Virkedag:   "Virkedag",
						Stemplinger: []models.MWTStempling{
							{
								StemplingTid: "2022-01-24T08:27:00",
								Retning:      "Inn",
								Type:         "B1",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2022-01-24T10:01:00",
								Retning:      "Ut på fravær",
								Type:         "B5",
								Fravarkode:   180,
							},
							{
								StemplingTid: "2022-01-24T11:27:00",
								Retning:      "Inn",
								Type:         "B1",
								Fravarkode:   0,
							},
							{
								StemplingTid: "2022-01-24T15:45:00",
								Retning:      "Ut",
								Type:         "B2",
								Fravarkode:   0,
							},
						},
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 500_000,
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
							Out: time.Date(2022, 1, 24, 10, 1, 0, 0, time.UTC),
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
				days: []models.MWTDag{
					{
						Dato:       "2022-10-05T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   2,
						Virkedag:   "Virkedag",
						Stemplinger: []models.MWTStempling{
							{
								StemplingTid:       "2022-10-05T07:21:42",
								Retning:            "Inn",
								Type:               "B1",
								Fravarkode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-05T15:24:14",
								Retning:            "Ut",
								Type:               "B2",
								Fravarkode:         0,
								OvertidBegrunnelse: "",
							},
						},
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 725000,
							},
						},
					},
					{
						Dato:       "2022-10-06T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   2,
						Virkedag:   "Virkedag",
						Stemplinger: []models.MWTStempling{
							{
								StemplingTid:       "2022-10-06T07:13:24",
								Retning:            "Inn",
								Type:               "B1",
								Fravarkode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-06T15:03:51",
								Retning:            "Ut",
								Type:               "B2",
								Fravarkode:         0,
								OvertidBegrunnelse: "",
							},
						},
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 725000,
							},
						},
					},
					{
						Dato:       "2022-10-07T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   2,
						Virkedag:   "Virkedag",
						Stemplinger: []models.MWTStempling{
							{
								StemplingTid:       "2022-10-07T07:18:52",
								Retning:            "Inn",
								Type:               "B1",
								Fravarkode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-07T15:06:59",
								Retning:            "Ut",
								Type:               "B2",
								Fravarkode:         0,
								OvertidBegrunnelse: "",
							},
						},
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 725000,
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
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 725000,
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
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 725000,
							},
						},
					},
					{
						Dato:       "2022-10-10T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   2,
						Virkedag:   "Virkedag",
						Stemplinger: []models.MWTStempling{
							{
								StemplingTid:       "2022-10-10T07:18:32",
								Retning:            "Inn",
								Type:               "B1",
								Fravarkode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-10T15:25:00",
								Retning:            "Ut",
								Type:               "B2",
								Fravarkode:         0,
								OvertidBegrunnelse: "",
							},
						},
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 725000,
							},
						},
					},
					{
						Dato:       "2022-10-11T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   2,
						Virkedag:   "Virkedag",
						Stemplinger: []models.MWTStempling{
							{
								StemplingTid:       "2022-10-11T07:09:58",
								Retning:            "Inn",
								Type:               "B1",
								Fravarkode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-11T15:23:41",
								Retning:            "Ut",
								Type:               "B2",
								Fravarkode:         0,
								OvertidBegrunnelse: "",
							},
						},
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 725000,
							},
						},
					},
					{
						Dato:       "2022-10-12T00:00:00",
						SkjemaTid:  7.75,
						SkjemaNavn: "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Godkjent:   2,
						Virkedag:   "Virkedag",
						Stemplinger: []models.MWTStempling{
							{
								StemplingTid:       "2022-10-12T08:00:00",
								Retning:            "Inn",
								Type:               "B1",
								Fravarkode:         0,
								OvertidBegrunnelse: "",
							},
							{
								StemplingTid:       "2022-10-12T09:00:00",
								Retning:            "Ut",
								Type:               "B2",
								Fravarkode:         0,
								OvertidBegrunnelse: "",
							},
						},
						Stillinger: []models.MWTStilling{
							{
								RATEK001: 725000,
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
							In:  time.Date(2022, 10, 5, 7, 21, 42, 0, time.UTC),
							Out: time.Date(2022, 10, 5, 15, 24, 14, 0, time.UTC),
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
							In:  time.Date(2022, 10, 6, 7, 13, 24, 0, time.UTC),
							Out: time.Date(2022, 10, 6, 15, 3, 51, 0, time.UTC),
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
							In:  time.Date(2022, 10, 7, 7, 18, 52, 0, time.UTC),
							Out: time.Date(2022, 10, 7, 15, 6, 59, 0, time.UTC),
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
							In:  time.Date(2022, 10, 10, 7, 18, 32, 0, time.UTC),
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
					Plan:        json.RawMessage(`{"id":"b4ac8e53-9d64-4557-8ef8-d00774ab9c06","user_id":"E123456","start_timestamp":"2022-10-05T12:00:00Z","end_timestamp":"2022-10-12T12:00:00Z","schedule":{"2022-10-05":[{"start_timestamp":"2022-10-05T12:00:00Z","end_timestamp":"2022-10-06T00:00:00Z"}],"2022-10-06":[{"start_timestamp":"2022-10-06T00:00:00Z","end_timestamp":"2022-10-07T00:00:00Z"}],"2022-10-07":[{"start_timestamp":"2022-10-07T00:00:00Z","end_timestamp":"2022-10-08T00:00:00Z"}],"2022-10-08":[{"start_timestamp":"2022-10-08T00:00:00Z","end_timestamp":"2022-10-09T00:00:00Z"}],"2022-10-09":[{"start_timestamp":"2022-10-09T00:00:00Z","end_timestamp":"2022-10-10T00:00:00Z"}],"2022-10-10":[{"start_timestamp":"2022-10-10T00:00:00Z","end_timestamp":"2022-10-11T00:00:00Z"}],"2022-10-11":[{"start_timestamp":"2022-10-11T00:00:00Z","end_timestamp":"2022-10-12T00:00:00Z"}],"2022-10-12":[{"start_timestamp":"2022-10-12T00:00:00Z","end_timestamp":"2022-10-12T12:00:00Z"}]}}`),
					PeriodBegin: time.Date(2022, 10, 5, 12, 0, 0, 0, time.UTC),
					PeriodEnd:   time.Date(2022, 10, 12, 12, 0, 0, 0, time.UTC),
				},
			},
			want: want{
				payroll: &models.Payroll{
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
					Stillingskode: "258",
				},
			},
		},

		{
			name: "Vanlig ukesvakt med litt overtid",
			args: args{
				beredskapsvakt: gensql.Beredskapsvakt{
					Ident:       "a123456",
					Plan:        json.RawMessage(`{"id":"b4ac8e53-9d64-4557-8ef8-d00774ab9c06","user_id":"E123456","start_timestamp":"2022-10-12T12:00:00Z","end_timestamp":"2022-10-19T12:00:00Z","schedule":{"2022-10-12":[{"start_timestamp":"2022-10-12T12:00:00Z","end_timestamp":"2022-10-13T00:00:00Z"}],"2022-10-13":[{"start_timestamp":"2022-10-13T00:00:00Z","end_timestamp":"2022-10-14T00:00:00Z"}],"2022-10-14":[{"start_timestamp":"2022-10-14T00:00:00Z","end_timestamp":"2022-10-15T00:00:00Z"}],"2022-10-15":[{"start_timestamp":"2022-10-15T00:00:00Z","end_timestamp":"2022-10-16T00:00:00Z"}],"2022-10-16":[{"start_timestamp":"2022-10-16T00:00:00Z","end_timestamp":"2022-10-17T00:00:00Z"}],"2022-10-17":[{"start_timestamp":"2022-10-17T00:00:00Z","end_timestamp":"2022-10-18T00:00:00Z"}],"2022-10-18":[{"start_timestamp":"2022-10-18T00:00:00Z","end_timestamp":"2022-10-19T00:00:00Z"}],"2022-10-19":[{"start_timestamp":"2022-10-19T00:00:00Z","end_timestamp":"2022-10-19T12:00:00Z"}]}}`),
					PeriodBegin: time.Date(2022, 10, 12, 12, 0, 0, 0, time.UTC),
					PeriodEnd:   time.Date(2022, 10, 9, 12, 0, 0, 0, time.UTC),
				},
			},
			want: want{
				payroll: &models.Payroll{
					ID:           uuid.MustParse("b4ac8e53-9d64-4557-8ef8-d00774ab9c06"),
					ApproverID:   "M654321",
					ApproverName: "Kalpana, Bran",
					Artskoder: models.Artskoder{
						Morgen: models.Artskode{
							Sum:   decimal.NewFromFloat(7002.97),
							Hours: 30,
						},
						Kveld: models.Artskode{
							Sum:   decimal.NewFromFloat(4668.65),
							Hours: 20,
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
							Sum:   decimal.NewFromFloat(0),
							Hours: 0,
						},
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
					Plan:        json.RawMessage(`{"id":"b4ac8e53-9d64-4557-8ef8-d00774ab9c06","user_id":"E123456","start_timestamp":"2022-10-26T12:00:00Z","end_timestamp":"2022-11-01T00:00:00Z","schedule":{"2022-10-26":[{"start_timestamp":"2022-10-26T12:00:00Z","end_timestamp":"2022-10-27T00:00:00Z"}],"2022-10-27":[{"start_timestamp":"2022-10-27T00:00:00Z","end_timestamp":"2022-10-28T00:00:00Z"}],"2022-10-28":[{"start_timestamp":"2022-10-28T00:00:00Z","end_timestamp":"2022-10-29T00:00:00Z"}],"2022-10-29":[{"start_timestamp":"2022-10-29T00:00:00Z","end_timestamp":"2022-10-30T00:00:00Z"}],"2022-10-30":[{"start_timestamp":"2022-10-30T00:00:00Z","end_timestamp":"2022-10-31T00:00:00Z"}],"2022-10-31":[{"start_timestamp":"2022-10-31T00:00:00Z","end_timestamp":"2022-11-01T00:00:00Z"}]}}`),
					PeriodBegin: time.Date(2022, 10, 26, 12, 0, 0, 0, time.UTC),
					PeriodEnd:   time.Date(2022, 11, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			want: want{
				payroll: &models.Payroll{
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
					Stillingskode: "258",
				},
			},
		},

		{
			name: "Helg med overtid ikke merket med bv",
			args: args{
				beredskapsvakt: gensql.Beredskapsvakt{
					Ident:       "a123456",
					Plan:        json.RawMessage(`{"id":"b4ac8e53-9d64-4557-8ef8-d00774ab9c06","user_id":"E123456","start_timestamp":"2022-10-15T00:00:00Z","end_timestamp":"2022-10-16T00:00:00Z","schedule":{"2022-10-15":[{"start_timestamp":"2022-10-15T00:00:00Z","end_timestamp":"2022-10-16T00:00:00Z"}],"2022-10-16":[{"start_timestamp":"2022-10-16T00:00:00Z","end_timestamp":"2022-10-17T00:00:00Z"}]}}`),
					PeriodBegin: time.Date(2022, 10, 15, 0, 0, 0, 0, time.UTC),
					PeriodEnd:   time.Date(2022, 10, 16, 0, 0, 0, 0, time.UTC),
				},
			},
			want: want{
				payroll: &models.Payroll{
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
							Sum:   decimal.NewFromFloat(8151.91),
							Hours: 48,
						},
						Utrykning: models.Artskode{
							Sum:   decimal.NewFromFloat(130),
							Hours: 2,
						},
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
					Plan:        json.RawMessage(`{"id":"b4ac8e53-9d64-4557-8ef8-d00774ab9c06","user_id":"E123456","start_timestamp":"2022-10-15T00:00:00Z","end_timestamp":"2022-10-16T00:00:00Z","schedule":{"2022-10-10":[{"start_timestamp":"2022-10-10T12:00:00Z","end_timestamp":"2022-10-11T00:00:00Z"}],"2022-10-11":[{"start_timestamp":"2022-10-11T00:00:00Z","end_timestamp":"2022-10-12T00:00:00Z"}],"2022-10-12":[{"start_timestamp":"2022-10-12T00:00:00Z","end_timestamp":"2022-10-13T00:00:00Z"}],"2022-10-13":[{"start_timestamp":"2022-10-13T00:00:00Z","end_timestamp":"2022-10-14T00:00:00Z"}],"2022-10-14":[{"start_timestamp":"2022-10-14T00:00:00Z","end_timestamp":"2022-10-15T00:00:00Z"}],"2022-10-15":[{"start_timestamp":"2022-10-15T00:00:00Z","end_timestamp":"2022-10-16T00:00:00Z"}],"2022-10-16":[{"start_timestamp":"2022-10-16T00:00:00Z","end_timestamp":"2022-10-17T00:00:00Z"}],"2022-10-17":[{"start_timestamp":"2022-10-17T00:00:00Z","end_timestamp":"2022-10-18T12:00:00Z"}]}}`),
					PeriodBegin: time.Date(2022, 10, 10, 0, 0, 0, 0, time.UTC),
					PeriodEnd:   time.Date(2022, 10, 17, 0, 0, 0, 0, time.UTC),
				},
			},
			want: want{
				payroll: &models.Payroll{
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
							Sum:   decimal.NewFromFloat(8151.91),
							Hours: 48,
						},
						Skift: models.Artskode{
							Sum:   decimal.NewFromFloat(115),
							Hours: 23,
						},
						Utrykning: models.Artskode{
							Sum:   decimal.NewFromFloat(130),
							Hours: 2,
						},
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
					Plan:        json.RawMessage(`{"id":"b4ac8e53-9d64-4557-8ef8-d00774ab9c06","user_id":"E123456","start_timestamp":"2022-12-31T00:00:00Z","end_timestamp":"2023-01-01T00:00:00Z","schedule":{"2022-12-31":[{"start_timestamp":"2022-12-31T00:00:00Z","end_timestamp":"2023-01-01T00:00:00Z"}]}}`),
					PeriodBegin: time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC),
					PeriodEnd:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			want: want{
				payroll: &models.Payroll{
					ID:           uuid.MustParse("b4ac8e53-9d64-4557-8ef8-d00774ab9c06"),
					ApproverID:   "M654321",
					ApproverName: "Kalpana, Bran",
					Artskoder: models.Artskoder{
						Helg: models.Artskode{
							Sum:   decimal.NewFromFloat(3366.59),
							Hours: 24,
						},
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
					Plan:        json.RawMessage(`{"id":"b4ac8e53-9d64-4557-8ef8-d00774ab9c06","user_id":"E123456","start_timestamp":"2023-06-17T23:00:00Z","end_timestamp":"2023-06-18T12:00:00Z","schedule":{"2023-06-17":[{"start_timestamp":"2023-06-17T23:00:00Z","end_timestamp":"2023-06-18T00:00:00Z"}],"2023-06-18":[{"start_timestamp":"2023-06-18T00:00:00Z","end_timestamp":"2023-06-18T12:00:00Z"}]}}`),
					PeriodBegin: time.Date(2023, 6, 17, 0, 0, 0, 0, time.UTC),
					PeriodEnd:   time.Date(2023, 6, 18, 0, 0, 0, 0, time.UTC),
				},
			},
			want: want{
				payroll: &models.Payroll{
					ID:           uuid.MustParse("b4ac8e53-9d64-4557-8ef8-d00774ab9c06"),
					ApproverID:   "M654321",
					ApproverName: "Kalpana, Bran",
					Artskoder: models.Artskoder{
						Helg: models.Artskode{
							Sum:   decimal.NewFromFloat(406),
							Hours: 12,
						},
						Utrykning: models.Artskode{
							Sum:   decimal.NewFromFloat(390),
							Hours: 6,
						},
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
				days: []models.MWTDag{
					{
						Godkjent: 0,
					},
				},
			},
			error: true,
		},
		{
			name: "Godkjent av vakthaver",
			args: args{
				days: []models.MWTDag{
					{
						Godkjent: 1,
					},
				},
			},
			error: true,
		},
		{
			name: "Godkjent av personalleder",
			args: args{
				days: []models.MWTDag{
					{
						Godkjent: 2,
					},
				},
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
