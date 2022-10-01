package calculator

import (
	"github.com/navikt/vaktor-lonn/pkg/compensation"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/navikt/vaktor-lonn/pkg/overtime"
	"github.com/shopspring/decimal"
	"testing"
	"time"
)

func TestCalculateEarnings(t *testing.T) {
	type args struct {
		minutes map[string]models.GuardDuty
		salary  decimal.Decimal
	}
	tests := []struct {
		name      string
		args      args
		minWinTid map[string][]string
		pocPeriod map[string][]models.Period
		want      decimal.Decimal
		wantErr   bool
	}{
		{
			name: "døgnvakt",
			args: args{
				salary: decimal.NewFromInt(500_000),
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
						Begin: time.Date(2022, 3, 14, 0, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 3, 15, 0, 0, 0, 0, time.UTC),
					},
				},
				"15.03.2022": {
					{
						Begin: time.Date(2022, 3, 15, 0, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 3, 16, 0, 0, 0, 0, time.UTC),
					},
				},
				"16.03.2022": {
					{
						Begin: time.Date(2022, 3, 16, 0, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 3, 17, 0, 0, 0, 0, time.UTC),
					},
				},
				"17.03.2022": {
					{
						Begin: time.Date(2022, 3, 17, 0, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 3, 18, 0, 0, 0, 0, time.UTC),
					},
				},
				"18.03.2022": {
					{
						Begin: time.Date(2022, 3, 18, 0, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 3, 19, 0, 0, 0, 0, time.UTC),
					},
				},
				"19.03.2022": {
					{
						Begin: time.Date(2022, 3, 19, 0, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 3, 20, 0, 0, 0, 0, time.UTC),
					},
				},
				"20.03.2022": {
					{
						Begin: time.Date(2022, 3, 20, 0, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 3, 21, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			want: decimal.NewFromFloat(15_412.86),
		},

		{
			name: "Utvidet beredskap",
			args: args{
				salary: decimal.NewFromInt(800_000),
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
						Begin: time.Date(2022, 7, 4, 6, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 4, 9, 0, 0, 0, time.UTC),
					},
					{
						Begin: time.Date(2022, 7, 4, 15, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 4, 21, 0, 0, 0, time.UTC),
					},
				},
				"05.07.2022": {
					{
						Begin: time.Date(2022, 7, 5, 6, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 5, 9, 0, 0, 0, time.UTC),
					},
					{
						Begin: time.Date(2022, 7, 5, 15, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 5, 21, 0, 0, 0, time.UTC),
					},
				},
				"06.07.2022": {
					{
						Begin: time.Date(2022, 7, 6, 6, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 6, 9, 0, 0, 0, time.UTC),
					},
					{
						Begin: time.Date(2022, 7, 6, 15, 30, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 6, 21, 0, 0, 0, time.UTC),
					},
				},
				"07.07.2022": {
					{
						Begin: time.Date(2022, 7, 7, 6, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 7, 9, 0, 0, 0, time.UTC),
					},
					{
						Begin: time.Date(2022, 7, 7, 15, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 7, 21, 0, 0, 0, time.UTC),
					},
				},
				"08.07.2022": {
					{
						Begin: time.Date(2022, 7, 8, 6, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 8, 9, 0, 0, 0, time.UTC),
					},
					{
						Begin: time.Date(2022, 7, 8, 15, 30, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 8, 21, 0, 0, 0, time.UTC),
					},
				},

				"11.07.2022": {
					{
						Begin: time.Date(2022, 7, 11, 6, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 11, 8, 0, 0, 0, time.UTC),
					},
					{
						Begin: time.Date(2022, 7, 11, 16, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 11, 21, 0, 0, 0, time.UTC),
					},
				},
				"12.07.2022": {
					{
						Begin: time.Date(2022, 7, 12, 6, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 12, 8, 0, 0, 0, time.UTC),
					},
					{
						Begin: time.Date(2022, 7, 12, 16, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 12, 21, 0, 0, 0, time.UTC),
					},
				},
				"13.07.2022": {
					{
						Begin: time.Date(2022, 7, 13, 6, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 13, 8, 0, 0, 0, time.UTC),
					},
					{
						Begin: time.Date(2022, 7, 13, 16, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 13, 21, 0, 0, 0, time.UTC),
					},
				},
				"14.07.2022": {
					{
						Begin: time.Date(2022, 7, 14, 6, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 14, 8, 0, 0, 0, time.UTC),
					},
					{
						Begin: time.Date(2022, 7, 14, 16, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 14, 21, 0, 0, 0, time.UTC),
					},
				},
				"15.07.2022": {
					{
						Begin: time.Date(2022, 7, 15, 6, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 15, 8, 0, 0, 0, time.UTC),
					},
					{
						Begin: time.Date(2022, 7, 15, 16, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 15, 21, 0, 0, 0, time.UTC),
					},
				},
				"16.07.2022": {
					{
						Begin: time.Date(2022, 7, 16, 9, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 16, 15, 0, 0, 0, time.UTC),
					},
				},
				"17.07.2022": {
					{
						Begin: time.Date(2022, 7, 17, 9, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 17, 15, 0, 0, 0, time.UTC),
					},
				},
			},
			want: decimal.NewFromFloat(14_018.76),
		},

		{
			name: "Vakt ved spesielle hendelser",
			args: args{
				salary: decimal.NewFromInt(800_000),
			},
			minWinTid: map[string][]string{},
			pocPeriod: map[string][]models.Period{
				"16.07.2022": {
					{
						Begin: time.Date(2022, 7, 16, 17, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 17, 0, 0, 0, 0, time.UTC),
					},
				},
				"17.07.2022": {
					{
						Begin: time.Date(2022, 7, 17, 7, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 17, 16, 0, 0, 0, time.UTC),
					},
				},
				"22.07.2022": {
					{
						Begin: time.Date(2022, 7, 22, 16, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 23, 0, 0, 0, 0, time.UTC),
					},
				},
				"23.07.2022": {
					{
						Begin: time.Date(2022, 7, 23, 0, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 24, 0, 0, 0, 0, time.UTC),
					},
				},
				"24.07.2022": {
					{
						Begin: time.Date(2022, 7, 24, 0, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 25, 0, 0, 0, 0, time.UTC),
					},
				},

				"25.07.2022": {
					{
						Begin: time.Date(2022, 7, 25, 0, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 7, 25, 7, 0, 0, 0, time.UTC),
					},
				},
			},
			want: decimal.NewFromFloat(15_294.65),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minutes, err := calculateMinutesToBeCompensated(tt.pocPeriod, tt.minWinTid)
			if err != nil {
				t.Errorf("calculateMinutesToBeCompensated() error : %v", err)
				return
			}
			tt.args.minutes = minutes
			compensationTotal := compensation.Calculate(minutes, tt.args.report.MinWinTid.Satser)
			overtimeTotal := overtime.Calculate(minutes, tt.args.salary)

			total := compensationTotal.Add(overtimeTotal)

			if !total.Equal(tt.want) {
				t.Errorf("calculateEarnings() got = %v, want %v", total, tt.want)
			}
		})
	}
}
