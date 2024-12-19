package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

type BearerClient struct {
	Client   *http.Client
	Endpoint string
	Body     string
}

func New(clientId, clientSecret, authEndpoint, scope string) BearerClient {
	values := url.Values{}
	values.Add("client_id", clientId)
	values.Add("client_secret", clientSecret)
	values.Add("grant_type", "client_credentials")
	if scope != "" {
		values.Add("scope", scope)
	}

	return BearerClient{
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
		Endpoint: authEndpoint,
		Body:     values.Encode(),
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

	defer resp.Body.Close()

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
