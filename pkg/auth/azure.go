package auth

import (
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
	values.Add("scope", scope)

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

	request.Header.Set("Accept", "application/problem+json")

	return handleResponse(bc.Client.Do(request))
}
