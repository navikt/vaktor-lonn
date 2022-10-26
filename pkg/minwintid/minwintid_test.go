package minwintid

import (
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/shopspring/decimal"
	"io"
	"net/http"
	"reflect"
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
			name: "helg med utrykning",
			args: args{
				days: []Dag{
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
			want: map[string]models.TimeSheet{
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
								StemplingTid: "2022-09-15T23:31:00",
								Retning:      "Overtid                 ",
								Type:         "B6",
								FravarKode:   0,
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
								StemplingTid: "2022-09-15T23:31:00",
								Retning:      "Overtid                 ",
								Type:         "B6",
								FravarKode:   0,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatTimesheet(tt.args.days)
			if (err != nil) != tt.wantErr {
				t.Errorf("formatTimesheet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("formatTimesheet() got = %v, want %v", got, tt.want)
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodeMinWinTid() got = %v, want %v", got, tt.want)
			}
		})
	}
}
