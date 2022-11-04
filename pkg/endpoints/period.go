package endpoints

import (
	"encoding/json"
	"fmt"
	"github.com/navikt/vaktor-lonn/pkg/models"
	gensql "github.com/navikt/vaktor-lonn/pkg/sql/gen"
	"go.uber.org/zap"
	"io"
	"net/http"
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

	oldPlan, err := h.Queries.GetPlan(h.Context, plan.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
		h.Log.Error("Error when trying to check for earlier period", zap.Error(err))
		return
	}

	if oldPlan.Ident == "" {
		err = h.Queries.CreatePlan(h.Context, gensql.CreatePlanParams{
			ID:          plan.ID,
			Ident:       plan.Ident,
			Plan:        body,
			PeriodBegin: plan.Begin,
			PeriodEnd:   plan.End,
		})
	} else {
		err = h.Queries.UpdatePlan(h.Context, gensql.UpdatePlanParams{
			ID:          plan.ID,
			Ident:       plan.Ident,
			Plan:        body,
			PeriodBegin: plan.Begin,
			PeriodEnd:   plan.End,
		})
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
		h.Log.Error("Error when trying to save period", zap.Error(err))
		return
	}

	h.Log.Info(fmt.Sprintf("Received period %v", plan.ID))
	_, err = fmt.Fprint(w, "{\"message\":\"Period saved\"}\n")
	if err != nil {
		h.Log.Error("Error when returning success", zap.Error(err))
		return
	}
}
