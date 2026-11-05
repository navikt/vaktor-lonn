package testhelpers

import (
	"time"

	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/shopspring/decimal"
)

// At parses "2006-01-02T15:04" or "2006-01-02T15:04:05" into time.Time (UTC).
// Panics on parse errors.
func At(s string) time.Time {
	if t, err := time.Parse("2006-01-02T15:04:05", s); err == nil {
		return t
	}
	t, err := time.Parse("2006-01-02T15:04", s)
	if err != nil {
		panic(err)
	}
	return t
}

// D parses "2006-01-02" into time.Time at midnight (UTC).
// Panics on parse errors.
func D(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

// Dec creates a decimal.Decimal from a float64.
func Dec(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}

// DecI creates a decimal.Decimal from an int64.
func DecI(i int64) decimal.Decimal {
	return decimal.NewFromInt(i)
}

// DefaultSatser is the most commonly used Satser in tests:
// Helg=65, Dag=15, Natt=25, Utvidet=25.
var DefaultSatser = models.Satser{
	Helg:    DecI(65),
	Dag:     DecI(15),
	Natt:    DecI(25),
	Utvidet: DecI(25),
}

// Common salary values used across many tests.
var (
	Salary500k = DecI(500_000)
	Salary725k = DecI(725_000)
	Salary750k = DecI(750_000)
	Salary800k = DecI(800_000)
)

// Period creates a models.Period from two "2006-01-02T15:04" strings.
func Period(begin, end string) models.Period {
	return models.Period{Begin: At(begin), End: At(end)}
}

// Clocking creates a models.Clocking from two "2006-01-02T15:04" strings.
func Clocking(in, out string) models.Clocking {
	return models.Clocking{In: At(in), Out: At(out)}
}

// ClockingOtG creates a models.Clocking with OtG=true from two "2006-01-02T15:04" strings.
func ClockingOtG(in, out string) models.Clocking {
	return models.Clocking{In: At(in), Out: At(out), OtG: true}
}

// WorkDay creates a models.TimeSheet for a normal workday.
func WorkDay(date string, hours float64, dayType string, formName string, salary decimal.Decimal, clockings ...models.Clocking) models.TimeSheet {
	return models.TimeSheet{
		Date:         D(date),
		WorkingHours: hours,
		WorkingDay:   dayType,
		FormName:     formName,
		Salary:       salary,
		Clockings:    clockings,
	}
}

// OffDay creates a models.TimeSheet for a weekend or holiday with no clockings.
func OffDay(date string, dayType string, formName string, salary decimal.Decimal) models.TimeSheet {
	return models.TimeSheet{
		Date:       D(date),
		WorkingDay: dayType,
		FormName:   formName,
		Salary:     salary,
		Clockings:  []models.Clocking{},
	}
}

// MWTDag creates a models.MWTDag with the given parameters.
func MWTDag(dato string, skjemaTid float64, skjemaNavn string, godkjent int, virkedag string, salary int, stemplinger ...models.MWTStempling) models.MWTDag {
	return models.MWTDag{
		Dato:        dato + "T00:00:00",
		SkjemaTid:   skjemaTid,
		SkjemaNavn:  skjemaNavn,
		Godkjent:    godkjent,
		Virkedag:    virkedag,
		Stemplinger: stemplinger,
		Stillinger: []models.MWTStilling{
			{RATEK001: salary},
		},
	}
}

// MWTStempling creates a models.MWTStempling.
func MWTStempling(stemplingTid, retning, typ string, fravarkode int) models.MWTStempling {
	return models.MWTStempling{
		StemplingTid: stemplingTid,
		Retning:      retning,
		Type:         typ,
		Fravarkode:   fravarkode,
	}
}

// MWTStemplingBV creates a models.MWTStempling with an overtime justification.
func MWTStemplingBV(stemplingTid, retning, typ string, fravarkode int, begrunnelse string) models.MWTStempling {
	return models.MWTStempling{
		StemplingTid:       stemplingTid,
		Retning:            retning,
		Type:               typ,
		Fravarkode:         fravarkode,
		OvertidBegrunnelse: begrunnelse,
	}
}
