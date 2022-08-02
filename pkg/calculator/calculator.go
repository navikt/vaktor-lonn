package calculator

import (
	"fmt"
	"github.com/navikt/vaktor-lonn/pkg/dummy"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"math"
	"strings"
	"time"
)

// Range representere en rekke med stigende heltall
// PS: ingen hard sjekk av at den er stigende
type Range struct {
	Begin int
	End   int
}

func (r Range) Count() int {
	return r.End - r.Begin
}

func (r Range) String() string {
	return fmt.Sprintf("(%d...%d)", r.Begin, r.End)
}

// calculateWorkInPeriode returns the number of minutes you have worked in a given period.
// If no work is done, it should return 0.
func calculateWorkInPeriode(work, period Range) int {
	if work.End <= period.Begin || work.Begin >= period.End {
		return 0
	}

	if work.Begin < period.End && work.End > period.Begin {
		if work.Begin >= period.Begin && work.End <= period.End {
			work.Count()
		}

		modified := Range{
			Begin: work.Begin,
			End:   work.End,
		}

		if work.Begin < period.Begin {
			modified.Begin = period.Begin
		}
		if work.End > period.End {
			modified.End = period.End
		}

		return modified.Count()
	}

	return 0
}

// timeToMinutes takes time in the format of 15:04 and converts it to minutes
func timeToMinutes(clock string) (int, error) {
	date, err := time.Parse("15:04", clock)
	if err != nil {
		return -1, err
	}
	return date.Hour()*60 + date.Minute(), nil
}

// timeToRange takes time in the format of 15:04-15:04 and converts it to a range of minutes
func timeToRange(workHours string) (Range, error) {
	hours := strings.Split(workHours, "-")
	begin, err := timeToMinutes(hours[0])
	if err != nil {
		return Range{}, err
	}
	end, err := timeToMinutes(hours[1])
	if err != nil {
		return Range{}, err
	}
	return Range{begin, end}, nil
}

// guardDutyMinutes keeps track of minutes not worked in a given guard duty
type guardDutyMinutes struct {
	Hvilende2006                 int
	Hvilende0620                 int
	Helgetillegg                 int
	Skifttillegg                 int
	WeekendOrHolidayCompensation bool
}

func createRangeForPeriod(day, dutyBegin, dutyEnd, begin, end string) (*Range, error) {
	vaktBegin, err := time.Parse("02.01.200615:04", day+dutyBegin)
	if err != nil {
		return nil, err
	}

	var vaktEnd time.Time
	if dutyEnd == "24:00" {
		vaktEnd = time.Date(vaktBegin.Year(), vaktBegin.Month(), vaktBegin.Day()+1, 0, 0, 0, 0, time.UTC)
	} else {
		vaktEnd, err = time.Parse("02.01.200615:04", day+dutyEnd)
		if err != nil {
			return nil, err
		}
	}
	periodBegin, err := time.Parse("02.01.200615:04", day+begin)
	if err != nil {
		return nil, err
	}

	var periodEnd time.Time
	if end == "24:00" {
		periodEnd = time.Date(periodBegin.Year(), periodBegin.Month(), periodBegin.Day()+1, 0, 0, 0, 0, time.UTC)
	} else {
		periodEnd, err = time.Parse("02.01.200615:04", day+end)
		if err != nil {
			return nil, err
		}
	}

	if vaktBegin == periodBegin || vaktBegin.Before(periodEnd) {
		periodeRange := &Range{
			Begin: periodBegin.Hour()*60 + periodBegin.Minute(),
			End:   periodEnd.Hour()*60 + periodEnd.Minute(),
		}

		if end == "24:00" {
			periodeRange.End = 24 * 60
		}

		// sjekk om vakt starter senere enn "normalen"
		if vaktBegin.After(periodBegin) {
			periodeRange.Begin = vaktBegin.Hour()*60 + vaktBegin.Minute()
		}
		// sjekk om vakt slutter før "normalen"
		if vaktEnd.Before(periodEnd) {
			periodeRange.End = vaktEnd.Hour()*60 + vaktEnd.Minute()
		}

		// personen har vakt i denne perioden!
		return periodeRange, nil
	}

	return nil, nil
}

