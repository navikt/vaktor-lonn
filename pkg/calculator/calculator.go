package calculator

import (
	"fmt"
	"github.com/navikt/vaktor-lonn/pkg/overtime"
	"time"

	"github.com/navikt/vaktor-lonn/pkg/compensation"
	"github.com/navikt/vaktor-lonn/pkg/models"
)

const (
	VaktorDateFormat = "2006-01-02"
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

// calculateMinutesOverlappingInPeriods returns the number of minutes that are overlapping two ranges.
// Returns 0 if the ranges does not overlap.
func calculateMinutesOverlappingInPeriods(a, b Range) float64 {
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
func createRangeForPeriod(period, threshold models.Period) *Range {
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
	return periodeRange
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
		dutyHours.Hvilende0006 += modifier

		for _, period := range periods {
			currentDay := timesheet[day]
			date := currentDay.Date

			// sjekk om man har vakt i perioden 00-06
			minutesWithGuardDuty := calculateMinutesWithGuardDutyInPeriod(period, models.Period{
				Begin: time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC),
				End:   time.Date(date.Year(), date.Month(), date.Day(), 6, 0, 0, 0, time.UTC),
			}, currentDay.Clockings)
			dutyHours.Hvilende0006 += minutesWithGuardDuty

			// sjekk om man har vakt i perioden 20-24
			minutesWithGuardDuty = calculateMinutesWithGuardDutyInPeriod(period, models.Period{
				Begin: time.Date(date.Year(), date.Month(), date.Day(), 20, 0, 0, 0, time.UTC),
				End:   time.Date(date.Year(), date.Month(), date.Day()+1, 0, 0, 0, 0, time.UTC),
			}, currentDay.Clockings)
			dutyHours.Hvilende2000 += minutesWithGuardDuty

			// sjekk om man har vakt i perioden 06-20
			minutesWithGuardDuty = calculateMinutesWithGuardDutyInPeriod(period, models.Period{
				Begin: time.Date(date.Year(), date.Month(), date.Day(), 6, 0, 0, 0, time.UTC),
				End:   time.Date(date.Year(), date.Month(), date.Day(), 20, 0, 0, 0, time.UTC),
			}, currentDay.Clockings)
			dutyHours.Hvilende0620 += minutesWithGuardDuty

			kjernetidModifier := calculateGuardDutyInKjernetid(currentDay, date, period)
			maxGuardDutyModifier := calculateMaxGuardDutyTime(currentDay, dutyHours.Hvilende0620+dutyHours.Hvilende2000+dutyHours.Hvilende0006)
			dutyHours.Hvilende0620 += kjernetidModifier - maxGuardDutyModifier

			if isWeekend(currentDay.WorkingDay) {
				// sjekk om man har vakt i perioden 00-24
				minutesWithGuardDuty = calculateMinutesWithGuardDutyInPeriod(period, models.Period{
					Begin: time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC),
					End:   time.Date(date.Year(), date.Month(), date.Day()+1, 0, 0, 0, 0, time.UTC)}, currentDay.Clockings)
				dutyHours.Helgetillegg += minutesWithGuardDuty
			} else {
				// sjekk om man har vakt i perioden 06-07
				minutesWithGuardDuty = calculateMinutesWithGuardDutyInPeriod(period, models.Period{
					Begin: time.Date(date.Year(), date.Month(), date.Day(), 6, 0, 0, 0, time.UTC),
					End:   time.Date(date.Year(), date.Month(), date.Day(), 7, 0, 0, 0, time.UTC)}, currentDay.Clockings)
				dutyHours.Skifttillegg += minutesWithGuardDuty

				// sjekk om man har vakt i perioden 17-20
				minutesWithGuardDuty = calculateMinutesWithGuardDutyInPeriod(period, models.Period{
					Begin: time.Date(date.Year(), date.Month(), date.Day(), 17, 0, 0, 0, time.UTC),
					End:   time.Date(date.Year(), date.Month(), date.Day(), 20, 0, 0, 0, time.UTC)}, currentDay.Clockings)
				dutyHours.Skifttillegg += minutesWithGuardDuty
			}

			dutyHours.WeekendOrHolidayCompensation = isWeekendOrHoliday(currentDay.WorkingDay)
		}
		guardHours[day] = dutyHours
	}

	return guardHours, nil
}

// calculateMaxGuardDutyTime fjerner minutter som overstiger lovlig antall tid med vakt man kan gå per dag.
func calculateMaxGuardDutyTime(currentDay models.TimeSheet, totalGuardDutyInADayInMinutes float64) float64 {
	if isWeekendOrHoliday(currentDay.WorkingDay) {
		return 0
	}

	maxGuardDutyInMinutes := 24*60 - currentDay.WorkingHours*60
	if totalGuardDutyInADayInMinutes > maxGuardDutyInMinutes {
		return maxGuardDutyInMinutes - totalGuardDutyInADayInMinutes
	}

	return 0
}

