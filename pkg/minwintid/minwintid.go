package minwintid

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/navikt/vaktor-lonn/pkg/calculator"
	"github.com/navikt/vaktor-lonn/pkg/endpoints"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"net/http"
	"sort"
	"time"
)

const (
	DateTimeFormat  = "2006-01-02T15:04:05"
	fravarKodeFerie = 210
)

func getTimesheetFromMinWinTid(ident string, periodBegin time.Time, periodEnd time.Time, handler endpoints.Handler) (models.Response, error) {
	config := handler.MinWinTidConfig
	req, err := http.NewRequest(http.MethodGet, config.Endpoint, nil)
	if err != nil {
		return models.Response{}, err
	}

	req.SetBasicAuth(config.Username, config.Password)
	values := req.URL.Query()
	values.Add("ident", ident)
	values.Add("fra_dato", periodBegin.Format(calculator.VaktorDateFormat))
	values.Add("til_dato", periodEnd.Format(calculator.VaktorDateFormat))

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
			return models.Response{}, err
		}
	}

	if r.StatusCode != http.StatusOK {
		return models.Response{}, fmt.Errorf("minWinTid returned http(%v)", r.StatusCode)
	}

	var response models.Response
	err = json.NewDecoder(r.Body).Decode(&response)
	if err != nil {
		return models.Response{}, err
	}

	return response, nil
}

func isTimesheetApproved(days []models.Dag) bool {
	for _, day := range days {
		if day.Godkjent == 0 {
			return false
		}
	}

	return true
}

