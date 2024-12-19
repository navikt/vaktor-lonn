package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/navikt/vaktor-lonn/pkg/calculator"
	"github.com/navikt/vaktor-lonn/pkg/models"
	gensql "github.com/navikt/vaktor-lonn/pkg/sql/gen"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

const (
	DateTimeFormat  = "2006-01-02T15:04:05"
	fravarKodeFerie = 210
	vaktplanId      = "vaktplanId"
)

func getTimesheetFromMinWinTid(ident string, periodBegin time.Time, periodEnd time.Time, handler Handler) (models.MWTRespons, error) {
	config := handler.MinWinTidConfig

	bearerToken, err := config.BearerClient.GenerateBearerToken()
	if err != nil {
		return models.MWTRespons{}, err
	}

	req, err := http.NewRequest(http.MethodGet, config.Endpoint, nil)
	if err != nil {
		return models.MWTRespons{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bearerToken))
	values := req.URL.Query()
	values.Add("nav_id", ident)
	values.Add("fra_dato", periodBegin.Format(DateTimeFormat))
	values.Add("til_dato", periodEnd.Format(DateTimeFormat))
	req.URL.RawQuery = values.Encode()

	backoffSchedule := []time.Duration{
		1 * time.Second,
		3 * time.Second,
		10 * time.Second,
	}

	resp, err := handler.Client.Do(req)
	if err != nil {
		for _, duration := range backoffSchedule {
			handler.Log.Info("Problem connecting to MinWinTid", zap.Error(err))
			time.Sleep(duration)
			resp, err = handler.Client.Do(req)
			if err == nil {
				break
			}
		}

		if err != nil {
			return models.MWTRespons{}, err
		}
	}

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return models.MWTRespons{}, err
		}

		return models.MWTRespons{}, fmt.Errorf("minWinTid returned http(%v): %v", resp.StatusCode, string(body))
	}

	var response models.MWTRespons
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return models.MWTRespons{}, fmt.Errorf("decoding MinWinTid response: %w", err)
	}

	dager := response.Dager
	sort.SliceStable(dager, func(i, j int) bool {
		return dager[i].Dato < dager[j].Dato
	})
	response.Dager = dager

	return response, nil
}

func isTimesheetApproved(days []models.MWTDag) error {
	for _, day := range days {
		if day.Godkjent < 2 {
			return fmt.Errorf("clocking %v has status %v, should be 2", day.Dato, day.Godkjent)
		}
	}

	return nil
}

