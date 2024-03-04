package calculator

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/navikt/vaktor-lonn/pkg/models"
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
				satser: models.Satser{
					Helg:    decimal.NewFromInt(65),
					Dag:     decimal.NewFromInt(15),
					Natt:    decimal.NewFromInt(25),
					Utvidet: decimal.NewFromInt(25),
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
			want: want{
				sum: decimal.NewFromFloat(16_178.86),
			},
		},

		{
			name: "døgnvakt uten stemplinger",
			args: args{
				satser: models.Satser{
					Helg:    decimal.NewFromInt(65),
					Dag:     decimal.NewFromInt(15),
					Natt:    decimal.NewFromInt(25),
					Utvidet: decimal.NewFromInt(25),
				},
				timesheet: map[string]models.TimeSheet{
					"2022-03-14": {
						Date:         time.Date(2022, 3, 14, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings:    []models.Clocking{},
					},
					"2022-03-15": {
						Date:         time.Date(2022, 3, 15, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings:    []models.Clocking{},
					},
					"2022-03-16": {
						Date:         time.Date(2022, 3, 16, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings:    []models.Clocking{},
					},
					"2022-03-17": {
						Date:         time.Date(2022, 3, 17, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings:    []models.Clocking{},
					},
					"2022-03-18": {
						Date:         time.Date(2022, 3, 18, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings:    []models.Clocking{},
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
			want: want{
				sum: decimal.NewFromFloat(16_467.1),
			},
		},

		{
			name: "døgnvakt med perfekt stempling",
			args: args{
				satser: models.Satser{
					Helg:    decimal.NewFromInt(65),
					Dag:     decimal.NewFromInt(15),
					Natt:    decimal.NewFromInt(25),
					Utvidet: decimal.NewFromInt(25),
				},
				timesheet: map[string]models.TimeSheet{
					"2022-03-14": {
						Date:         time.Date(2022, 3, 14, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 3, 14, 8, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 3, 14, 15, 45, 0, 0, time.UTC),
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
								In:  time.Date(2022, 3, 15, 8, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 3, 15, 15, 45, 0, 0, time.UTC),
							},
						},
					},
					"2022-03-16": {
						Date:         time.Date(2022, 3, 16, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 3, 16, 8, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 3, 16, 15, 45, 0, 0, time.UTC),
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
								Out: time.Date(2022, 3, 17, 15, 45, 0, 0, time.UTC),
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
								In:  time.Date(2022, 3, 18, 8, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 3, 18, 15, 45, 0, 0, time.UTC),
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
			want: want{
				sum: decimal.NewFromFloat(16_467.1),
			},
		},

		{
			name: "Utvidet beredskap",
			args: args{
				satser: models.Satser{
					Helg:    decimal.NewFromInt(65),
					Dag:     decimal.NewFromInt(15),
					Natt:    decimal.NewFromInt(25),
					Utvidet: decimal.NewFromInt(25),
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
			want: want{
				sum: decimal.NewFromFloat(14_577.76),
			},
		},

		{
			name: "Vakt ved spesielle hendelser",
			args: args{
				satser: models.Satser{
					Helg:    decimal.NewFromInt(65),
					Dag:     decimal.NewFromInt(15),
					Natt:    decimal.NewFromInt(25),
					Utvidet: decimal.NewFromInt(25),
				},
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
			want: want{
				sum: decimal.NewFromFloat(15_825.65),
			},
		},

		{
			name: "Vakt når klokka stilles til sommertid",
			args: args{
				satser: models.Satser{
					Helg:    decimal.NewFromInt(65),
					Dag:     decimal.NewFromInt(15),
					Natt:    decimal.NewFromInt(25),
					Utvidet: decimal.NewFromInt(25),
				},
				timesheet: map[string]models.TimeSheet{
					"2022-03-27": {
						Date:         time.Date(2022, 3, 27, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Søndag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings:    []models.Clocking{},
					},
				},
				guardPeriod: map[string][]models.Period{
					"2022-03-27": {
						{
							Begin: time.Date(2022, 3, 27, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 3, 28, 0, 0, 0, 0, time.UTC),
						},
					},
				},
			},
			want: want{
				sum: decimal.NewFromFloat(3_220.49),
			},
		},

		{
			name: "Vakt når klokka stilles til normaltid",
			args: args{
				satser: models.Satser{
					Helg:    decimal.NewFromInt(65),
					Dag:     decimal.NewFromInt(15),
					Natt:    decimal.NewFromInt(25),
					Utvidet: decimal.NewFromInt(25),
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
			want: want{
				sum: decimal.NewFromFloat(3_512.70),
			},
		},

		{
			name: "Utvidet åpningstid dagen klokka stilles til normaltid",
			args: args{
				satser: models.Satser{
					Helg:    decimal.NewFromInt(65),
					Dag:     decimal.NewFromInt(15),
					Natt:    decimal.NewFromInt(25),
					Utvidet: decimal.NewFromInt(25),
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
							Begin: time.Date(2022, 10, 30, 8, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 30, 14, 0, 0, 0, time.UTC),
						},
					},
				},
			},
			want: want{
				sum: decimal.NewFromFloat(816.65),
			},
		},

		{
			name: "Helgevakt med utrykning",
			args: args{
				satser: models.Satser{
					Helg:    decimal.NewFromInt(65),
					Dag:     decimal.NewFromInt(15),
					Natt:    decimal.NewFromInt(25),
					Utvidet: decimal.NewFromInt(25),
				},
				timesheet: map[string]models.TimeSheet{
					"2022-09-24": {
						Date:         time.Date(2022, 9, 24, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Lørdag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 9, 24, 20, 30, 0, 0, time.UTC),
								Out: time.Date(2022, 9, 24, 22, 30, 0, 0, time.UTC),
								OtG: true,
							},
						},
					}, "2022-09-25": {
						Date:         time.Date(2022, 9, 25, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Søndag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings:    []models.Clocking{},
					},
				},
				guardPeriod: map[string][]models.Period{
					"2022-09-24": {
						{
							Begin: time.Date(2022, 9, 24, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 9, 25, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-09-25": {
						{
							Begin: time.Date(2022, 9, 25, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 9, 26, 0, 0, 0, 0, time.UTC),
						},
					},
				},
			},
			want: want{
				sum: decimal.NewFromFloat(6_837.19),
			},
		},

		{
			name: "Helgevakt uten utrykning",
			args: args{
				satser: models.Satser{
					Helg:    decimal.NewFromInt(65),
					Dag:     decimal.NewFromInt(15),
					Natt:    decimal.NewFromInt(25),
					Utvidet: decimal.NewFromInt(25),
				},
				timesheet: map[string]models.TimeSheet{
					"2022-09-24": {
						Date:         time.Date(2022, 9, 24, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						FormName:     "BV Lørdag IKT",
						WorkingDay:   "Lørdag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings:    []models.Clocking{},
					}, "2022-09-25": {
						Date:         time.Date(2022, 9, 25, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						FormName:     "BV Søndag IKT",
						WorkingDay:   "Søndag",
						Salary:       decimal.NewFromInt(500_000),
						Clockings:    []models.Clocking{},
					},
				},
				guardPeriod: map[string][]models.Period{
					"2022-09-24": {
						{
							Begin: time.Date(2022, 9, 24, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 9, 25, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-09-25": {
						{
							Begin: time.Date(2022, 9, 25, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 9, 26, 0, 0, 0, 0, time.UTC),
						},
					},
				},
			},
			want: want{
				sum: decimal.NewFromFloat(6_733.19),
			},
		},

		{
			name: "vakt en dag med utrykning",
			args: args{
				satser: models.Satser{
					Helg:    decimal.NewFromInt(65),
					Dag:     decimal.NewFromInt(15),
					Natt:    decimal.NewFromInt(25),
					Utvidet: decimal.NewFromInt(25),
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
							{
								In:  time.Date(2022, 3, 14, 20, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 3, 14, 22, 0, 0, 0, time.UTC),
								OtG: true,
							},
						},
					},
				},
				guardPeriod: map[string][]models.Period{
					"2022-03-14": {
						{
							Begin: time.Date(2022, 3, 14, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 3, 15, 0, 0, 0, 0, time.UTC),
						},
					},
				},
			},
			want: want{
				sum: decimal.NewFromFloat(1_927.57),
			},
		},

		{
			name: "En tilfeldig døgnkontinuerlig vaktuke",
			args: args{
				satser: models.Satser{
					Helg:    decimal.NewFromInt(65),
					Dag:     decimal.NewFromInt(15),
					Natt:    decimal.NewFromInt(25),
					Utvidet: decimal.NewFromInt(25),
				},
				timesheet: map[string]models.TimeSheet{
					"2022-10-05": {
						Date:         time.Date(2022, 10, 5, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						FormName:     "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Salary:       decimal.NewFromInt(725000),
						Koststed:     "000000",
						Formal:       "000000",
						Aktivitet:    "000000",
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
						Koststed:     "000000",
						Formal:       "000000",
						Aktivitet:    "000000",
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
						Koststed:     "000000",
						Formal:       "000000",
						Aktivitet:    "000000",
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
						Koststed:   "000000",
						Formal:     "000000",
						Aktivitet:  "000000",
						Clockings:  []models.Clocking{},
					},
					"2022-10-09": {
						Date:       time.Date(2022, 10, 9, 0, 0, 0, 0, time.UTC),
						WorkingDay: "Søndag",
						FormName:   "BV Søndag IKT",
						Salary:     decimal.NewFromInt(725000),
						Koststed:   "000000",
						Formal:     "000000",
						Aktivitet:  "000000",
						Clockings:  []models.Clocking{},
					},
					"2022-10-10": {
						Date:         time.Date(2022, 10, 10, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						FormName:     "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Salary:       decimal.NewFromInt(725000),
						Koststed:     "000000",
						Formal:       "000000",
						Aktivitet:    "000000",
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
						Koststed:     "000000",
						Formal:       "000000",
						Aktivitet:    "000000",
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
						Koststed:     "000000",
						Formal:       "000000",
						Aktivitet:    "000000",
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 10, 12, 8, 0, 0, 0, time.UTC),
								Out: time.Date(2022, 10, 12, 9, 0, 0, 0, time.UTC),
							},
						},
					},
				},
				guardPeriod: map[string][]models.Period{
					"2022-10-05": {
						{
							Begin: time.Date(2022, 10, 5, 12, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 6, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-10-06": {
						{
							Begin: time.Date(2022, 10, 6, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 7, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-10-07": {
						{
							Begin: time.Date(2022, 10, 7, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 8, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-10-08": {
						{
							Begin: time.Date(2022, 10, 8, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 9, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-10-09": {
						{
							Begin: time.Date(2022, 10, 9, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 10, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-10-10": {
						{
							Begin: time.Date(2022, 10, 10, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 11, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-10-11": {
						{
							Begin: time.Date(2022, 10, 11, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 12, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-10-12": {
						{
							Begin: time.Date(2022, 10, 12, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 12, 12, 0, 0, 0, time.UTC),
						},
					},
				},
			},
			want: want{
				sum: decimal.NewFromFloat(22_365.75),
			},
		},

		{
			name: "Overtid utenfor vaktperioden",
			args: args{
				satser: models.Satser{
					Helg:    decimal.NewFromInt(65),
					Dag:     decimal.NewFromInt(15),
					Natt:    decimal.NewFromInt(25),
					Utvidet: decimal.NewFromInt(25),
				},
				guardPeriod: map[string][]models.Period{
					"2023-05-12": {
						{
							Begin: time.Date(2023, 5, 12, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2023, 5, 13, 0, 0, 0, 0, time.UTC),
						},
					},
					"2023-05-13": {
						{
							Begin: time.Date(2023, 5, 13, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2023, 5, 13, 12, 0, 0, 0, time.UTC),
						},
					},
				},
				timesheet: map[string]models.TimeSheet{
					"2023-05-12": {
						Date:         time.Date(2023, 5, 12, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Virkedag",
						FormName:     "BV 0800-1545 m",
						Salary:       decimal.NewFromInt(500_000),
						Clockings: []models.Clocking{
							{
								In:  time.Date(2023, 5, 12, 8, 0, 0, 0, time.UTC),
								Out: time.Date(2023, 5, 12, 15, 45, 0, 0, time.UTC),
							},
						},
					},
					"2023-05-13": {
						Date:         time.Date(2023, 5, 13, 0, 0, 0, 0, time.UTC),
						WorkingHours: 0,
						WorkingDay:   "Lørdag",
						FormName:     "BV Lørdag IKT",
						Salary:       decimal.NewFromInt(500_000),
						Clockings: []models.Clocking{
							{
								In:  time.Date(2023, 5, 13, 22, 0, 0, 0, time.UTC),
								Out: time.Date(2023, 5, 13, 23, 20, 0, 0, time.UTC),
								OtG: true,
							},
						},
					},
				},
			},
			want: want{
				sum: decimal.NewFromFloat(3_620.87),
				details: &models.Artskoder{
					Morgen: models.Artskode{
						Sum:   decimal.NewFromFloat(798.65),
						Hours: 6,
					},
					Kveld: models.Artskode{
						Sum:   decimal.NewFromFloat(532.43),
						Hours: 4,
					},
					Dag: models.Artskode{
						Sum:   decimal.NewFromFloat(576.49),
						Hours: 6,
					},
					Helg: models.Artskode{
						Sum:   decimal.NewFromFloat(1_693.3),
						Hours: 12,
					},
					Skift: models.Artskode{
						Sum:   decimal.NewFromFloat(20),
						Hours: 4,
					},
				},
			},
		},

		{
			name: "To vaktdager med forskjellig lønn",
			args: args{
				satser: models.Satser{
					Helg:    decimal.NewFromInt(65),
					Dag:     decimal.NewFromInt(15),
					Natt:    decimal.NewFromInt(25),
					Utvidet: decimal.NewFromInt(25),
				},
				guardPeriod: map[string][]models.Period{
					"2022-10-05": {
						{
							Begin: time.Date(2022, 10, 5, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 6, 0, 0, 0, 0, time.UTC),
						},
					},
					"2022-20-06": {
						{
							Begin: time.Date(2022, 10, 6, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2022, 10, 7, 0, 0, 0, 0, time.UTC),
						},
					},
				},
				timesheet: map[string]models.TimeSheet{
					"2022-10-05": {
						Date:         time.Date(2022, 10, 5, 0, 0, 0, 0, time.UTC),
						WorkingHours: 7.75,
						WorkingDay:   "Virkedag",
						FormName:     "BV 0800-1545 m/Beredskapsvakt, start vakt kl 1600 (2018)",
						Salary:       decimal.NewFromInt(725000),
						Koststed:     "000000",
						Formal:       "000000",
						Aktivitet:    "000000",
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
						Salary:       decimal.NewFromInt(750000),
						Koststed:     "000000",
						Formal:       "000000",
						Aktivitet:    "000000",
						Clockings: []models.Clocking{
							{
								In:  time.Date(2022, 10, 6, 7, 13, 24, 0, time.UTC),
								Out: time.Date(2022, 10, 6, 15, 3, 51, 0, time.UTC),
							},
						},
					},
				},
			},
			want: want{
				sum: decimal.NewFromFloat(2_632.98),
				details: &models.Artskoder{
					Morgen: models.Artskode{
						Sum:   decimal.NewFromFloat(1090.54),
						Hours: 6,
					},
					Kveld: models.Artskode{
						Sum:   decimal.NewFromFloat(727.03),
						Hours: 4,
					},
					Dag: models.Artskode{
						Sum:   decimal.NewFromFloat(795.41),
						Hours: 6,
					},
					Skift: models.Artskode{
						Sum:   decimal.NewFromFloat(20),
						Hours: 4,
					},
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