func isWeekendOrHoliday(day string) bool {
	return day != "Virkedag"
}

func isWeekend(day string) bool {
	return day == "Lørdag" || day == "Søndag"
}

// calculateGuardDutyInKjernetid sjekker om man hadde vakt i kjernetiden. Man vil ikke kunne få vakttillegg i
// kjernetiden, da andre skal være på jobb til å ta seg av uforutsette hendelser.
func calculateGuardDutyInKjernetid(currentDay models.TimeSheet, date time.Time, period models.Period) float64 {
	if isWeekendOrHoliday(currentDay.WorkingDay) {
		return 0
	}

	kjernetid := createKjernetid(date, currentDay.FormName)
	return calculateMinutesWithGuardDutyInPeriod(period, kjernetid, currentDay.Clockings)
}

// createKjernetid returns the current day kjernetid. Except for three days, it's always from 09 til 1430
func createKjernetid(date time.Time, formName string) models.Period {
	startOfKjernetid := time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, time.UTC)
	endOfKjernetid := time.Date(date.Year(), date.Month(), date.Day(), 14, 30, 0, 0, time.UTC)
	if formName == "Julaften 0800-1200 *" || formName == "Onsdag før Påske 0800-1200 *" {
		startOfKjernetid = time.Date(date.Year(), date.Month(), date.Day(), 8, 0, 0, 0, time.UTC)
		endOfKjernetid = time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, time.UTC)
	} else if formName == "Nyttårsaften 1000-1200 *" {
		startOfKjernetid = time.Date(date.Year(), date.Month(), date.Day(), 10, 0, 0, 0, time.UTC)
		endOfKjernetid = time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, time.UTC)
	}

	return models.Period{
		Begin: startOfKjernetid,
		End:   endOfKjernetid,
	}
}

// TODO: Lag en egen test av at vi stiller klokken
// calculateDaylightSavingTimeModifier returns either -60 or 60 minutes if $day is when the clock is advanced
func calculateDaylightSavingTimeModifier(day string) (float64, error) {
	date, err := time.Parse(VaktorDateFormat, day)
	if err != nil {
		return 0, err
	}

	summerTime := time.Date(date.Year(), time.March, 31, 0, 0, 0, 0, time.UTC)
	summerTime = summerTime.AddDate(0, 0, -int(summerTime.Weekday()))
	if summerTime.Year() == date.Year() && summerTime.YearDay() == date.YearDay() {
		return -60, nil
	}

	winterTime := time.Date(date.Year(), time.March, 31, 0, 0, 0, 0, time.UTC)
	winterTime = winterTime.AddDate(0, 0, -int(winterTime.Weekday()))
	if winterTime.Year() == date.Year() && winterTime.YearDay() == date.YearDay() {
		return 60, nil
	}

	return 0, nil
}

// calculateMinutesWithGuardDutyInPeriod return the number of minutes that you have non-working guard duty
func calculateMinutesWithGuardDutyInPeriod(vaktPeriod models.Period, compPeriod models.Period, timesheet []models.Clocking) float64 {
	dutyRange := createRangeForPeriod(vaktPeriod, compPeriod)
	minutesWithGuardDuty := 0.0

	if dutyRange != nil {
		for _, workHours := range timesheet {
			workRange := timeToRange(workHours)
			minutesWithGuardDuty += calculateMinutesOverlappingInPeriods(workRange, *dutyRange)
		}

		return dutyRange.Count() - minutesWithGuardDuty
	}

	return minutesWithGuardDuty
}

func GuarddutySalary(plan models.Vaktplan, minWinTid models.MinWinTid) error {
	minutes, err := calculateMinutesToBeCompensated(plan.Schedule, minWinTid.Timesheet)
	if err != nil {
		return err
	}

	compensationTotal := compensation.Calculate(minutes, minWinTid.Satser)
	overtimeTotal, err := overtime.Calculate(minutes, minWinTid.Timesheet)
	if err != nil {
		return err
	}

	// TODO: Må returnere penger, og hvor mye per tillegg!
	fmt.Printf("Money earned %v + %v", compensationTotal, overtimeTotal)
	//report.Earnings.Compensation.Total = compensationTotal
	//report.Earnings.Overtime.Total = overtimeTotal
	//report.Earnings.Total = compensationTotal.Add(overtimeTotal)

	return nil
}
