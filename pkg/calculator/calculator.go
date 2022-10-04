package calculator

import (
	"fmt"
	"github.com/navikt/vaktor-lonn/pkg/compensation"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/navikt/vaktor-lonn/pkg/overtime"
	"time"

	"github.com/vjeantet/eastertime"
)

const (
	VaktorDateFormat = "2006-01-02" // TODO: Bytt ut alle date-formater til denne
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

// calculateMinutesOverlappingInPeriods returns the number of minutes that are overlapping two ranges.
// Returns 0 if the ranges does not overlap.
func calculateMinutesOverlappingInPeriods(a, b Range) int {
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
func timeToMinutes(clock time.Time) int {
	return clock.Hour()*60 + clock.Minute()
}

// timeToRange takes time in the format of 15:04-15:04 and converts it to a range of minutes
func timeToRange(workHours models.Clocking) Range {
	return Range{timeToMinutes(workHours.In), timeToMinutes(workHours.Out)}
}

// createRangeForPeriod creates a range of minutes based on two dates. This will fit the threshold used.
// Returns nil if the period is outside the threshold.
func createRangeForPeriod(period, threshold models.Period) (*Range, error) {
	if period.Begin.After(threshold.End) ||
		period.Begin.Equal(threshold.End) ||
		period.End.Before(threshold.Begin) ||
		period.End.Equal(threshold.Begin) {
		return nil, nil
	}

	periodeRange := &Range{
		Begin: threshold.Begin.Hour()*60 + threshold.Begin.Minute(),
		End:   threshold.End.Hour()*60 + threshold.End.Minute(),
	}
	if threshold.End.Day() > threshold.Begin.Day() {
		periodeRange.End = 24*60 + threshold.End.Minute()
	}

	// sjekk om vakt starter senere enn "normalen"
	if period.Begin.After(threshold.Begin) {
		periodeRange.Begin = period.Begin.Hour()*60 + period.Begin.Minute()
	}
	// sjekk om vakt slutter før "normalen"
	if period.End.Before(threshold.End) {
		if period.End.Day() > period.Begin.Day() {
			periodeRange.End = 24*60 + period.End.Minute()
		} else {
			periodeRange.End = period.End.Hour()*60 + period.End.Minute()
		}
	}

	// personen har vakt i denne perioden!
	return periodeRange, nil
}

// calculateMinutesToBeCompensated returns an object with the minutes you have been having guard duty each day in a given periode
func calculateMinutesToBeCompensated(schedule map[string][]models.Period, timesheet map[string]models.TimeSheet) (map[string]models.GuardDuty, error) {
	guardHours := map[string]models.GuardDuty{}

	for day, periods := range schedule {
		dutyHours := models.GuardDuty{}

		modifier, err := calculateDaylightSavingTimeModifier(day)
		if err != nil {
			return nil, err
		}
		dutyHours.Hvilende2006 += modifier

		for _, period := range periods {
			currentDay := timesheet[day]
			date := currentDay.Date

			// sjekk om man har vakt i perioden 00-06
			minutesWorked, err := calculateMinutesWithGuardDutyInPeriod(period, models.Period{
				Begin: time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC),
				End:   time.Date(date.Year(), date.Month(), date.Day(), 6, 0, 0, 0, time.UTC),
			}, currentDay.Clockings)
			if err != nil {
				return nil, err
			}
			dutyHours.Hvilende2006 += minutesWorked

			// sjekk om man har vakt i perioden 20-24
			minutesWorked, err = calculateMinutesWithGuardDutyInPeriod(period, models.Period{
				Begin: time.Date(date.Year(), date.Month(), date.Day(), 20, 0, 0, 0, time.UTC),
				End:   time.Date(date.Year(), date.Month(), date.Day()+1, 0, 0, 0, 0, time.UTC),
			}, currentDay.Clockings)
			if err != nil {
				return nil, err
			}
			dutyHours.Hvilende2006 += minutesWorked

			// sjekk om man har vakt i perioden 06-20
			minutesWorked, err = calculateMinutesWithGuardDutyInPeriod(period, models.Period{
				Begin: time.Date(date.Year(), date.Month(), date.Day(), 6, 0, 0, 0, time.UTC),
				End:   time.Date(date.Year(), date.Month(), date.Day(), 20, 0, 0, 0, time.UTC),
			}, currentDay.Clockings)
			if err != nil {
				return nil, err
			}
			dutyHours.Hvilende0620 += minutesWorked

			validateHowMuchDutyHours, err := validateHowMuchDutyHours(date, false) // TODO Helligdag
			if err != nil {
				return nil, err
			}
			if validateHowMuchDutyHours {
				// TODO: En vaktperiode kan ikke være lengre enn 17t i døgnet mandag-fredag under sommertid, og 16t15m mandag-fredag under vintertid
				// Sjekk om en person har Hvilende0620 mer enn 8,5t eller Hvilende0620+Hvilende2006 mer enn 17t/16t15min.
				// Det er unntak følgende dager: onsdag før skjærtorsdag, julaften, romjulen, nyttårsaften

				// Dette er tiden du ikke jobbet i kjernetiden. Da vil man ikke kunne få vakttillegg, da andre er på jobb til å ta uforutsette hendelser.
				// TODO: Bruk arbeidstid/kjernetid fra MinWinTid
				kjerneTid := models.Period{
					Begin: time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, time.UTC),
					End:   time.Date(date.Year(), date.Month(), date.Day(), 14, 30, 0, 0, time.UTC),
				}
				minutesNotWorkedInCoreWorkingHours := 0
				minutesNotWorkedInCoreWorkingHours, err = calculateMinutesWithGuardDutyInPeriod(period, kjerneTid, currentDay.Clockings)
				dutyHours.Hvilende0620 -= minutesNotWorkedInCoreWorkingHours

				// TODO: Sjekk om det er sommertid eller vintertid for NAV, og at personen som jobber følger det
				NAVSummerTimeBegin := time.Date(date.Year(), time.May, 15, 0, 0, 0, 0, time.UTC)
				NAVSummerTimeEnd := time.Date(date.Year(), time.September, 15, 0, 0, 0, 0, time.UTC)
				maxDutyMinutes := 16*60 + 15
				if date.After(NAVSummerTimeBegin) && date.Before(NAVSummerTimeEnd) {
					maxDutyMinutes = 17 * 60
				}
				addedDutyMinutes := dutyHours.Hvilende0620 + dutyHours.Hvilende2006
				if addedDutyMinutes > maxDutyMinutes {
					// Personen har fått registert for mye vakt den dagen. Fjern diff-en
					dutyHours.Hvilende0620 -= maxDutyMinutes - addedDutyMinutes
					if dutyHours.Hvilende0620 < 0 {
						dutyHours.Hvilende2006 += dutyHours.Hvilende0620
						dutyHours.Hvilende0620 = 0
					}
				}
			}

			if currentDay.WeekendCompensation {
				// sjekk om man har vakt i perioden 00-24
				minutesWorked, err = calculateMinutesWithGuardDutyInPeriod(period, models.Period{
					Begin: time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC),
					End:   time.Date(date.Year(), date.Month(), date.Day()+1, 0, 0, 0, 0, time.UTC)}, currentDay.Clockings)
				if err != nil {
					return nil, err
				}
				dutyHours.Helgetillegg += minutesWorked
				dutyHours.WeekendOrHolidayCompensation = true
			} else {
				// sjekk om man har vakt i perioden 06-07
				minutesWorked, err = calculateMinutesWithGuardDutyInPeriod(period, models.Period{
					Begin: time.Date(date.Year(), date.Month(), date.Day(), 6, 0, 0, 0, time.UTC),
					End:   time.Date(date.Year(), date.Month(), date.Day(), 7, 0, 0, 0, time.UTC)}, currentDay.Clockings)
				if err != nil {
					return nil, err
				}
				dutyHours.Skifttillegg += minutesWorked

				// sjekk om man har vakt i perioden 17-20
				minutesWorked, err = calculateMinutesWithGuardDutyInPeriod(period, models.Period{
					Begin: time.Date(date.Year(), date.Month(), date.Day(), 17, 0, 0, 0, time.UTC),
					End:   time.Date(date.Year(), date.Month(), date.Day(), 20, 0, 0, 0, time.UTC)}, currentDay.Clockings)
				if err != nil {
					return nil, err
				}
				dutyHours.Skifttillegg += minutesWorked
			}
		}
		guardHours[day] = dutyHours
	}

	return guardHours, nil
}