func isThereRegisteredVacationAtTheSameTimeAsGuardDuty(days []models.Dag, vaktplan models.Vaktplan) (bool, error) {
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

func formatTimesheet(days []models.Dag) (map[string]models.TimeSheet, error) {
	timesheet := make(map[string]models.TimeSheet)

	for _, day := range days {
		stemplingDate, err := time.Parse(DateTimeFormat, day.Dato)
		if err != nil {
			return nil, err
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

		stemplinger := day.Stemplinger
		if len(stemplinger) > 0 {
			sort.Slice(stemplinger, func(i, j int) bool {
				return stemplinger[i].StemplingTid < stemplinger[j].StemplingTid
			})

			for len(stemplinger) != 0 {
				// TODO: Hva betyr B1 og B2?
				innStempling := stemplinger[0]
				stemplinger = stemplinger[1:]

				utStempling := stemplinger[0]
				stemplinger = stemplinger[1:]

				if innStempling.Retning == "Inn" && innStempling.Type == "B1" &&
					utStempling.Retning == "Ut" && utStempling.Type == "B2" {
					// A correct stempling!
					innStemplingDate, err := time.Parse(DateTimeFormat, innStempling.StemplingTid)
					if err != nil {
						return nil, err
					}

					utStemplingDate, err := time.Parse(DateTimeFormat, utStempling.StemplingTid)
					if err != nil {
						return nil, err
					}

					ts.Clockings = append(ts.Clockings, models.Clocking{
						In:  innStemplingDate,
						Out: utStemplingDate,
					})
					continue
				}

				if innStempling.Retning == "B4" && innStempling.Type == "B4" && // kan dette ha noe med dette er siste dag i ferien?
					utStempling.Retning == "Ut" && utStempling.Type == "B2" {
					// TODO: Kan dette være at man avslutter en lengre stempling (som både kan være ferie og avspasering)?
					// Bruker har hatt ferie denne stemplingen
					continue
				}
				if innStempling.Retning == "Inn" && innStempling.Type == "B1" && // kan dette ha noe med at dette er første dag med avspasering?
					utStempling.Retning == "B5" && utStempling.Type == "B5" {
					// TODO: Denne slår ut på dager man har ferie også!
					// Bruker har avspasert denne stemplingen
					continue
				}

				return nil, fmt.Errorf("did not get expected direction or type, expected 'Inn' and 'B1', got direction=%v, type=%v", innStempling.Retning, innStempling.Type)
			}
		}

		timesheet[simpleStemplingDate] = ts
	}
	return timesheet, nil
}

func helper(handler endpoints.Handler) error {
	beredskapsvakter, err := handler.Queries.ListBeredskapsvakter(context.TODO())
	if err != nil {
		handler.Log.Error("Failed while listing beredskapsvakter", zap.Error(err))
		return err
	}

	for _, beredskapsvakt := range beredskapsvakter {
		response, err := getTimesheetFromMinWinTid(beredskapsvakt.Ident, beredskapsvakt.PeriodBegin, beredskapsvakt.PeriodEnd, handler)
		if err != nil {
			handler.Log.Error("Failed while retrieving data from MinWinTid", zap.Error(err))
			continue
		}

		rows := response.VaktorVaktorTiddataResponse.VaktorVaktorTiddataResult.VaktorRow
		if len(rows) == 1 {
			row := rows[0]
			var dager []models.Dag
			err := json.Unmarshal([]byte(row.VaktorDager), &dager)
			if err != nil {
				handler.Log.Error("Failed while unmarshaling VaktorDager", zap.Error(err), zap.String("MinWinTid", row.VaktorResourceId))
				continue
			}
			if !isTimesheetApproved(dager) {
				continue
			}

			var vaktplan models.Vaktplan
			err = json.Unmarshal(beredskapsvakt.Plan, &vaktplan)
			if err != nil {
				handler.Log.Error("Failed while unmarshaling beredskapsvaktperiode", zap.Error(err), zap.String("vaktplanId", vaktplan.ID.String()))
				continue
			}

			vacationAtTheSameTimeAsGuardDuty, err := isThereRegisteredVacationAtTheSameTimeAsGuardDuty(dager, vaktplan)
			if err != nil {
				handler.Log.Error("Failed while parsing date from MinWinTid", zap.Error(err), zap.String("vaktplanId", vaktplan.ID.String()))
				continue
			}
			if vacationAtTheSameTimeAsGuardDuty {
				handler.Log.Info("En bruker har hatt beredskapsvakt under ferien", zap.String("vaktplanId", vaktplan.ID.String()))
				continue
			}

			timesheet, err := formatTimesheet(dager)
			if err != nil {
				handler.Log.Error("Failed trying to format MinWinTid stemplinger", zap.Error(err), zap.String("vaktplanId", vaktplan.ID.String()))
				continue
			}

			minWinTid := models.MinWinTid{
				Ident:      row.VaktorNavId,
				ResourceID: row.VaktorResourceId,
				Satser: map[string]decimal.Decimal{
					"lørsøn":  decimal.NewFromInt(55),
					"0620":    decimal.NewFromInt(10),
					"2006":    decimal.NewFromInt(20),
					"utvidet": decimal.NewFromInt(15),
				}, // TODO: Dette skal egentlig komme fra MinWinTid
				Timesheet: timesheet,
			}

			err = calculator.GuarddutySalary(vaktplan, minWinTid)
			if err != nil {
				handler.Log.Error("Failed while calculating salary", zap.Error(err), zap.String("vaktplanId", vaktplan.ID.String()))
				continue
			}

			// TODO: Kall til Vaktor Plan
			// curl -X POST -h "Authorization: bearer $TOKEN" -d {"id": uuid, "artskoder": [{"2600B": 1234.00, "2603B": 4132.00}]} vaktor/

			err = handler.Queries.DeletePlan(context.TODO(), beredskapsvakt.ID)
			if err != nil {
				handler.Log.Error("Failed while deleting beredskapsvakt", zap.Error(err), zap.String("vaktplanId", vaktplan.ID.String()))
				// TODO: Dette er litt krise, for det betyr at kjøringen fortsatt gjøres :thinking:
				// Kan prøve å oppdatere feltet, og ha en slette-kolonne
			}
		}
	}

	return nil
}
