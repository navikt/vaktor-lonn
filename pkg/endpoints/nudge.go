package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/navikt/vaktor-lonn/pkg/calculator"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/rs/zerolog/log"
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
		log.Err(err)
		return
	}

	for _, beredskapsvakt := range beredskapsvakter {
		var vaktplan models.Vaktplan
		err := json.Unmarshal(beredskapsvakt.Plan, &vaktplan)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
			log.Err(err)
			return
		}

		log.Printf("Calculating salary for %s", beredskapsvakt.Ident)
		report, err := calculator.GuarddutySalary(vaktplan)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
			log.Err(err)
			return
		}
		err = json.NewEncoder(w).Encode(report)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
			log.Err(err)
			return
		}
	}
}
