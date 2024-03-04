package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
)

type TokenResponse struct {
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	ExtExpiresIn int    `json:"ext_expires_in"`
	AccessToken  string `json:"access_token"`
}

type BearerClient struct {
	Client   *http.Client
	Endpoint string
	Body     string
	Log      *zap.Logger
}

func New(logger *zap.Logger, clientId, clientSecret, authEndpoint string) BearerClient {
	values := url.Values{}
	values.Add("client_id", clientId)
	values.Add("client_secret", clientSecret)
	values.Add("grant_type", "client_credentials")
	values.Add("scope", "https://graph.microsoft.com/.default")

	return BearerClient{
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
		Endpoint: authEndpoint,
		Body:     values.Encode(),
		Log:      logger,
	}
}

func (bc BearerClient) GenerateBearerToken() (string, error) {
	request, err := http.NewRequest(
		http.MethodPost,
		bc.Endpoint,
		strings.NewReader(bc.Body))
	if err != nil {
		return "", err
	}

	resp, err := bc.Client.Do(request)
	if err != nil {
		return "", err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			bc.Log.Error("Failed while closing body", zap.Error(err))
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("httpStatus: %v", resp.StatusCode)
	}

	var tokenRespons TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenRespons)
	if err != nil {
		return "", err
	}

	return tokenRespons.AccessToken, nil
}
