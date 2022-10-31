package ranges

import (
	"fmt"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"time"
)

// Range representere en rekke med stigende heltall
// PS: ingen hard sjekk av at den er stigende
type Range struct {
	Begin int
	End   int
}

func (r Range) Count() float64 {
	return float64(r.End - r.Begin)
}

func (r Range) String() string {
	return fmt.Sprintf("(%d...%d)", r.Begin, r.End)
}

// CalculateMinutesOverlapping returns the number of minutes that are overlapping two ranges.
// Returns 0 if the ranges does not overlap.
func CalculateMinutesOverlapping(a, b Range) float64 {
	if a.End <= b.Begin || a.Begin >= b.End {
		return 0
	}

	if a.Begin < b.End && a.End > b.Begin {
		if a.Begin >= b.Begin && a.End <= b.End {
			a.Count()
		}

		modified := Range{
			Begin: a.Begin,
			End:   a.End,
		}

		if a.Begin < b.Begin {
			modified.Begin = b.Begin
		}
		if a.End > b.End {
			modified.End = b.End
		}

		return modified.Count()
	}

	return 0
}

// timeToMinutes takes time in the format of 15:04 and converts it to minutes
func timeToMinutes(hours, minutes int) int {
	return hours*60 + minutes
}

// FromTime takes time in the format of 15:04-15:04 and converts it to a range of minutes
func FromTime(in, out time.Time) Range {
	outHour := out.Hour()
	if in.YearDay() < out.YearDay() {
		outHour = 24
	}
	return Range{
		timeToMinutes(in.Hour(), in.Minute()),
		timeToMinutes(outHour, out.Minute()),
	}
}

// CreateForPeriod creates a range of minutes based on two dates. This will fit the threshold used.
// Returns nil if the period is outside the threshold.
func CreateForPeriod(period, threshold models.Period) *Range {
	if period.Begin.After(threshold.End) ||
		period.Begin.Equal(threshold.End) ||
		period.End.Before(threshold.Begin) ||
		period.End.Equal(threshold.Begin) {
		return nil
	}

	periodeRange := &Range{
		Begin: threshold.Begin.Hour()*60 + threshold.Begin.Minute(),
		End:   threshold.End.Hour()*60 + threshold.End.Minute(),
	}
	if threshold.End.YearDay() > threshold.Begin.YearDay() {
		periodeRange.End = 24*60 + threshold.End.Minute()
	}

	// sjekk om vakt starter senere enn "normalen"
	if period.Begin.After(threshold.Begin) {
		periodeRange.Begin = period.Begin.Hour()*60 + period.Begin.Minute()
	}
	// sjekk om vakt slutter før "normalen"
	if period.End.Before(threshold.End) {
		if period.End.YearDay() > period.Begin.YearDay() {
			periodeRange.End = 24*60 + period.End.Minute()
		} else {
			periodeRange.End = period.End.Hour()*60 + period.End.Minute()
		}
	}

	// personen har vakt i denne perioden!
	return periodeRange
}
