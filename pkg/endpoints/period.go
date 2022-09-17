package endpoints

import (
	"context"
	"fmt"
	"github.com/navikt/vaktor-lonn/pkg/models"
	gensql "github.com/navikt/vaktor-lonn/pkg/sql/gen"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

func (h Handler) Period(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
		log.Err(err)
		return
	}

	// TODO: Skal vi validere input?
	var plan models.Vaktplan
	err = h.Queries.CreatePlan(context.TODO(), gensql.CreatePlanParams{
		ID:    plan.ID,
		Ident: plan.Ident,
		Plan:  body,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
		log.Err(err)
		return
	}
}
