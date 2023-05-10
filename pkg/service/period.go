package service

import (
	"encoding/json"
	"fmt"
	"github.com/navikt/vaktor-lonn/pkg/calculator"
	"github.com/navikt/vaktor-lonn/pkg/models"
	gensql "github.com/navikt/vaktor-lonn/pkg/sql/gen"
	"go.uber.org/zap"
	"io"
	"net/http"
	"sort"
	"time"
)

func (h Handler) Period(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			h.Log.Error("Error while closing body", zap.Error(err))
		}
	}(r.Body)

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
		h.Log.Error("Error when reading body from request", zap.Error(err))
		return
	}

	// TODO: Hvordan kan vi validere input?
	var plan models.Vaktplan
	err = json.Unmarshal(body, &plan)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
		h.Log.Error("Error when decoding body from request", zap.Error(err))
		return
	}

	var dates []string
	for key, _ := range plan.Schedule {
		dates = append(dates, key)
	}
	sort.Strings(dates)
	periodBegin, err := time.Parse(calculator.VaktorDateFormat, dates[0])
	periodEnd, err := time.Parse(calculator.VaktorDateFormat, dates[len(dates)-1])

	err = h.Queries.CreatePlan(h.Context, gensql.CreatePlanParams{
		ID:          plan.ID,
		Ident:       plan.Ident,
		Plan:        body,
		PeriodBegin: periodBegin,
		PeriodEnd:   periodEnd,
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
		h.Log.Error("Error when trying to save period", zap.Error(err), zap.String(vaktplanId, plan.ID.String()))
		return
	}

	h.Log.Info(fmt.Sprintf("Received period %v", plan.ID))
	_, err = fmt.Fprint(w, "{\"message\":\"Period saved\"}\n")
	if err != nil {
		h.Log.Error("Error when returning success", zap.Error(err), zap.String(vaktplanId, plan.ID.String()))
		return
	}

	beredskapsvakt := gensql.Beredskapsvakt{
		ID:          plan.ID,
		Ident:       plan.Ident,
		Plan:        body,
		PeriodBegin: periodBegin,
		PeriodEnd:   periodEnd,
	}
	go triggerHandleOfTransaction(h, beredskapsvakt)
}

func triggerHandleOfTransaction(handler Handler, beredskapsvakt gensql.Beredskapsvakt) {
	bearerToken, err := handler.BearerClient.GenerateBearerToken()
	if err != nil {
		handler.Log.Error("Problem generating bearer token", zap.Error(err))
	}
	handleTransaction(handler, beredskapsvakt, bearerToken)
}
