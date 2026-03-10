package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
)

type KeycloakIdentityProvider struct {
	baseURL    string
	realm      string
	clientID   string
	httpClient *http.Client
}

type keycloakTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func NewKeycloakIdentityProvider(
	baseURL string,
	realm string,
	clientID string,
	httpClient *http.Client,
) *KeycloakIdentityProvider {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	return &KeycloakIdentityProvider{
		baseURL:    strings.TrimRight(baseURL, "/"),
		realm:      realm,
		clientID:   clientID,
		httpClient: httpClient,
	}
}

func (k *KeycloakIdentityProvider) Login(
	ctx context.Context,
	username string,
	password string,
) (*ports.IdentityLoginResult, error) {

	form := url.Values{}
	form.Set("client_id", k.clientID)
	form.Set("grant_type", "password")
	form.Set("username", username)
	form.Set("password", password)

	endpoint := k.baseURL + "/realms/" + k.realm + "/protocol/openid-connect/token"

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, middleware.NewInternalError(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, middleware.NewInternalError(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, middleware.NewInternalError(err)
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusBadRequest {
		return nil, middleware.NewInvalidCredentialsError()
	}

	if resp.StatusCode != http.StatusOK {
		return nil, middleware.NewUnauthorizedError("keycloak authentication failed")
	}

	var tr keycloakTokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, middleware.NewInternalError(err)
	}

	sub, err := extractSubFromJWT(tr.AccessToken)
	if err != nil {
		return nil, middleware.NewInternalError(err)
	}

	return &ports.IdentityLoginResult{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		Subject:      sub,
		ExpiresIn:    tr.ExpiresIn,
	}, nil
}

func extractSubFromJWT(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", errors.New("invalid jwt format")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}

	var payload struct {
		Sub string `json:"sub"`
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return "", err
	}

	if payload.Sub == "" {
		return "", errors.New("sub claim missing")
	}

	return payload.Sub, nil
}