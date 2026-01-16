package service

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/navikt/vaktor-lonn/pkg/auth"
	gensql "github.com/navikt/vaktor-lonn/pkg/sql/gen"
	"go.uber.org/zap"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type MinWinTidConfig struct {
	BearerClient   auth.BasicAuthClient
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
	azureClientId, azureClientSecret, azureOpenIdTokenEndpoint, vaktorPlanEndpoint string, minWinTidConfig MinWinTidConfig,
) (Handler, error) {
	db, err := openDB(logger, dbString)
	if err != nil {
		return Handler{}, err
	}

	handler := Handler{
		BearerClient: auth.New(azureClientId, azureClientSecret, azureOpenIdTokenEndpoint, "https://graph.microsoft.com/.default"),
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

func openDB(logger *zap.Logger, dbString string) (*sql.DB, error) {
	var db *sql.DB
	var err error
	maxRetries := 5
	backoff := time.Second

	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("pgx", dbString)
		if err == nil {
			break
		}

		logger.Info("Failed to open database connection, retrying...", zap.Error(err))
		time.Sleep(backoff)
		backoff *= 2
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open database connection after retries: %s", err)
	}

	return db, nil
}
