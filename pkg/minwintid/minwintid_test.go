package minwintid

import (
	"github.com/navikt/vaktor-lonn/pkg/models"
	"reflect"
	"testing"
	"time"
)

func Test_formatTimesheet(t *testing.T) {
	type args struct {
		days []models.Dag
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]models.TimeSheet
		wantErr bool
	}{
		{
			name: "",
			args: args{
				days: []models.Dag{
					{
						Dato:                 "2022-07-15T00:00:00",
						SkjemaTid:            7,
						SkjemaNavn:           "Heltid 0800-1500 (2018)",
						Godkjent:             3,
						AnsattDatoGodkjentAv: "a123456",
						GodkjentDato:         "2022-08-01T13:17:21",
						Virkedag:             "Virkedag",
						Stemplinger: []models.Stempling{
							{
								StemplingTid: "2022-07-15T15:00:00",
								Retning:      "B4",
								Type:         "B4",
							},
							{
								StemplingTid: "2022-07-15T15:00:01",
								Retning:      "Ut",
								Type:         "B2",
							},
						},
					},
				},
			},
			want: map[string]models.TimeSheet{
				"2022-07-15": {
					Date:                time.Date(2022, 7, 15, 0, 0, 0, 0, time.UTC),
					WorkingHours:        7,
					WeekendCompensation: false,
					Clockings:           []models.Clocking{},
				},
			},
			wantErr: false,
		},
		{
			name: "",
			args: args{
				days: []models.Dag{
					{
						Dato:                 "2022-08-02T00:00:00",
						SkjemaTid:            7,
						SkjemaNavn:           "Heltid 0800-1500 (2018)",
						Godkjent:             3,
						AnsattDatoGodkjentAv: "a123456",
						GodkjentDato:         "2022-09-01T10:32:41",
						Virkedag:             "Virkedag",
						Stemplinger: []models.Stempling{
							{
								StemplingTid: "2022-08-02T14:30:11",
								Retning:      "Ut",
								Type:         "B2",
							},
							{
								StemplingTid: "2022-08-02T16:00:01",
								Retning:      "Ut",
								Type:         "B2",
							},
							{
								StemplingTid: "2022-08-02T16:00:00",
								Retning:      "B4",
								Type:         "B4",
							},
							{
								StemplingTid: "2022-08-02T14:31:01",
								Retning:      "B5",
								Type:         "B5",
							},
							{
								StemplingTid: "2022-08-02T14:31:00",
								Retning:      "Inn",
								Type:         "B1",
							},
							{
								StemplingTid: "2022-08-02T07:45:10",
								Retning:      "Inn",
								Type:         "B1",
							},
						},
					},
				},
			},
			want: map[string]models.TimeSheet{
				"2022-08-02": {
					Date:                time.Date(2022, 8, 2, 0, 0, 0, 0, time.UTC),
					WorkingHours:        7,
					WeekendCompensation: false,
					Clockings: []models.Clocking{
						{
							In:  time.Date(2022, 8, 2, 7, 45, 10, 0, time.UTC),
							Out: time.Date(2022, 8, 2, 14, 30, 11, 0, time.UTC),
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
