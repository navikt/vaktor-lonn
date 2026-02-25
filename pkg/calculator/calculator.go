package calculator

import (
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/navikt/vaktor-lonn/pkg/callout"
	"github.com/navikt/vaktor-lonn/pkg/ranges"

	"github.com/navikt/vaktor-lonn/pkg/kronetillegg"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/navikt/vaktor-lonn/pkg/overtime"
	"github.com/shopspring/decimal"
)

const (
	VaktorDateFormat = "2006-01-02"
)

// calculateMinutesToBePaid returns an object with the minutes you have been having guard duty each day in a given periode
func calculateMinutesToBePaid(schedule map[string][]models.Period, timesheet map[string]models.TimeSheet) (map[string]models.GuardDuty, error) {
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
				End:   time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC).Add(24 * time.Hour),
			}, currentDay.Clockings)
			dutyHours.Hvilende2000 += minutesWithGuardDuty

			// sjekk om man har vakt i perioden 06-20
			minutesWithGuardDuty = calculateMinutesWithGuardDutyInPeriod(period, models.Period{
				Begin: time.Date(date.Year(), date.Month(), date.Day(), 6, 0, 0, 0, time.UTC),
				End:   time.Date(date.Year(), date.Month(), date.Day(), 20, 0, 0, 0, time.UTC),
			}, currentDay.Clockings)
			dutyHours.Hvilende0620 += minutesWithGuardDuty

			// TODO: Disse modifiers burde begge trekkes fra, tungvindt å legge til et negativt tall
			kjernetidModifier := calculateGuardDutyInKjernetid(currentDay, period)
			dutyHours.Hvilende0620 -= kjernetidModifier

			// TODO: Lag en skikkelig test av denne
			maxGuardDutyModifier := calculateMaxGuardDutyTime(currentDay, dutyHours.Hvilende0620+dutyHours.Hvilende2000+dutyHours.Hvilende0006)
			dutyHours.Hvilende0620 += maxGuardDutyModifier

			if isWeekend(currentDay.Date) {
				// sjekk om man har vakt i perioden 00-24
				minutesWithGuardDuty = calculateMinutesWithGuardDutyInPeriod(period, models.Period{
					Begin: time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC),
					End:   time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC).Add(24 * time.Hour),
				}, currentDay.Clockings)
				dutyHours.Helgetillegg += minutesWithGuardDuty
			} else {
				// sjekk om man har vakt i perioden 06-07
				minutesWithGuardDuty = calculateMinutesWithGuardDutyInPeriod(period, models.Period{
					Begin: time.Date(date.Year(), date.Month(), date.Day(), 6, 0, 0, 0, time.UTC),
					End:   time.Date(date.Year(), date.Month(), date.Day(), 7, 0, 0, 0, time.UTC),
				}, currentDay.Clockings)
				dutyHours.Skifttillegg += minutesWithGuardDuty

				// sjekk om man har vakt i perioden 17-20
				minutesWithGuardDuty = calculateMinutesWithGuardDutyInPeriod(period, models.Period{
					Begin: time.Date(date.Year(), date.Month(), date.Day(), 17, 0, 0, 0, time.UTC),
					End:   time.Date(date.Year(), date.Month(), date.Day(), 20, 0, 0, 0, time.UTC),
				}, currentDay.Clockings)
				dutyHours.Skifttillegg += minutesWithGuardDuty
			}

			dutyHours.IsWeekend = isWeekend(currentDay.Date)

			// Det er ingen økonomiske fordeler med helligdager i helg, kun i ukedagene.
			// Derfor bryr vi oss ikke om helligdager i helgene.
			if !dutyHours.IsWeekend && isHoliday(currentDay.FormName) {
				if currentDay.FormName == "Helligdag" {
					dutyHours.Helligdag0620 = dutyHours.Hvilende0620
					dutyHours.Hvilende0620 = 0
				} else {
					// Tre dager i året er det kun helligdag etter kl12, så de må spesialhåndteres
					// det er kun tiden før kjernetid som er relevant for helligdager som starter kl12.
					if currentDay.FormName == "Nyttårsaften 1000-1200 *" {
						// Nyttårsaften har kjernetid fra kl10 til kl12
						minutesWithGuardDuty = calculateMinutesWithGuardDutyInPeriod(period, models.Period{
							Begin: time.Date(date.Year(), date.Month(), date.Day(), 6, 0, 0, 0, time.UTC),
							End:   time.Date(date.Year(), date.Month(), date.Day(), 10, 0, 0, 0, time.UTC),
						}, currentDay.Clockings)
					} else {
						// Julaften og onsdag før påske har kjernetid fra kl08 til kl12
						minutesWithGuardDuty = calculateMinutesWithGuardDutyInPeriod(period, models.Period{
							Begin: time.Date(date.Year(), date.Month(), date.Day(), 6, 0, 0, 0, time.UTC),
							End:   time.Date(date.Year(), date.Month(), date.Day(), 8, 0, 0, 0, time.UTC),
						}, currentDay.Clockings)
					}
					dutyHours.Helligdag0620 = dutyHours.Hvilende0620 - minutesWithGuardDuty
					dutyHours.Hvilende0620 = minutesWithGuardDuty
				}
			}
		}
		guardHours[day] = dutyHours
	}

	return guardHours, nil
}