// ParsePeriode returns an object with the minutes you have been having guard duty each day in a given periode
func ParsePeriode(periods map[string]models.Period, timesheet map[string][]string) (map[string]guardDutyMinutes, error) {
	guardHours := map[string]guardDutyMinutes{}

	for day, period := range periods {
		date, err := time.Parse("02.01.2006", day)
		if err != nil {
			return guardHours, err
		}

		dutyHours := guardDutyMinutes{}
		// TODO: Ta høyde for sommer- og vintertid

		// sjekk om man har vakt i perioden 00-06
		hvilende0006Range, _ := createRangeForPeriod(day, period.Fra, period.Til, "00:00", "06:00")
		if hvilende0006Range != nil {
			// Personen har faktisk vakt!
			// hvor mye jobbet personen i denne tidsperioden?
			minutesWorked := 0
			for _, workHours := range timesheet[day] {
				workRange, err := timeToRange(workHours)
				if err != nil {
					return guardHours, err
				}
				minutesWorked += calculateWorkInPeriode(workRange, *hvilende0006Range)
			}
			dutyHours.Hvilende2006 += hvilende0006Range.Count() - minutesWorked
		}

		// sjekk om man har vakt i perioden 20-24
		hvilende2000Range, _ := createRangeForPeriod(day, period.Fra, period.Til, "20:00", "24:00")
		if hvilende2000Range != nil {
			// Personen har faktisk vakt!
			// hvor mye jobbet personen i denne tidsperioden?
			minutesWorked := 0
			for _, workHours := range timesheet[day] {
				workRange, err := timeToRange(workHours)
				if err != nil {
					return guardHours, err
				}

				minutesWorked += calculateWorkInPeriode(workRange, *hvilende2000Range)
			}
			dutyHours.Hvilende2006 += hvilende2000Range.Count() - minutesWorked
		}

		// sjekk om man har vakt i perioden 06-20
		hvilende0620Range, _ := createRangeForPeriod(day, period.Fra, period.Til, "06:00", "20:00")
		if hvilende0620Range != nil {
			// Personen har faktisk vakt!
			// hvor mye jobbet personen i denne tidsperioden?
			minutesWorked := 0
			for _, workHours := range timesheet[day] {
				workRange, err := timeToRange(workHours)
				if err != nil {
					return guardHours, err
				}

				minutesWorked += calculateWorkInPeriode(workRange, *hvilende0620Range)
			}
			dutyHours.Hvilende0620 += hvilende0620Range.Count() - minutesWorked
		}

		if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday || period.Helligdag {
			// sjekk om man har vakt i perioden 00-24
			helgetillegg0024Range, _ := createRangeForPeriod(day, period.Fra, period.Til, "00:00", "24:00")
			if helgetillegg0024Range != nil {
				// Personen har faktisk vakt!
				// hvor mye jobbet personen i denne tidsperioden?
				minutesWorked := 0
				for _, workHours := range timesheet[day] {
					workRange, err := timeToRange(workHours)
					if err != nil {
						return guardHours, err
					}

					minutesWorked += calculateWorkInPeriode(workRange, *helgetillegg0024Range)
				}
				dutyHours.Helgetillegg += helgetillegg0024Range.Count() - minutesWorked
				dutyHours.WeekendOrHolidayCompensation = true
			}
		} else {
			// sjekk om man har vakt i perioden 06-07
			skifttillegg0607Range, _ := createRangeForPeriod(day, period.Fra, period.Til, "06:00", "07:00")
			if skifttillegg0607Range != nil {
				// Personen har faktisk vakt!
				// hvor mye jobbet personen i denne tidsperioden?
				minutesWorked := 0
				for _, workHours := range timesheet[day] {
					workRange, err := timeToRange(workHours)
					if err != nil {
						return guardHours, err
					}

					minutesWorked += calculateWorkInPeriode(workRange, *skifttillegg0607Range)
				}
				dutyHours.Skifttillegg += skifttillegg0607Range.Count() - minutesWorked
			}

			// sjekk om man har vakt i perioden 17-20
			skifttillegg1720Range, _ := createRangeForPeriod(day, period.Fra, period.Til, "17:00", "20:00")
			if skifttillegg1720Range != nil {
				// Personen har faktisk vakt!
				// hvor mye jobbet personen i denne tidsperioden?
				minutesWorked := 0
				for _, workHours := range timesheet[day] {
					workRange, err := timeToRange(workHours)
					if err != nil {
						return guardHours, err
					}

					minutesWorked += calculateWorkInPeriode(workRange, *skifttillegg1720Range)
				}
				dutyHours.Skifttillegg += skifttillegg1720Range.Count() - minutesWorked
			}
		}

		guardHours[day] = dutyHours
	}

	return guardHours, nil
}

func CalculateCompensation(minutes map[string]guardDutyMinutes) (float64, error) {
	var compensation struct {
		Night   float64
		Day     float64
		Weekend float64
		Utvidet float64
	}

	for _, duty := range minutes {
		compensation.Day += float64(duty.Hvilende0620)
		compensation.Night += float64(duty.Hvilende2006)
		compensation.Utvidet += float64(duty.Skifttillegg)
		compensation.Weekend += float64(duty.Helgetillegg)
	}

	// TODO: runne av til nærmeste 2 desimaler
	return math.Round(compensation.Day/60)*10.0 +
		math.Round(compensation.Night/60)*20.0 +
		(math.Round(compensation.Weekend/60) * 55.0 / 5) +
		(math.Round(compensation.Utvidet/60) * 15.0 / 5), nil
}

func CalculateOvertime(minutes map[string]guardDutyMinutes, salary float64) (float64, error) {
	overtimeWeekendMinutes := 0.0
	overtimeWorkDayMinutes := 0.0
	overtimeWorkNightMinutes := 0.0

	for _, duty := range minutes {
		if duty.WeekendOrHolidayCompensation {
			overtimeWeekendMinutes += float64(duty.Hvilende0620 + duty.Hvilende2006)
		} else {
			overtimeWorkDayMinutes += float64(duty.Hvilende0620)
			overtimeWorkNightMinutes += float64(duty.Hvilende2006)
		}
	}

	ots50 := math.Round((salary/1850*1.5)*100) / 100
	ots100 := math.Round((salary/1850*2.0)*100) / 100

	overtimeWork := (math.Round(overtimeWorkDayMinutes/60.0)*ots50 + math.Round(overtimeWorkNightMinutes/60.0)*ots100) / 5.0
	overtimeWeekend := math.Round(overtimeWeekendMinutes/60.0) * ots100 / 5.0

	return overtimeWeekend + overtimeWork, nil
}

func CalculateEarnings(minutes map[string]guardDutyMinutes, salary int) (float64, error) {
	compensation, err := CalculateCompensation(minutes)
	if err != nil {
		return -1, err
	}
	overtime, err := CalculateOvertime(minutes, float64(salary))
	if err != nil {
		return -1, err
	}

	return compensation + overtime, nil
}

func GuarddutySalary(plan models.Plan) (float64, error) {
	minWinTid := dummy.GetMinWinTid(plan.Ident)
	salary := dummy.GetSalary(plan.Ident)

	minutes, err := ParsePeriode(plan.Periods, minWinTid)
	if err != nil {
		return -1, err
	}

	return CalculateEarnings(minutes, salary)
}
