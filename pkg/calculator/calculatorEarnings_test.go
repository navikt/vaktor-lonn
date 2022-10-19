package calculator

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/navikt/vaktor-lonn/pkg/compensation"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/navikt/vaktor-lonn/pkg/overtime"
	"github.com/shopspring/decimal"
	"testing"
	"time"
)

func TestCalculateEarnings(t *testing.T) {
	type args struct {
		satser      map[string]decimal.Decimal
		timesheet   map[string]models.TimeSheet
		guardPeriod map[string][]models.Period
	}
	tests := []struct {
		name    string
		args    args
		want    decimal.Decimal
		wantErr bool
	}{
		{
			name: "døgnvakt",
			args: args{
				satser: map[string]decimal.Decimal{
					"lørsøn":  decimal.NewFromInt(55),
					"0620":    decimal.NewFromInt(10),
					"2006":    decimal.NewFromInt(20),
					"utvidet": decimal.NewFromInt(15),
				},
				timesheet: map[string]models.TimeSheet{
					"2022-03-14": {
						Date:         time.Date(2022, 3, 14, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 3, 14, 7, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 3, 14, 15, 0, 0, 0, time.UTC),
							},
						},
					},
					"2022-03-15": {
						Date:         time.Date(2022, 3, 15, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 3, 15, 7, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 3, 15, 16, 0, 0, 0, time.UTC),
							},
						},
						//"07:00-16:00"
					},
					"2022-03-16": {
						Date:         time.Date(2022, 3, 16, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 3, 16, 7, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 3, 16, 15, 0, 0, 0, time.UTC),
							},
						},
					},
					"2022-03-17": {
						Date:         time.Date(2022, 3, 17, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 3, 17, 8, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 3, 17, 16, 0, 0, 0, time.UTC),
							},
						},
					},
					"2022-03-18": {
						Date:         time.Date(2022, 3, 18, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 3, 18, 7, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 3, 18, 16, 0, 0, 0, time.UTC),
							},
						},
					},
					"2022-03-19": {
						Date:         time.Date(2022, 3, 19, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Lørdag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings:    []models.Clocking{},
					},
					"2022-03-20": {
						Date:         time.Date(2022, 3, 20, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Søndag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings:    []models.Clocking{},
					},
				},
				guardPeriod: map[string][]models.Period{
					"2022-03-14": {
						{
							Begin: time.Date(2022, 3, 14, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 3, 15, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-03-15": {
						{
							Begin: time.Date(2022, 3, 15, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 3, 16, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-03-16": {
						{
							Begin: time.Date(2022, 3, 16, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 3, 17, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-03-17": {
						{
							Begin: time.Date(2022, 3, 17, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 3, 18, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-03-18": {
						{
							Begin: time.Date(2022, 3, 18, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 3, 19, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-03-19": {
						{
							Begin: time.Date(2022, 3, 19, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 3, 20, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-03-20": {
						{
							Begin: time.Date(2022, 3, 20, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 3, 21, 0, 0, 0, 0, time.UTC),
						},
					},
				},
			},
			want: decimal.NewFromFloat(15_412.86),
		},

		{
			name: "Utvidet beredskap",
			args: args{
				satser: map[string]decimal.Decimal{
					"lørsøn":  decimal.NewFromInt(55),
					"0620":    decimal.NewFromInt(10),
					"2006":    decimal.NewFromInt(20),
					"utvidet": decimal.NewFromInt(15),
				},
				timesheet: map[string]models.TimeSheet{
					"2022-07-04": {
						Date:         time.Date(2022, 7, 4, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(800_000),
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 7, 4, 9, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 7, 4, 15, 0, 0, 0, time.UTC),
							},
						},
					},
					"2022-07-05": {
						Date:         time.Date(2022, 7, 5, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(800_000),
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 7, 5, 9, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 7, 5, 15, 0, 0, 0, time.UTC),
							},
						},
					},
					"2022-07-06": {
						Date:         time.Date(2022, 7, 6, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(800_000),
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 7, 6, 9, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 7, 6, 15, 30, 0, 0, time.UTC),
							},
						},
					},
					"2022-07-07": {
						Date:         time.Date(2022, 7, 7, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(800_000),
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 7, 7, 9, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 7, 7, 15, 0, 0, 0, time.UTC),
							},
						},
					},
					"2022-07-08": {
						Date:         time.Date(2022, 7, 8, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(800_000),
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 7, 8, 9, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 7, 8, 15, 30, 0, 0, time.UTC),
							},
						},
					},

					"2022-07-11": {
						Date:         time.Date(2022, 7, 11, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(800_000),
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 7, 11, 8, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 7, 11, 16, 0, 0, 0, time.UTC),
							},
						},
					},
					"2022-07-12": {
						Date:         time.Date(2022, 7, 12, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(800_000),
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 7, 12, 8, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 7, 12, 16, 0, 0, 0, time.UTC),
							},
						},
					},
					"2022-07-13": {
						Date:         time.Date(2022, 7, 13, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(800_000),
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 7, 13, 8, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 7, 13, 16, 0, 0, 0, time.UTC),
							},
						},
					},
					"2022-07-14": {
						Date:         time.Date(2022, 7, 14, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(800_000),
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 7, 14, 8, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 7, 14, 16, 0, 0, 0, time.UTC),
							},
						},
					},
					"2022-07-15": {
						Date:         time.Date(2022, 7, 15, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(800_000),
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 7, 15, 8, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 7, 15, 16, 0, 0, 0, time.UTC),
							},
						},
					},
					"2022-07-16": {Date: time.Date(2022, 7, 16, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Lørdag",
						Salary:       decimal.NewFromInt(800_000),
						Clockings:    []models.Clocking{},
					},
					"2022-07-17": {
						Date:         time.Date(2022, 7, 17, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Søndag",
						Salary:       decimal.NewFromInt(800_000),
						Clockings:    []models.Clocking{},
					},
				},
				guardPeriod: map[string][]models.Period{
					"2022-07-04": {
						{
							Begin: time.Date(2022, 7, 4, 6, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 4, 9, 0, 0, 0, time.UTC),
						},
						{
							Begin: time.Date(2022, 7, 4, 15, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 4, 21, 0, 0, 0, time.UTC),
						},
					},
					"2022-07-05": {
						{
							Begin: time.Date(2022, 7, 5, 6, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 5, 9, 0, 0, 0, time.UTC),
						},
						{
							Begin: time.Date(2022, 7, 5, 15, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 5, 21, 0, 0, 0, time.UTC),
						},
					},
					"2022-07-06": {
						{
							Begin: time.Date(2022, 7, 6, 6, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 6, 9, 0, 0, 0, time.UTC),
						},
						{
							Begin: time.Date(2022, 7, 6, 15, 30, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 6, 21, 0, 0, 0, time.UTC),
						},
					},
					"2022-07-07": {
						{
							Begin: time.Date(2022, 7, 7, 6, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 7, 9, 0, 0, 0, time.UTC),
						},
						{
							Begin: time.Date(2022, 7, 7, 15, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 7, 21, 0, 0, 0, time.UTC),
						},
					},
					"2022-07-08": {
						{
							Begin: time.Date(2022, 7, 8, 6, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 8, 9, 0, 0, 0, time.UTC),
						},
						{
							Begin: time.Date(2022, 7, 8, 15, 30, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 8, 21, 0, 0, 0, time.UTC),
						},
					},

					"2022-07-11": {
						{
							Begin: time.Date(2022, 7, 11, 6, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 11, 8, 0, 0, 0, time.UTC),
						},
						{
							Begin: time.Date(2022, 7, 11, 16, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 11, 21, 0, 0, 0, time.UTC),
						},
					},
					"2022-07-12": {
						{
							Begin: time.Date(2022, 7, 12, 6, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 12, 8, 0, 0, 0, time.UTC),
						},
						{
							Begin: time.Date(2022, 7, 12, 16, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 12, 21, 0, 0, 0, time.UTC),
						},
					},
					"2022-07-13": {
						{
							Begin: time.Date(2022, 7, 13, 6, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 13, 8, 0, 0, 0, time.UTC),
						},
						{
							Begin: time.Date(2022, 7, 13, 16, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 13, 21, 0, 0, 0, time.UTC),
						},
					},
					"2022-07-14": {
						{
							Begin: time.Date(2022, 7, 14, 6, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 14, 8, 0, 0, 0, time.UTC),
						},
						{
							Begin: time.Date(2022, 7, 14, 16, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 14, 21, 0, 0, 0, time.UTC),
						},
					},
					"2022-07-15": {
						{
							Begin: time.Date(2022, 7, 15, 6, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 15, 8, 0, 0, 0, time.UTC),
						},
						{
							Begin: time.Date(2022, 7, 15, 16, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 15, 21, 0, 0, 0, time.UTC),
						},
					},
					"2022-07-16": {
						{
							Begin: time.Date(2022, 7, 16, 9, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 16, 15, 0, 0, 0, time.UTC),
						},
					},
					"2022-07-17": {
						{
							Begin: time.Date(2022, 7, 17, 9, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 17, 15, 0, 0, 0, time.UTC),
						},
					},
				},
			},
			want: decimal.NewFromFloat(14_018.76),
		},

		{
			name: "Vakt ved spesielle hendelser",
			args: args{
				satser: map[string]decimal.Decimal{
					"lørsøn":  decimal.NewFromInt(55),
					"0620":    decimal.NewFromInt(10),
					"2006":    decimal.NewFromInt(20),
					"utvidet": decimal.NewFromInt(15),
				},
				// TODO: Mangler vi arbeidstid her?
				timesheet: map[string]models.TimeSheet{
					"2022-07-16": {
						Date:         time.Date(2022, 7, 16, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Lørdag",
						Salary:       decimal.NewFromInt(800_000),
						Clockings:    []models.Clocking{},
					},
					"2022-07-17": {
						Date:         time.Date(2022, 7, 17, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Søndag",
						Salary:       decimal.NewFromInt(800_000),
						Clockings:    []models.Clocking{},
					},
					"2022-07-22": {
						Date:         time.Date(2022, 7, 22, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(800_000),
						Clockings:    []models.Clocking{},
					},
					"2022-07-23": {
						Date:         time.Date(2022, 7, 23, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Lørdag",
						Salary:       decimal.NewFromInt(800_000),
						Clockings:    []models.Clocking{},
					},
					"2022-07-24": {
						Date:         time.Date(2022, 7, 24, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Søndag",
						Salary:       decimal.NewFromInt(800_000),
						Clockings:    []models.Clocking{},
					},
					"2022-07-25": {
						Date:         time.Date(2022, 7, 25, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(800_000),
						Clockings:    []models.Clocking{},
					},
				},
				guardPeriod: map[string][]models.Period{
					"2022-07-16": {
						{
							Begin: time.Date(2022, 7, 16, 17, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 17, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-07-17": {
						{
							Begin: time.Date(2022, 7, 17, 7, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 17, 16, 0, 0, 0, time.UTC),
						},
					},
					"2022-07-22": {
						{
							Begin: time.Date(2022, 7, 22, 16, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 23, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-07-23": {
						{
							Begin: time.Date(2022, 7, 23, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 24, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-07-24": {
						{
							Begin: time.Date(2022, 7, 24, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 25, 0, 0, 0, 0, time.UTC),
						},
					},

					"2022-07-25": {
						{
							Begin: time.Date(2022, 7, 25, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 7, 25, 7, 0, 0, 0, time.UTC),
						},
					},
				},
			},
			want: decimal.NewFromFloat(15_294.65),
		},

		{
			name: "Utvidet beredskap",
			args: args{
				satser: map[string]decimal.Decimal{
					"lørsøn":  decimal.NewFromInt(55),
					"0620":    decimal.NewFromInt(10),
					"2006":    decimal.NewFromInt(20),
					"utvidet": decimal.NewFromInt(15),
				},
				timesheet: map[string]models.TimeSheet{
					"2022-10-30": {
						Date:         time.Date(2022, 10, 30, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Søndag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings:    []models.Clocking{},
					},
				},
				guardPeriod: map[string][]models.Period{
					"2022-10-30": {
						{
							Begin: time.Date(2022, 10, 30, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 31, 0, 0, 0, 0, time.UTC),
						},
					},
				},
			},
			want: decimal.NewFromFloat(3_337.70),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minutes, err := calculateMinutesToBeCompensated(tt.args.guardPeriod, tt.args.timesheet)
			if err != nil {
				t.Errorf("calculateMinutesToBeCompensated() error : %v", err)
				return
			}

			var payroll *models.Payroll
			payroll = &models.Payroll{
				ID:         uuid.UUID{},
				ResourceID: "123456",
				Approver:   "Scathan",
				TypeCodes: map[string]decimal.Decimal{
					models.ArtskodeMorgen: {},
					models.ArtskodeDag:    {},
					models.ArtskodeKveld:  {},
					models.ArtskodeHelg:   {},
				},
			}

			compensation.Calculate(minutes, tt.args.satser, *payroll)
			err = overtime.Calculate(minutes, tt.args.timesheet, payroll)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateEarnings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			total := decimal.Decimal{}
			for _, typeCode := range payroll.TypeCodes {
				total = total.Add(typeCode)
			}

			if !total.Equal(tt.want) {
				t.Errorf("calculateEarnings() got = %v, want %v", total, tt.want)
				fmt.Printf("Morgen: %v, Dag: %v, Kveld: %v, Helg: %v\n", payroll.TypeCodes[models.ArtskodeMorgen],
					payroll.TypeCodes[models.ArtskodeDag], payroll.TypeCodes[models.ArtskodeKveld], payroll.TypeCodes[models.ArtskodeHelg])
			}
		})
	}
}
