package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type BasicAuthClient struct {
	Client       *http.Client
	Endpoint     string
	ClientID     string
	ClientSecret string
	Body         string
}

func NewWithBasicAuth(clientId, clientSecret, authEndpoint string) BasicAuthClient {
	values := url.Values{}
	values.Add("grant_type", "client_credentials")

	return BasicAuthClient{
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
		Endpoint:     authEndpoint,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Body:         values.Encode(),
	}
}

func (bc BasicAuthClient) GenerateBearerToken() (string, error) {
	request, err := http.NewRequest(
		http.MethodPost,
		bc.Endpoint,
		strings.NewReader(bc.Body))
	if err != nil {
		return "", err
	}

	request.Header.Set("Accept", "application/problem+json")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(bc.ClientID, bc.ClientSecret)

	return handleResponse(bc.Client.Do(request))
}

func handleResponse(resp *http.Response, err error) (string, error) {
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("httpStatus: %v (reading body failed)", resp.StatusCode)
		}

		return "", fmt.Errorf("httpStatus: %v, %s", resp.StatusCode, string(bodyBytes))
	}

	var tokenRespons TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenRespons)
	if err != nil {
		return "", err
	}

	return tokenRespons.AccessToken, nil
}