func isThereRegisteredVacationAtTheSameTimeAsGuardDuty(days []models.MWTDag, vaktplan models.Vaktplan) (bool, error) {
	for _, day := range days {
		for _, stempling := range day.Stemplinger {
			// Denne tar ikke høyde for planlagt ferie over lengre tid
			if stempling.Fravarkode == fravarKodeFerie {
				date, err := time.Parse(DateTimeFormat, stempling.StemplingTid)
				if err != nil {
					return false, err
				}
				if len(vaktplan.Schedule[date.Format(calculator.VaktorDateFormat)]) > 0 {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func createClocking(innTid, utTid string) (models.Clocking, error) {
	innStemplingDate, err := time.Parse(DateTimeFormat, innTid)
	if err != nil {
		return models.Clocking{}, err
	}

	utStemplingDate, err := time.Parse(DateTimeFormat, utTid)
	if err != nil {
		return models.Clocking{}, err
	}

	return models.Clocking{In: innStemplingDate, Out: utStemplingDate}, nil
}

func formatTimesheet(days []models.MWTDag) (map[string]models.TimeSheet, []zap.Field) {
	timesheet := make(map[string]models.TimeSheet)
	var nextDay []models.Clocking

	for _, day := range days {
		stemplingDate, err := time.Parse(DateTimeFormat, day.Dato)
		if err != nil {
			return nil, []zap.Field{zap.Error(err)}
		}
		simpleStemplingDate := stemplingDate.Format(calculator.VaktorDateFormat)
		stilling := day.Stillinger[0]

		ts := models.TimeSheet{
			Date:          stemplingDate,
			WorkingHours:  day.SkjemaTid,
			WorkingDay:    day.Virkedag,
			FormName:      day.SkjemaNavn,
			Salary:        decimal.NewFromInt(int64(stilling.RATEK001)),
			Stillingskode: stilling.Stillingskode,
			Koststed:      stilling.Koststed,
			Formal:        stilling.Formal,
			Aktivitet:     stilling.Aktivitet,
			Clockings:     []models.Clocking{},
		}

		if len(nextDay) != 0 {
			ts.Clockings = append(ts.Clockings, nextDay...)
			nextDay = []models.Clocking{}
		}

		stemplinger := day.Stemplinger
		if len(stemplinger) == 0 && day.SkjemaTid != 0 {
			ts.Clockings = append(ts.Clockings, createPerfectClocking(day.SkjemaTid, stemplingDate))
		}

		if len(stemplinger) == 1 {
			return nil, []zap.Field{
				zap.Error(fmt.Errorf("there are too few clockings")),
				zap.Any("stemplinger", day.Stemplinger),
			}
		}

		sort.SliceStable(stemplinger, func(i, j int) bool {
			return stemplinger[i].StemplingTid < stemplinger[j].StemplingTid
		})

		for len(stemplinger) >= 2 {
			innStempling := stemplinger[0]
			stemplinger = stemplinger[1:]

			utStempling := stemplinger[0]
			stemplinger = stemplinger[1:]

			if innStempling.Retning == "Inn" && innStempling.Type == "B1" {
				// Dette er en vanlig stempling
				if utStempling.Retning == "Ut" && utStempling.Type == "B2" {
					clocking, err := createClocking(innStempling.StemplingTid, utStempling.StemplingTid)
					if err != nil {
						return nil, []zap.Field{zap.Error(err)}
					}

					ts.Clockings = append(ts.Clockings, clocking)
					continue
				}

				// Dette er en stempling med overtid
				if utStempling.Retning == "Overtid" && utStempling.Type == "B6" {
					nesteStempling := utStempling
					var overtimeBecauseOfGuardDuty bool
					// Man kan ha flere overtidsstemplinger etter hverandre, så vi må sjekke om minst en av dem er BV
					for nesteStempling.Retning == "Overtid" && nesteStempling.Type == "B6" {
						if !overtimeBecauseOfGuardDuty {
							overtimeBecauseOfGuardDuty = strings.Contains(strings.ToLower(nesteStempling.OvertidBegrunnelse), "bv")
						}

						nesteStempling = stemplinger[0]
						stemplinger = stemplinger[1:]
					}

					utOvertid := nesteStempling

					// Før 1. februar 2023 så var det ikke krav om å merke overtiden sin med BV
					if stemplingDate.Before(time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)) {
						overtimeBecauseOfGuardDuty = true
					}

					if utOvertid.Retning == "Ut" && utOvertid.Type == "B2" {
						innStemplingDate, err := time.Parse(DateTimeFormat, innStempling.StemplingTid)
						if err != nil {
							return nil, []zap.Field{zap.Error(err)}
						}

						utStemplingDate, err := time.Parse(DateTimeFormat, utOvertid.StemplingTid)
						if err != nil {
							return nil, []zap.Field{zap.Error(err)}
						}

						if utStemplingDate.YearDay() > innStemplingDate.YearDay() &&
							(utStemplingDate.Hour() != 0 || utStemplingDate.Minute() != 0) {
							// Overtid over midnatt, flytter resten av tiden til neste dag
							truncateOut := utStemplingDate.Truncate(24 * time.Hour)
							nextDay = append(nextDay, models.Clocking{
								In:  truncateOut,
								Out: utStemplingDate,
								OtG: overtimeBecauseOfGuardDuty,
							})
							utStemplingDate = truncateOut
						}

						ts.Clockings = append(ts.Clockings, models.Clocking{
							In:  innStemplingDate,
							Out: utStemplingDate,
							OtG: overtimeBecauseOfGuardDuty,
						})
						continue
					}
					return nil, []zap.Field{
						zap.Error(fmt.Errorf("did not get expected overtime clock-out, got direction=%v and type=%v", utOvertid.Retning, utOvertid.Type)),
						zap.Any("stemplinger", day.Stemplinger),
					}
				}

				// Dette er en stempling ut på fravær
				if utStempling.Retning == "Ut på fravær" && utStempling.Type == "B5" {
					innDate, err := time.Parse(DateTimeFormat, innStempling.StemplingTid)
					if err != nil {
						return nil, []zap.Field{zap.Error(err)}
					}
					utDate, err := time.Parse(DateTimeFormat, utStempling.StemplingTid)
					if err != nil {
						return nil, []zap.Field{zap.Error(err)}
					}

					// Dette er en heldagsstempling
					if (innDate.Hour() == 8 && innDate.Minute() == 0 && innDate.Second() == 0) &&
						(utDate.Hour() == 8 && utDate.Minute() == 0 && utDate.Second() == 1) {
						date, err := time.Parse(DateTimeFormat, innStempling.StemplingTid)
						if err != nil {
							return nil, []zap.Field{zap.Error(err)}
						}

						workdayLengthRestMinutes := int(math.Mod(ts.WorkingHours, 1) * 60)

						ts.Clockings = append(ts.Clockings, models.Clocking{
							In:  date,
							Out: time.Date(date.Year(), date.Month(), date.Day(), 15, workdayLengthRestMinutes, 0, 0, time.UTC),
						})

						if len(stemplinger) >= 2 {
							innFravar := stemplinger[0]
							utFravar := stemplinger[1]
							if innFravar.Retning == "Inn fra fravær" && innFravar.Type == "B4" &&
								utFravar.Retning == "Ut" && utFravar.Type == "B2" {
								stemplinger = stemplinger[2:]
							}
						}
						continue
					}

					clocking, err := createClocking(innStempling.StemplingTid, utStempling.StemplingTid)
					if err != nil {
						return nil, []zap.Field{zap.Error(err)}
					}

					ts.Clockings = append(ts.Clockings, clocking)
					continue
				}
			} else if innStempling.Retning == "Inn fra fravær" && innStempling.Type == "B4" &&
				(utStempling.Retning == "Ut" && utStempling.Type == "B2" ||
					utStempling.Retning == "Ut på fravær" && utStempling.Type == "B5") {
				// Fravær i arbeidstid
				clocking, err := createClocking(innStempling.StemplingTid, utStempling.StemplingTid)
				if err != nil {
					return nil, []zap.Field{zap.Error(err)}
				}

				ts.Clockings = append(ts.Clockings, clocking)
				continue
			}

			return nil, []zap.Field{
				zap.Error(fmt.Errorf("did not get expected direction or type, got inn{direction=%v, type=%v} and out{direction=%v, type=%v}", innStempling.Retning, innStempling.Type, utStempling.Retning, utStempling.Type)),
				zap.Any("stemplinger", day.Stemplinger),
			}
		}

		if len(stemplinger) != 0 {
			return nil, []zap.Field{
				zap.Error(fmt.Errorf("there are clockings left")),
				zap.Any("stemplinger", day.Stemplinger),
			}
		}

		timesheet[simpleStemplingDate] = ts
	}
	return timesheet, nil
}

func createPerfectClocking(tid float64, date time.Time) models.Clocking {
	in := time.Date(date.Year(), date.Month(), date.Day(), 8, 0, 0, 0, time.UTC)
	out := in.Add(time.Duration(tid*60) * time.Minute)
	return models.Clocking{
		In:  in,
		Out: out,
	}
}

func postError(handler Handler, beredskapsvakt gensql.Beredskapsvakt, message, bearerToken string) error {
	blob := map[string]string{
		"error": message,
		"ok":    "false",
	}

	payload, err := json.Marshal(blob)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%v/error", beredskapsvakt.ID)

	return postToPlan(handler, payload, url, bearerToken)
}

func postPayroll(handler Handler, payroll models.Payroll, bearerToken string) error {
	payload, err := json.Marshal(payroll)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("confirm_calculations/%v", payroll.ID)

	return postToPlan(handler, payload, url, bearerToken)
}

func postToPlan(handler Handler, payload []byte, url, bearerToken string) error {
	req, err := http.NewRequest(http.MethodPost, handler.VaktorPlanEndpoint+url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", fmt.Sprintf("bearer %v", bearerToken))

	response, err := handler.Client.Do(req)
	if err != nil {
		return err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			handler.Log.Error("Failed while closing body", zap.Error(err))
		}
	}(response.Body)

	if response.StatusCode != http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("vaktorPlan returned http(%v) with body: %v", response.StatusCode, string(body))
	}

	return nil
}

func calculateSalary(beredskapsvakt gensql.Beredskapsvakt, tiddataResult models.MWTRespons) (*models.Payroll, string, error) {
	if err := isTimesheetApproved(tiddataResult.Dager); err != nil {
		return nil, "Timelisten din er ikke godkjent", nil
	}

	var vaktplan models.Vaktplan
	if err := json.Unmarshal(beredskapsvakt.Plan, &vaktplan); err != nil {
		return nil, "Ukjent data fra Vaktor Plan", fmt.Errorf("unmarshaling beredskapsvaktperiode: %w", err)
	}

	vacationAtTheSameTimeAsGuardDuty, err := isThereRegisteredVacationAtTheSameTimeAsGuardDuty(tiddataResult.Dager, vaktplan)
	if err != nil {
		return nil, "Klarte ikke sjekke om du har hatt ferie under beredkapsvakt", fmt.Errorf("parsing date from MinWinTid: %w", err)
	}
	if vacationAtTheSameTimeAsGuardDuty {
		return nil, "Du har hatt ferie under beredskapsvakt", fmt.Errorf("user has had guard duty during vacation")
	}

	timesheet, errFields := formatTimesheet(tiddataResult.Dager)
	if len(errFields) != 0 {
		return nil, "Data fra MinWinTid er ikke gyldig", fmt.Errorf("tried to create timesheet: %v", errFields)
	}

	minWinTid := models.MinWinTid{
		ResourceID:   tiddataResult.NavID,
		ApproverID:   tiddataResult.LederNavID,
		ApproverName: tiddataResult.LederNavn,
		Satser: models.Satser{
			Helg:    decimal.NewFromInt(65),
			Dag:     decimal.NewFromInt(15),
			Natt:    decimal.NewFromInt(25),
			Utvidet: decimal.NewFromInt(25),
		},
		Timesheet: timesheet,
	}

	payroll, err := calculator.GuarddutySalary(vaktplan, minWinTid)
	if err != nil {
		return nil, "Klarte ikke å beregne utbetaling", fmt.Errorf("calculating guard duty salary: %w", err)
	}

	return &payroll, "", nil
}

func handleTransaction(handler Handler, beredskapsvakt gensql.Beredskapsvakt) {
	bearerToken, err := handler.BearerClient.GenerateBearerToken()
	if err != nil {
		handler.Log.Error("Problem generating bearer token", zap.Error(err), zap.String(vaktplanId, beredskapsvakt.ID.String()))
	}

	response, err := getTimesheetFromMinWinTid(beredskapsvakt.Ident, beredskapsvakt.PeriodBegin, beredskapsvakt.PeriodEnd, handler)
	if err != nil {
		handler.Log.Error("Failed while retrieving data from MinWinTid", zap.Error(err), zap.String(vaktplanId, beredskapsvakt.ID.String()))
		return
	}

	payroll, message, err := calculateSalary(beredskapsvakt, response)
	if err != nil || message != "" {
		handler.Log.Info("calculateSalary feilet, sender info til Plan", zap.Error(err), zap.String(vaktplanId, beredskapsvakt.ID.String()), zap.String("message", message))
		if err := postError(handler, beredskapsvakt, message, bearerToken); err != nil {
			handler.Log.Error("Failed while posting error to Vaktor Plan", zap.Error(err), zap.String(vaktplanId, beredskapsvakt.ID.String()))
		}

		return
	}

	if payroll == nil {
		return
	}

	if err := postPayroll(handler, *payroll, bearerToken); err != nil {
		handler.Log.Error("Failed while posting to Vaktor Plan", zap.Error(err), zap.String(vaktplanId, beredskapsvakt.ID.String()))
		return
	}

	err = handler.Queries.DeletePlan(handler.Context, beredskapsvakt.ID)
	if err != nil {
		handler.Log.Error("Failed while deleting beredskapsvakt", zap.Error(err), zap.String(vaktplanId, beredskapsvakt.ID.String()))
		return
	}
}

func handleTransactions(handler Handler) error {
	beredskapsvakter, err := handler.Queries.ListBeredskapsvakter(handler.Context)
	if err != nil {
		return err
	}

	for _, beredskapsvakt := range beredskapsvakter {
		handleTransaction(handler, beredskapsvakt)
	}

	return nil
}

func Run(handler Handler) {
	ticker := time.NewTicker(handler.MinWinTidConfig.TickerInterval)
	defer ticker.Stop()

	for {
		err := handleTransactions(handler)
		if err != nil {
			handler.Log.Error("Failed while handling transactions", zap.Error(err))
		}

		select {
		case <-handler.Context.Done():
			return
		case <-ticker.C:
		}
	}
}
