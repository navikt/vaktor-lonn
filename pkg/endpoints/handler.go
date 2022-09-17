package endpoints

import (
	"database/sql"
	gensql "github.com/navikt/vaktor-lonn/pkg/sql/gen"
	"net/http"
	"time"
)

type Handler struct {
	DB      *sql.DB
	Client  http.Client
	Queries *gensql.Queries
}

func NewHandler(dbString string) (Handler, error) {
	db, err := sql.Open("postgres", dbString)
	if err != nil {
		return Handler{}, err
	}

	handler := Handler{
		DB: db,
		Client: http.Client{
			Timeout: 10 * time.Second,
		},
		Queries: gensql.New(db),
	}

	return handler, nil
}