func validateHowMuchDutyHours(date time.Time, helligdag bool) (bool, error) {
	// Det er unntak følgende dager: onsdag før skjærtorsdag, julaften, romjulen, nyttårsaften
	christmasEve := time.Date(date.Year(), time.December, 24, 0, 0, 0, 0, time.UTC)
	if date.YearDay() == christmasEve.YearDay() {
		return false, nil
	}
	newYearsEve := time.Date(date.Year(), time.December, 31, 0, 0, 0, 0, time.UTC)
	if date.YearDay() == newYearsEve.YearDay() {
		return false, nil
	}
	if date.After(christmasEve) && date.Before(newYearsEve) {
		return false, nil
	}
	easterEve, err := eastertime.CatholicByYear(date.Year())
	if err != nil {
		return false, err
	}
	if date.YearDay() == easterEve.YearDay()-3 {
		return false, nil
	}
	if date.Weekday() == time.Saturday {
		return false, nil
	}
	if date.Weekday() == time.Sunday {
		return false, err
	}
	if helligdag {
		return false, err
	}
	return true, nil
}

// calculateDaylightSavingTimeModifier returns either -60 or 60 minutes if $day is when the clock is advanced
func calculateDaylightSavingTimeModifier(day string) (int, error) {
	date, err := time.Parse(VaktorDateFormat, day)
	if err != nil {
		return 0, err
	}

	summerTime := time.Date(date.Year(), time.March, 31, 0, 0, 0, 0, time.UTC)
	summerTime = summerTime.AddDate(0, 0, -int(summerTime.Weekday()))
	if summerTime.Year() == date.Year() && summerTime.YearDay() == date.YearDay() {
		fmt.Println("It's summertime madness!")
		return -60, nil
	}

	winterTime := time.Date(date.Year(), time.March, 31, 0, 0, 0, 0, time.UTC)
	winterTime = winterTime.AddDate(0, 0, -int(winterTime.Weekday()))
	if winterTime.Year() == date.Year() && winterTime.YearDay() == date.YearDay() {
		fmt.Println("It's wintertime sadness!")
		return 60, nil
	}

	return 0, nil
}

