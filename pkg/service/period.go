package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/navikt/vaktor-lonn/pkg/calculator"
	"github.com/navikt/vaktor-lonn/pkg/models"
	gensql "github.com/navikt/vaktor-lonn/pkg/sql/gen"
	"go.uber.org/zap"
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

	var plan models.Vaktplan
	err = json.Unmarshal(body, &plan)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
		h.Log.Error("Error when decoding body from request", zap.Error(err))
		return
	}

	if len(plan.Schedule) == 0 {
		h.Log.Error("No schedule found in request")
		_, err := fmt.Fprint(w, "{\"message\":\"No schedule found in request\"}\n")
		if err != nil {
			h.Log.Error("Error when returning error", zap.Error(err))
		}

		return
	}

	var dates []string
	for key := range plan.Schedule {
		dates = append(dates, key)
	}
	sort.Strings(dates)

	periodBegin, err := time.Parse(calculator.VaktorDateFormat, dates[0])
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
		h.Log.Error("Error when parsing period begin", zap.Error(err))
		return
	}

	periodEnd, err := time.Parse(calculator.VaktorDateFormat, dates[len(dates)-1])
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
		h.Log.Error("Error when parsing period end", zap.Error(err))
		return
	}

	if err := h.Queries.CreatePlan(h.Context, gensql.CreatePlanParams{
		ID:          plan.ID,
		Ident:       plan.Ident,
		Plan:        body,
		PeriodBegin: periodBegin,
		PeriodEnd:   periodEnd,
	}); err != nil {
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

	go handleTransaction(h, beredskapsvakt)
}
