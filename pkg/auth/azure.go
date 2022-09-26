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
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	ExtExpiresIn int    `json:"ext_expires_in"`
	AccessToken  string `json:"access_token"`
}

type BearerClient struct {
	client  *http.Client
	request *http.Request
}

func New(clientId, clientSecret, authEndpoint string) (bc BearerClient, err error) {
	bc.client = &http.Client{
		Timeout: 10 * time.Second,
	}

	body := url.Values(map[string][]string{
		"client_id":     {clientId},
		"client_secret": {clientSecret},
		"grant_type":    {"client_credentials"},
		"scope":         {"https://graph.microsoft.com/.default"}})

	bc.request, err = http.NewRequest(
		http.MethodPost,
		authEndpoint,
		strings.NewReader(body.Encode()))
	if err != nil {
		return
	}

	bc.request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return
}

func (bc BearerClient) GenerateBearerToken() (string, error) {
	resp, err := bc.client.Do(bc.request)
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