// calculateMinutesWithGuardDutyInPeriod return the number of minutes that you have non-working guard duty
func calculateMinutesWithGuardDutyInPeriod(vaktPeriod models.Period, compPeriod models.Period, timesheet []models.Clocking) (int, error) {
	dutyRange, err := createRangeForPeriod(vaktPeriod, compPeriod)
	if err != nil {
		return 0, err
	}

	minutesWorked := 0

	if dutyRange != nil {
		for _, workHours := range timesheet {
			workRange := timeToRange(workHours)
			if err != nil {
				return 0, err
			}

			minutesWorked += calculateMinutesOverlappingInPeriods(workRange, *dutyRange)
		}

		return dutyRange.Count() - minutesWorked, nil
	}

	return minutesWorked, nil
}

func GuarddutySalary(plan models.Vaktplan, minWinTid models.MinWinTid) error {
	minutes, err := calculateMinutesToBeCompensated(plan.Schedule, minWinTid.Timesheet)
	if err != nil {
		return err
	}

	compensationTotal := compensation.Calculate(minutes, minWinTid.Satser)
	overtimeTotal := overtime.Calculate(minutes, minWinTid.Salary)

	// TODO: Må returnere penger, og hvor mye per tillegg!
	fmt.Printf("Money earned %v + %v", compensationTotal, overtimeTotal)
	//report.Earnings.Compensation.Total = compensationTotal
	//report.Earnings.Overtime.Total = overtimeTotal
	//report.Earnings.Total = compensationTotal.Add(overtimeTotal)

	return nil
}
