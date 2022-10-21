package endpoints

import (
	"database/sql"
	"github.com/navikt/vaktor-lonn/pkg/auth"
	gensql "github.com/navikt/vaktor-lonn/pkg/sql/gen"
	"go.uber.org/zap"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type minWinTidConfig struct {
	Username       string
	Password       string
	Endpoint       string
	TickerInterval time.Duration
}

type Handler struct {
	BearerClient       auth.BearerClient
	DB                 *sql.DB
	Client             http.Client
	MinWinTidConfig    minWinTidConfig
	Queries            *gensql.Queries
	Log                *zap.Logger
	VaktorPlanEndpoint string
}

func NewHandler(logger *zap.Logger, dbString, vaktorPlanEndpoint,
	azureClientId, azureClientSecret, azureOpenIdTokenEndpoint,
	minWinTidUsername, minWinTidPassword, minWinTidEndpoint, minWinTidInterval string) (Handler, error) {
	db, err := sql.Open("pgx", dbString)
	if err != nil {
		return Handler{}, err
	}

	bearerClient, err := auth.New(azureClientId, azureClientSecret, azureOpenIdTokenEndpoint)
	if err != nil {
		return Handler{}, err
	}

	minWinTidTicketInterval, err := time.ParseDuration(minWinTidInterval)
	if err != nil {
		return Handler{}, err
	}

	handler := Handler{
		BearerClient: bearerClient,
		DB:           db,
		Client: http.Client{
			Timeout: 10 * time.Second,
		},
		MinWinTidConfig: minWinTidConfig{
			Username:       minWinTidUsername,
			Password:       minWinTidPassword,
			Endpoint:       minWinTidEndpoint,
			TickerInterval: minWinTidTicketInterval * time.Minute,
		},
		Queries:            gensql.New(db),
		Log:                logger,
		VaktorPlanEndpoint: vaktorPlanEndpoint,
	}

	return handler, nil
}
