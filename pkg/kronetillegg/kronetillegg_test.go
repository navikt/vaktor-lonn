package kronetillegg

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/navikt/vaktor-lonn/pkg/testhelpers"
)

func TestCalculate(t *testing.T) {
	type args struct {
		minutes map[string]models.GuardDuty
		satser  models.Satser
		payroll *models.Payroll
	}
	tests := []struct {
		name string
		args args
		want models.Artskoder
	}{
		{
			name: "Beredskapsvakt en uke",
			args: args{
				payroll: &models.Payroll{},
				satser:  testhelpers.DefaultSatser,
				minutes: map[string]models.GuardDuty{
					"2022-10-12": {Hvilende2000: 240, Hvilende0006: 0, Hvilende0620: 255, Skifttillegg: 180},
					"2022-10-13": {Hvilende2000: 240, Hvilende0006: 360, Hvilende0620: 375, Skifttillegg: 240},
					"2022-10-14": {Hvilende2000: 240, Hvilende0006: 360, Hvilende0620: 375, Skifttillegg: 240},
					"2022-10-15": {Hvilende2000: 240, Hvilende0006: 360, Hvilende0620: 840, Helgetillegg: 1440, IsWeekend: true},
					"2022-10-16": {Hvilende2000: 240, Hvilende0006: 360, Hvilende0620: 840, Helgetillegg: 1440, IsWeekend: true},
					"2022-10-17": {Hvilende2000: 240, Hvilende0006: 360, Hvilende0620: 375, Skifttillegg: 240},
					"2022-10-18": {Hvilende2000: 240, Hvilende0006: 360, Hvilende0620: 375, Skifttillegg: 240},
					"2022-10-19": {Hvilende2000: 0, Hvilende0006: 360, Hvilende0620: 120, Skifttillegg: 60},
				},
			},
			want: models.Artskoder{
				Morgen: models.Artskode{Sum: testhelpers.DecI(750)},
				Dag:    models.Artskode{Sum: testhelpers.DecI(465)},
				Kveld:  models.Artskode{Sum: testhelpers.DecI(500)},
				Helg:   models.Artskode{Sum: testhelpers.DecI(1544)},
				Skift:  models.Artskode{Sum: testhelpers.DecI(100), Hours: 20},
			},
		},

		{
			name: "Utrykning i helg (søndag)",
			args: args{
				payroll: &models.Payroll{},
				satser:  testhelpers.DefaultSatser,
				minutes: map[string]models.GuardDuty{
					"2022-10-15": {Hvilende2000: 240, Hvilende0006: 360, Hvilende0620: 840, Helgetillegg: 1440, IsWeekend: true},
					"2022-10-16": {Hvilende2000: 120, Hvilende0006: 360, Hvilende0620: 840, Helgetillegg: 1320, IsWeekend: true},
				},
			},
			want: models.Artskoder{
				Helg: models.Artskode{Sum: testhelpers.DecI(1468)},
			},
		},

		{
			name: "Beredskapsvakt en bevegelig hellig dag",
			args: args{
				payroll: &models.Payroll{},
				satser:  testhelpers.DefaultSatser,
				minutes: map[string]models.GuardDuty{
					"2022-10-12": {
						Hvilende2000:  240,
						Hvilende0006:  360,
						Helligdag0620: 840,
						Hvilende0620:  0,
						Helgetillegg:  0,
						Skifttillegg:  240,
					},
				},
			},
			want: models.Artskoder{
				Morgen: models.Artskode{Sum: testhelpers.DecI(150)},
				Kveld:  models.Artskode{Sum: testhelpers.DecI(100)},
				Dag:    models.Artskode{Sum: testhelpers.DecI(210)},
				Skift:  models.Artskode{Sum: testhelpers.DecI(20), Hours: 4},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Calculate(tt.args.minutes, tt.args.satser, tt.args.payroll)

			if diff := cmp.Diff(tt.want, tt.args.payroll.Artskoder); diff != "" {
				t.Errorf("Calculate() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
