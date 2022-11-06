package calculator

import (
	"fmt"
	"github.com/navikt/vaktor-lonn/pkg/ranges"
	"os"
	"time"

	"github.com/navikt/vaktor-lonn/pkg/compensation"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/navikt/vaktor-lonn/pkg/overtime"
	"github.com/shopspring/decimal"
)

const (
	VaktorDateFormat = "2006-01-02"
)

// calculateMinutesToBeCompensated returns an object with the minutes you have been having guard duty each day in a given periode
func calculateMinutesToBeCompensated(schedule map[string][]models.Period, timesheet map[string]models.TimeSheet) (map[string]models.GuardDuty, error) {
	guardHours := map[string]models.GuardDuty{}

	for day, periods := range schedule {
		currentDay := timesheet[day]
		date := currentDay.Date
		dutyHours := models.GuardDuty{}

		modifier := calculateDaylightSavingTimeModifier(periods, date)
		dutyHours.Hvilende0006 += modifier
		dutyHours.Helgetillegg += modifier

		for _, period := range periods {
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

			// TODO: Disse modifiers burde begge trekkes fra, tungvindt å legge til et negativt tall
			// TODO: Lag en skikkelig test av denne
			kjernetidModifier := calculateGuardDutyInKjernetid(currentDay, date, period)
			dutyHours.Hvilende0620 -= kjernetidModifier

			// TODO: Lag en skikkelig test av denne
			maxGuardDutyModifier := calculateMaxGuardDutyTime(currentDay, dutyHours.Hvilende0620+dutyHours.Hvilende2000+dutyHours.Hvilende0006)
			dutyHours.Hvilende0620 += maxGuardDutyModifier

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

// calculateDaylightSavingTimeModifier returns either -60 or 60 minutes if $day is when the clock is advanced
func calculateDaylightSavingTimeModifier(periods []models.Period, date time.Time) float64 {
	nightShift := models.Period{
		Begin: time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC),
		End:   time.Date(date.Year(), date.Month(), date.Day(), 6, 0, 0, 0, time.UTC),
	}

	var minutes float64
	for _, period := range periods {
		minutes += calculateMinutesWithGuardDutyInPeriod(period, nightShift, []models.Clocking{})
	}
	if minutes == 0 {
		return 0
	}

	summerTime := time.Date(date.Year(), time.March, 31, 0, 0, 0, 0, time.UTC)
	summerTime = summerTime.AddDate(0, 0, -int(summerTime.Weekday()))
	if summerTime.YearDay() == date.YearDay() {
		return -60
	}

	winterTime := time.Date(date.Year(), time.October, 31, 0, 0, 0, 0, time.UTC)
	winterTime = winterTime.AddDate(0, 0, -int(winterTime.Weekday()))
	if winterTime.YearDay() == date.YearDay() {
		return 60
	}

	return 0
}

// calculateMinutesWithGuardDutyInPeriod return the number of minutes that you have non-working guard duty
func calculateMinutesWithGuardDutyInPeriod(vaktPeriod models.Period, compPeriod models.Period, timesheet []models.Clocking) float64 {
	dutyRange := ranges.CreateForPeriod(vaktPeriod, compPeriod)
	minutesWithGuardDuty := 0.0

	if dutyRange != nil {
		for _, workHours := range timesheet {
			workRange := ranges.FromTime(workHours.In, workHours.Out)
			minutesWithGuardDuty += ranges.CalculateMinutesOverlapping(workRange, *dutyRange)
		}

		return dutyRange.Count() - minutesWithGuardDuty
	}

	return minutesWithGuardDuty
}

func getFormal(timesheet map[string]models.TimeSheet) (string, error) {
	var formal string
	for _, period := range timesheet {
		if formal == "" {
			formal = period.Formal
			continue
		}
		if formal != period.Formal {
			return "", fmt.Errorf("formål has changed")
		}
	}

	return formal, nil
}

func getAktivitet(timesheet map[string]models.TimeSheet) (string, error) {
	var aktivitet string
	for _, period := range timesheet {
		if aktivitet == "" {
			aktivitet = period.Aktivitet
			continue
		}
		if aktivitet != period.Aktivitet {
			return "", fmt.Errorf("aktivitet has changed")
		}
	}

	return aktivitet, nil
}

func getKoststed(timesheet map[string]models.TimeSheet) (string, error) {
	var koststed string
	for _, period := range timesheet {
		if koststed == "" {
			koststed = period.Koststed
			continue
		}
		if koststed != period.Koststed {
			return "", fmt.Errorf("koststed has changed")
		}
	}

	return koststed, nil
}

func GuarddutySalary(plan models.Vaktplan, minWinTid models.MinWinTid) (models.Payroll, error) {
	minutes, err := calculateMinutesToBeCompensated(plan.Schedule, minWinTid.Timesheet)
	if err != nil {
		return models.Payroll{}, err
	}

	formal, err := getFormal(minWinTid.Timesheet)
	if err != nil {
		return models.Payroll{}, err
	}

	aktivitet, err := getAktivitet(minWinTid.Timesheet)
	if err != nil {
		return models.Payroll{}, err
	}

	koststed, err := getKoststed(minWinTid.Timesheet)
	if err != nil {
		return models.Payroll{}, err
	}

	var payroll *models.Payroll
	payroll = &models.Payroll{
		ID:           plan.ID,
		ApproverID:   minWinTid.ApproverID,
		ApproverName: minWinTid.ApproverName,
		TypeCodes: map[string]decimal.Decimal{
			models.ArtskodeMorgen: {},
			models.ArtskodeDag:    {},
			models.ArtskodeKveld:  {},
			models.ArtskodeHelg:   {},
		},
		CommitSHA: os.Getenv("NAIS_APP_IMAGE"),
		Formal:    formal,
		Koststed:  koststed,
		Aktivitet: aktivitet,
	}

	compensation.Calculate(minutes, minWinTid.Satser, payroll)
	err = overtime.Calculate(minutes, minWinTid.Timesheet, payroll)
	if err != nil {
		return models.Payroll{}, err
	}

	return *payroll, nil
}
