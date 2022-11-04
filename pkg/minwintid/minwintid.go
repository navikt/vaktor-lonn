package minwintid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/navikt/vaktor-lonn/pkg/calculator"
	"github.com/navikt/vaktor-lonn/pkg/endpoints"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"io"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"
)

const (
	DateTimeFormat  = "2006-01-02T15:04:05"
	fravarKodeFerie = 210
	vaktplanId      = "vaktplanId"
)

func getTimesheetFromMinWinTid(ident string, periodBegin time.Time, periodEnd time.Time, handler endpoints.Handler) (*http.Response, error) {
	config := handler.MinWinTidConfig
	req, err := http.NewRequest(http.MethodGet, config.Endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(config.Username, config.Password)
	values := req.URL.Query()
	values.Add("ident", ident)
	values.Add("fra_dato", periodBegin.Format(calculator.VaktorDateFormat))
	values.Add("til_dato", periodEnd.Format(calculator.VaktorDateFormat))
	req.URL.RawQuery = values.Encode()

	backoffSchedule := []time.Duration{
		1 * time.Second,
		3 * time.Second,
		10 * time.Second,
	}

	r, err := handler.Client.Do(req)
	if err != nil {
		for _, duration := range backoffSchedule {
			handler.Log.Info("Problem connecting to MinWinTid", zap.Error(err))
			time.Sleep(duration)
			r, err = handler.Client.Do(req)
			if err == nil {
				break
			}
		}

		if err != nil {
			return nil, err
		}
	}

	if r.StatusCode != http.StatusOK {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("minWinTid returned http(%v) with body: %v", r.StatusCode, string(body))
	}

	return r, nil
}

func isTimesheetApproved(days []Dag) bool {
	for _, day := range days {
		if day.Godkjent == 0 {
			return false
		}
	}

	return true
}

func isThereRegisteredVacationAtTheSameTimeAsGuardDuty(days []Dag, vaktplan models.Vaktplan) (bool, error) {
	for _, day := range days {
		for _, stempling := range day.Stemplinger {
			// TODO: Denne tar ikke høyde for planlagt ferie over lengre tid
			if stempling.FravarKode == fravarKodeFerie {
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

func formatTimesheet(days []Dag) (map[string]models.TimeSheet, []zap.Field) {
	timesheet := make(map[string]models.TimeSheet)
	var nextDay []models.Clocking

	for _, day := range days {
		stemplingDate, err := time.Parse(DateTimeFormat, day.Dato)
		if err != nil {
			return nil, []zap.Field{zap.Error(err)}
		}
		simpleStemplingDate := stemplingDate.Format(calculator.VaktorDateFormat)
		stillig := day.Stillinger[0]

		ts := models.TimeSheet{
			Date:         stemplingDate,
			WorkingHours: day.SkjemaTid,
			WorkingDay:   day.Virkedag,
			FormName:     day.SkjemaNavn,
			Salary:       decimal.NewFromInt(int64(stillig.RATEK001)),
			Koststed:     stillig.Koststed,
			Formal:       stillig.Formal,
			Aktivitet:    stillig.Aktivitet,
			Clockings:    []models.Clocking{},
		}

		if len(nextDay) != 0 {
			ts.Clockings = append(ts.Clockings, nextDay...)
			nextDay = []models.Clocking{}
		}

		stemplinger := day.Stemplinger
		if len(stemplinger) == 1 {
			return nil, []zap.Field{zap.Error(fmt.Errorf("there are too few clockings")),
				zap.Any("stemplinger", day.Stemplinger)}
		}

		if len(stemplinger) > 0 {
			sort.SliceStable(stemplinger, func(i, j int) bool {
				return stemplinger[i].StemplingTid < stemplinger[j].StemplingTid
			})

			for len(stemplinger) > 1 {
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

					// TODO: Det vil være en sjekk i Vaktor som sjekker om de gamle kodene er brukt
					// Dette er en stempling med overtid
					if utStempling.Retning == "Overtid                 " && utStempling.Type == "B6" {
						utOvertid := stemplinger[0]
						stemplinger = stemplinger[1:]

						if utOvertid.Retning == "Ut" && utOvertid.Type == "B2" {
							innStemplingDate, err := time.Parse(DateTimeFormat, innStempling.StemplingTid)
							if err != nil {
								return nil, []zap.Field{zap.Error(err)}
							}

							utStemplingDate, err := time.Parse(DateTimeFormat, utOvertid.StemplingTid)
							if err != nil {
								return nil, []zap.Field{zap.Error(err)}
							}

							overtimeBecauseOfGuardDuty := strings.EqualFold(utStempling.OvertidBegrunnelse, "BV")

							if utStemplingDate.YearDay() > innStemplingDate.YearDay() &&
								!(utStemplingDate.Hour() == 0 && utStemplingDate.Minute() == 00) {
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
						return nil, []zap.Field{zap.Error(fmt.Errorf("did not get expected overtime clock-out, got direction=%v and type=%v", utOvertid.Retning, utOvertid.Type)),
							zap.Any("stemplinger", day.Stemplinger)}
					}

					// Dette er en stempling med fravær
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
							continue
						}

						if len(stemplinger) >= 2 {
							innFravar := stemplinger[0]
							utFravar := stemplinger[1]

							// Fravær i arbeidstid
							if innFravar.Retning == "Inn fra fravær" && innFravar.Type == "B4" &&
								utFravar.Retning == "Ut" && utFravar.Type == "B2" {
								stemplinger = stemplinger[2:]

								clocking, err := createClocking(innStempling.StemplingTid, utFravar.StemplingTid)
								if err != nil {
									return nil, []zap.Field{zap.Error(err)}
								}

								ts.Clockings = append(ts.Clockings, clocking)
								continue
							}
						}

						clocking, err := createClocking(innStempling.StemplingTid, utStempling.StemplingTid)
						if err != nil {
							return nil, []zap.Field{zap.Error(err)}
						}

						ts.Clockings = append(ts.Clockings, clocking)
						continue
					}

					return nil, []zap.Field{zap.Error(fmt.Errorf("unknown clocking out(direction=%v, type=%v)", utStempling.Retning, utStempling.Type)),
						zap.Any("stemplinger", day.Stemplinger)}
				}

				return nil, []zap.Field{zap.Error(fmt.Errorf("did not get expected direction or type, got inn{direction=%v, type=%v} and out{direction=%v, type=%v}", innStempling.Retning, innStempling.Type, utStempling.Retning, utStempling.Type)),
					zap.Any("stemplinger", day.Stemplinger)}
			}

			if len(stemplinger) != 0 {
				return nil, []zap.Field{zap.Error(fmt.Errorf("there are clockings left")),
					zap.Any("stemplinger", day.Stemplinger)}
			}
		}

		timesheet[simpleStemplingDate] = ts
	}
	return timesheet, nil
}

func postToVaktorPlan(handler endpoints.Handler, payroll models.Payroll, bearerToken string) error {
	bufferBody, err := json.Marshal(payroll)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%v/%v", handler.VaktorPlanEndpoint, payroll.ID)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bufferBody))
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

func decodeMinWinTid(httpResponse *http.Response) (TiddataResult, error) {
	var response Response
	err := json.NewDecoder(httpResponse.Body).Decode(&response)
	if err != nil {
		return TiddataResult{}, err
	}

	results := response.VaktorVaktorTiddataResponse.VaktorVaktorTiddataResult
	if len(results) != 1 {
		return TiddataResult{}, fmt.Errorf("not enough data from MinWinTid, missing TiddataResult")
	}

	result := results[0]
	var dager []Dag
	err = json.Unmarshal([]byte(result.VaktorDager), &dager)
	if err != nil {
		return TiddataResult{}, err
	}

	result.Dager = dager
	result.VaktorDager = ""

	return result, err
}

func handleTransactions(handler endpoints.Handler) error {
	bearerToken, err := handler.BearerClient.GenerateBearerToken()
	if err != nil {
		handler.Log.Error("Problem generating bearer token", zap.Error(err))
	}

	beredskapsvakter, err := handler.Queries.ListBeredskapsvakter(handler.Context)
	if err != nil {
		handler.Log.Error("Failed while listing beredskapsvakter", zap.Error(err))
		return err
	}

	for _, beredskapsvakt := range beredskapsvakter {
		httpResponse, err := getTimesheetFromMinWinTid(beredskapsvakt.Ident, beredskapsvakt.PeriodBegin, beredskapsvakt.PeriodEnd, handler)
		if err != nil {
			handler.Log.Error("Failed while retrieving data from MinWinTid", zap.Error(err), zap.String(vaktplanId, beredskapsvakt.ID.String()))
			continue
		}

		tiddataResult, err := decodeMinWinTid(httpResponse)
		if err != nil {
			handler.Log.Error("Failed while decoding MinWinTid data", zap.Error(err), zap.String(vaktplanId, beredskapsvakt.ID.String()))
			continue
		}

		if !isTimesheetApproved(tiddataResult.Dager) {
			continue
		}

		var vaktplan models.Vaktplan
		err = json.Unmarshal(beredskapsvakt.Plan, &vaktplan)
		if err != nil {
			handler.Log.Error("Failed while unmarshaling beredskapsvaktperiode", zap.Error(err), zap.String(vaktplanId, vaktplan.ID.String()))
			continue
		}

		vacationAtTheSameTimeAsGuardDuty, err := isThereRegisteredVacationAtTheSameTimeAsGuardDuty(tiddataResult.Dager, vaktplan)
		if err != nil {
			handler.Log.Error("Failed while parsing date from MinWinTid", zap.Error(err), zap.String(vaktplanId, vaktplan.ID.String()))
			continue
		}
		if vacationAtTheSameTimeAsGuardDuty {
			handler.Log.Info("En bruker har hatt beredskapsvakt under ferien", zap.String(vaktplanId, vaktplan.ID.String()))
			continue
		}

		timesheet, errFields := formatTimesheet(tiddataResult.Dager)
		if len(errFields) != 0 {
			errFields = append(errFields, zap.String(vaktplanId, vaktplan.ID.String()))
			handler.Log.Error("Failed trying to format MinWinTid stemplinger", errFields...)
			continue
		}

		minWinTid := models.MinWinTid{
			ResourceID:   tiddataResult.VaktorNavId,
			ApproverID:   tiddataResult.VaktorLederNavId,
			ApproverName: tiddataResult.VaktorLederNavn,
			Satser: map[string]decimal.Decimal{
				"lørsøn":  decimal.NewFromInt(65),
				"0620":    decimal.NewFromInt(15),
				"2006":    decimal.NewFromInt(25),
				"utvidet": decimal.NewFromInt(25),
			},
			Timesheet: timesheet,
		}

		payroll, err := calculator.GuarddutySalary(vaktplan, minWinTid)
		if err != nil {
			handler.Log.Error("Failed while calculating salary", zap.Error(err), zap.String(vaktplanId, vaktplan.ID.String()))
			continue
		}

		err = postToVaktorPlan(handler, payroll, bearerToken)
		if err != nil {
			handler.Log.Error("Failed while posting to Vaktor Plan", zap.Error(err), zap.String(vaktplanId, vaktplan.ID.String()))
			continue
		}

		err = handler.Queries.DeletePlan(handler.Context, beredskapsvakt.ID)
		if err != nil {
			handler.Log.Error("Failed while deleting beredskapsvakt", zap.Error(err), zap.String(vaktplanId, vaktplan.ID.String()))
			continue
		}
	}

	return nil
}

func Run(handler endpoints.Handler) {
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
