package calculator

import (
	"fmt"
	"github.com/navikt/vaktor-lonn/pkg/dummy"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"math"
	"strings"
	"time"

	"github.com/vjeantet/eastertime"
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

// createRangeForPeriod creates a range of minutes based on two dates. This will fit the threshold used.
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
func ParsePeriode(report *models.Report, dutyPeriods map[string]models.Period, timesheet map[string][]string) (map[string]models.GuardDuty, error) {
	guardHours := map[string]models.GuardDuty{}

	for day, dutyPeriod := range dutyPeriods {
		date, err := time.Parse("02.01.2006", day)
		if err != nil {
			return guardHours, err
		}

		dutyHours := models.GuardDuty{}

		// sjekk om man har vakt i perioden 00-06
		dutyHours.Hvilende2006, err = calculateMinutesWithGuardDutyInPeriod(report, day, dutyPeriod, models.Period{Fra: "00:00", Til: "06:00"}, timesheet[day])
		if err != nil {
			return nil, err
		}
		modifier, err := calculateDaylightSavingTimeModifier(day)
		dutyHours.Hvilende2006 += modifier

		// sjekk om man har vakt i perioden 20-24
		minutesWorked, err := calculateMinutesWithGuardDutyInPeriod(report, day, dutyPeriod, models.Period{Fra: "20:00", Til: "24:00"}, timesheet[day])
		if err != nil {
			return nil, err
		}
		dutyHours.Hvilende2006 += minutesWorked

		// sjekk om man har vakt i perioden 06-20
		dutyHours.Hvilende0620, err = calculateMinutesWithGuardDutyInPeriod(report, day, dutyPeriod, models.Period{Fra: "06:00", Til: "20:00"}, timesheet[day])
		if err != nil {
			return nil, err
		}

		validateHowMuchDutyHours, err := validateHowMuchDutyHours(date, dutyPeriod.Helligdag)
		if err != nil {
			return nil, err
		}
		if validateHowMuchDutyHours {
			// TODO: En vaktperiode kan ikke være lengre enn 17t i døgnet mandag-fredag under sommertid, og 16t15m mandag-fredag under vintertid
			// Sjekk om en person har Hvilende0620 mer enn 8,5t eller Hvilende0620+Hvilende2006 mer enn 17t/16t15min.
			// Det er unntak følgende dager: onsdag før skjærtorsdag, julaften, romjulen, nyttårsaften

			// Dette er tiden du ikke jobbet i kjernetiden. Da vil man ikke kunne få vakttillegg, da andre er på jobb til å ta uforutsette hendelser.
			minutesNotWorkedInCoreWorkingHours := 0
			minutesNotWorkedInCoreWorkingHours, err = calculateMinutesWithGuardDutyInPeriod(report, day, dutyPeriod, models.Period{Fra: "09:00", Til: "14:30"}, timesheet[day])
			dutyHours.Hvilende0620 -= minutesNotWorkedInCoreWorkingHours
			report.MinutesNotWorkedinCoreWorkHours = minutesNotWorkedInCoreWorkingHours

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
				report.TooMuchDutyMinutes = maxDutyMinutes - addedDutyMinutes
				dutyHours.Hvilende0620 -= maxDutyMinutes - addedDutyMinutes
				if dutyHours.Hvilende0620 < 0 {
					dutyHours.Hvilende2006 += dutyHours.Hvilende0620
					dutyHours.Hvilende0620 = 0
				}
			}
		}

		if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday || dutyPeriod.Helligdag {
			// sjekk om man har vakt i perioden 00-24
			dutyHours.Helgetillegg, err = calculateMinutesWithGuardDutyInPeriod(report, day, dutyPeriod, models.Period{Fra: "00:00", Til: "24:00"}, timesheet[day])
			if err != nil {
				return nil, err
			}

			dutyHours.WeekendOrHolidayCompensation = true
		} else {
			// sjekk om man har vakt i perioden 06-07
			dutyHours.Skifttillegg, err = calculateMinutesWithGuardDutyInPeriod(report, day, dutyPeriod, models.Period{Fra: "06:00", Til: "07:00"}, timesheet[day])
			if err != nil {
				return nil, err
			}

			// sjekk om man har vakt i perioden 17-20
			minutesWorked, err = calculateMinutesWithGuardDutyInPeriod(report, day, dutyPeriod, models.Period{Fra: "17:00", Til: "20:00"}, timesheet[day])
			if err != nil {
				return nil, err
			}
			dutyHours.Skifttillegg += minutesWorked
		}

		guardHours[day] = dutyHours
		t := report.TimesheetEachDay[day]
		t.MinutesWithDuty = dutyHours
		report.TimesheetEachDay[day] = t
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
	date, err := time.Parse("02.01.2006", day)
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
func calculateMinutesWithGuardDutyInPeriod(report *models.Report, day string, dutyPeriod models.Period, compPeriod models.Period, timesheet []string) (int, error) {
	dutyRange, err := createRangeForPeriod(day, dutyPeriod.Fra, dutyPeriod.Til, compPeriod.Fra, compPeriod.Til)
	if err != nil {
		return 0, err
	}

	minutesWorked := 0

	if dutyRange != nil {
		for _, workHours := range timesheet {
			workRange, err := timeToRange(workHours)
			if err != nil {
				return 0, err
			}

			minutesWorked += calculateMinutesOverlappingInPeriods(workRange, *dutyRange)
		}

		timesheet := report.TimesheetEachDay[day]
		switch compPeriod.Fra {
		case "00:00":
			if compPeriod.Til == "06:00" {
				timesheet.MinutesWorked.Hvilende2006 += minutesWorked
			} else {
				timesheet.MinutesWorked.Helgetillegg += minutesWorked
			}
		case "06:00":
			if compPeriod.Til == "07:00" {
				timesheet.MinutesWorked.Skifttillegg += minutesWorked
			} else {
				timesheet.MinutesWorked.Hvilende0620 += minutesWorked
			}
		case "17:00":
			timesheet.MinutesWorked.Skifttillegg += minutesWorked
		case "20:00":
			timesheet.MinutesWorked.Hvilende2006 += minutesWorked
		}
		report.TimesheetEachDay[day] = timesheet
		return dutyRange.Count() - minutesWorked, nil
	}

	return minutesWorked, nil
}

func CalculateCompensation(report *models.Report, minutes map[string]models.GuardDuty) (float64, error) {
	var compensation models.GuardDuty

	for _, duty := range minutes {
		compensation.Hvilende0620 += duty.Hvilende0620
		compensation.Hvilende2006 += duty.Hvilende2006
		compensation.Skifttillegg += duty.Skifttillegg
		compensation.Helgetillegg += duty.Helgetillegg
	}

	report.GuardDutyMinutes.Hvilende0620 = compensation.Hvilende0620
	report.GuardDutyMinutes.Hvilende2006 = compensation.Hvilende2006
	report.GuardDutyMinutes.Skifttillegg = compensation.Skifttillegg
	report.GuardDutyMinutes.Helgetillegg = compensation.Helgetillegg
	report.GuardDutyHours.Hvilende0620 = int(math.Round(float64(compensation.Hvilende0620 / 60)))
	report.GuardDutyHours.Hvilende2006 = int(math.Round(float64(compensation.Hvilende2006 / 60)))
	report.GuardDutyHours.Skifttillegg = int(math.Round(float64(compensation.Skifttillegg / 60)))
	report.GuardDutyHours.Helgetillegg = int(math.Round(float64(compensation.Helgetillegg / 60)))

	// TODO: runne av til nærmeste 2 desimaler
	return math.Round(float64(compensation.Hvilende0620/60.0))*report.Satser["0620"] +
		math.Round(float64(compensation.Hvilende2006/60.0))*report.Satser["2006"] +
		(math.Round(float64(compensation.Helgetillegg/60.0)) * report.Satser["lørsøn"] / 5) +
		(math.Round(float64(compensation.Skifttillegg/60.0)) * report.Satser["utvidet"] / 5), nil
}

func CalculateOvertime(report *models.Report, minutes map[string]models.GuardDuty, salary float64) (float64, error) {
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

	report.OTS100 = ots100
	report.OTS50 = ots50

	overtimeWork := (math.Round(overtimeWorkDayMinutes/60.0)*ots50 + math.Round(overtimeWorkNightMinutes/60.0)*ots100) / 5.0
	overtimeWeekend := math.Round(overtimeWeekendMinutes/60.0) * ots100 / 5.0
	report.Earnings.Overtime.Work = overtimeWork
	report.Earnings.Overtime.Weekend = overtimeWeekend
	return overtimeWeekend + overtimeWork, nil
}

func CalculateEarnings(report *models.Report, minutes map[string]models.GuardDuty, salary int) error {
	compensation, err := CalculateCompensation(report, minutes)
	if err != nil {
		return err
	}
	overtime, err := CalculateOvertime(report, minutes, float64(salary))
	if err != nil {
		return err
	}

	report.Earnings.Compensation.Total = compensation
	report.Earnings.Overtime.Total = overtime
	report.Earnings.Total = compensation + overtime
	return nil
}

func GuarddutySalary(plan models.Plan) (models.Report, error) {
	minWinTid := dummy.GetMinWinTid(plan.Ident)
	salary := dummy.GetSalary(plan.Ident)

	report := &models.Report{
		Ident:            plan.Ident,
		Salary:           float64(salary),
		Satser:           plan.Satser,
		TimesheetEachDay: map[string]models.Timesheet{},
	}

	for day, work := range minWinTid {
		timesheet := models.Timesheet{
			GuardDuty: plan.Periods[day],
			Work:      work,
		}
		report.TimesheetEachDay[day] = timesheet
	}

	minutes, err := ParsePeriode(report, plan.Periods, minWinTid)
	if err != nil {
		return *report, err
	}

	err = CalculateEarnings(report, minutes, salary)
	return *report, err
}
