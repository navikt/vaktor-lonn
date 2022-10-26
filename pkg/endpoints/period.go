package endpoints

import (
	"context"
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

	err = h.Queries.CreatePlan(context.TODO(), gensql.CreatePlanParams{
		ID:          plan.ID,
		Ident:       plan.Ident,
		Plan:        body,
		PeriodBegin: plan.Begin,
		PeriodEnd:   plan.End,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
		h.Log.Error("Error when trying to save period", zap.Error(err))
		return
	}

	_, err = fmt.Fprint(w, "Period save")
	if err != nil {
		h.Log.Error("Error when returning success", zap.Error(err))
		return
	}
}