// calculateMaxGuardDutyTime fjerner minutter som overstiger lovlig antall tid med vakt man kan gå per dag.
func calculateMaxGuardDutyTime(currentDay models.TimeSheet, totalGuardDutyInADayInMinutes float64) float64 {
	if isWeekend(currentDay.Date) || currentDay.FormName == "Helligdag" {
		return 0
	}

	maxGuardDutyInMinutes := 24*60 - currentDay.WorkingHours*60
	if totalGuardDutyInADayInMinutes > maxGuardDutyInMinutes {
		return maxGuardDutyInMinutes - totalGuardDutyInADayInMinutes
	}

	return 0
}

func isWeekend(day time.Time) bool {
	return day.Weekday() == time.Saturday || day.Weekday() == time.Sunday
}

func isHoliday(formName string) bool {
	holidays := []string{"Helligdag", "Julaften 0800-1200 *", "Onsdag før Påske 0800-1200 *", "Nyttårsaften 1000-1200 *"}
	return slices.Contains(holidays, formName)
}

// calculateGuardDutyInKjernetid sjekker om man hadde vakt i kjernetiden. Man vil ikke kunne få vakttillegg i
// kjernetiden, da andre skal være på jobb til å ta seg av uforutsette hendelser.
func calculateGuardDutyInKjernetid(currentDay models.TimeSheet, period models.Period) float64 {
	if isWeekend(currentDay.Date) || currentDay.FormName == "Helligdag" {
		return 0
	}

	kjernetid := createKjernetid(currentDay.Date, currentDay.FormName)
	return calculateMinutesWithGuardDutyInPeriod(period, kjernetid, currentDay.Clockings)
}

// createKjernetid returns the current day kjernetid. Except for three days, it's always from 0900 till 1430
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
			if workHours.OtG {
				// Overtid ved utrykning regnes ikke som arbeidstid
				continue
			}

			workRange := ranges.FromTime(workHours.In, workHours.Out)
			minutesWithGuardDuty += ranges.CalculateMinutesOverlapping(workRange, *dutyRange)
		}

		return dutyRange.Count() - minutesWithGuardDuty
	}

	return minutesWithGuardDuty
}

func getDailySalaries(timesheet map[string]models.TimeSheet) map[string][]string {
	salaries := make(map[string][]string)
	for date, period := range timesheet {
		key := period.Salary.String()
		if _, ok := salaries[key]; !ok {
			salaries[key] = []string{date}
		} else {
			salaries[key] = append(salaries[key], date)
		}
	}

	return salaries
}

func getStillingskode(timesheet map[string]models.TimeSheet) (string, error) {
	var stillingskode string
	for _, period := range timesheet {
		if stillingskode == "" {
			stillingskode = period.Stillingskode
			continue
		}
		if stillingskode != period.Stillingskode {
			return "", fmt.Errorf("stillingskode has changed")
		}
	}

	return stillingskode, nil
}

func GuarddutySalary(plan models.Vaktplan, minWinTid models.MinWinTid) (models.Payroll, error) {
	minutes, err := calculateMinutesToBePaid(plan.Schedule, minWinTid.Timesheet)
	if err != nil {
		return models.Payroll{}, err
	}

	stillingskode, err := getStillingskode(minWinTid.Timesheet)
	if err != nil {
		return models.Payroll{}, err
	}

	payroll := &models.Payroll{
		ID:            plan.ID,
		ApproverID:    minWinTid.ApproverID,
		ApproverName:  minWinTid.ApproverName,
		CommitSHA:     os.Getenv("NAIS_APP_IMAGE"),
		Stillingskode: stillingskode,
	}

	kronetillegg.Calculate(minutes, minWinTid.Satser, payroll)
	callout.Calculate(plan.Schedule, minWinTid.Timesheet, minWinTid.Satser, payroll)

	salariesWithDates := getDailySalaries(minWinTid.Timesheet)
	if err != nil {
		return models.Payroll{}, err
	}

	for salaryAsString, dates := range salariesWithDates {
		salaryBasedMinutes := make(map[string]models.GuardDuty)
		for _, date := range dates {
			salaryBasedMinutes[date] = minutes[date]
		}

		salary, err := decimal.NewFromString(salaryAsString)
		if err != nil {
			return models.Payroll{}, err
		}

		overtime.Calculate(salaryBasedMinutes, salary, payroll)
	}

	return *payroll, nil
}
