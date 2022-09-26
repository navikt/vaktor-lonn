package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/navikt/vaktor-lonn/pkg/calculator"
	"github.com/navikt/vaktor-lonn/pkg/dummy"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"go.uber.org/zap"
	"net/http"
)

func (h Handler) Nudge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// TODO: Trigger sjekk av MinWinTid og hva vi har i databasen.
	beredskapsvakter, err := h.Queries.ListBeredskapsvakter(context.TODO())
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
		h.Log.Error("Error when trying list periods", zap.Error(err))
		return
	}

	token, err := h.BearerClient.GenerateBearerToken()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
		h.Log.Error("Error when unmarshaling plan", zap.Error(err))
		return
	}

	for _, beredskapsvakt := range beredskapsvakter {
		var vaktplan models.Vaktplan
		err := json.Unmarshal(beredskapsvakt.Plan, &vaktplan)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
			h.Log.Error("Error when unmarshaling plan", zap.Error(err))
			return
		}

		// TODO: Bytt med en korrekt implementasjon av kommunikasjon med MinWinTid
		minWinTid := dummy.GetMinWinTid(token, vaktplan)

		h.Log.Info("Calculating salary", zap.String("ident", beredskapsvakt.Ident))
		// TODO: Lage transaksjonsliste
		// TODO: Bytt ut med en go routine, da vi ikke skal svare med lønn her
		report, err := calculator.GuarddutySalary(vaktplan, minWinTid)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
			h.Log.Error("Error when calculating salary", zap.Error(err))
			return
		}
		err = json.NewEncoder(w).Encode(report)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
			h.Log.Error("Error when encoding json", zap.Error(err))
			return
		}
	}

	fmt.Fprint(w, "Nudged")
}
