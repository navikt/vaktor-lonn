package service

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/navikt/vaktor-lonn/pkg/auth"
	gensql "github.com/navikt/vaktor-lonn/pkg/sql/gen"
	"go.uber.org/zap"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type MinWinTidConfig struct {
	BearerClient   auth.BearerClient
	Endpoint       string
	TickerInterval time.Duration
}

type Handler struct {
	BearerClient       auth.BearerClient
	DB                 *sql.DB
	Client             http.Client
	Context            context.Context
	MinWinTidConfig    MinWinTidConfig
	VaktorPlanEndpoint string
	Queries            *gensql.Queries
	Log                *zap.Logger
}

func NewHandler(logger *zap.Logger, dbString,
	azureClientId, azureClientSecret, azureOpenIdTokenEndpoint, vaktorPlanEndpoint string, minWinTidConfig MinWinTidConfig) (Handler, error) {

	db, err := sql.Open("pgx", dbString)
	if err != nil {
		return Handler{}, err
	}

	handler := Handler{
		BearerClient: auth.New(azureClientId, azureClientSecret, azureOpenIdTokenEndpoint),
		DB:           db,
		Client: http.Client{
			Timeout: 10 * time.Second,
		},
		MinWinTidConfig:    minWinTidConfig,
		VaktorPlanEndpoint: vaktorPlanEndpoint,
		Queries:            gensql.New(db),
		Log:                logger,
	}

	return handler, nil
}
