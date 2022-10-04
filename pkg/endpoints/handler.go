package endpoints

import (
	"database/sql"
	gensql "github.com/navikt/vaktor-lonn/pkg/sql/gen"
	"go.uber.org/zap"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type minWinTidConfig struct {
	Username string
	Password string
	Endpoint string
}

type Handler struct {
	DB              *sql.DB
	Client          http.Client
	MinWinTidConfig minWinTidConfig
	Queries         *gensql.Queries
	Log             *zap.Logger
}

func NewHandler(logger *zap.Logger, dbString, minWinTidUsername, minWinTidPassword, minWinTidEndpoint string) (Handler, error) {
	db, err := sql.Open("pgx", dbString)
	if err != nil {
		return Handler{}, err
	}

	handler := Handler{
		DB: db,
		Client: http.Client{
			Timeout: 10 * time.Second,
		},
		MinWinTidConfig: minWinTidConfig{
			Username: minWinTidUsername,
			Password: minWinTidPassword,
			Endpoint: minWinTidEndpoint,
		},
		Queries: gensql.New(db),
		Log:     logger,
	}

	return handler, nil
}
