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
	baseURL           string
	realm             string
	clientID          string
	adminClientID     string
	adminClientSecret string
	httpClient        *http.Client
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
	adminClientID string,
	adminClientSecret string,
	httpClient *http.Client,
) *KeycloakIdentityProvider {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	return &KeycloakIdentityProvider{
		baseURL:           strings.TrimRight(baseURL, "/"),
		realm:             realm,
		clientID:          clientID,
		adminClientID:     adminClientID,
		adminClientSecret: adminClientSecret,
		httpClient:        httpClient,
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

func (k *KeycloakIdentityProvider) RegisterUser(
	ctx context.Context,
	email,
	password,
	firstName,
	lastName string,
) (string, error) {
	adminToken, err := k.getAdminToken(ctx)
	if err != nil {
		return "", err
	}

	userPayload := map[string]interface{}{
		"username":      email,
		"email":         email,
		"firstName":     firstName,
		"lastName":      lastName,
		"enabled":       true,
		"emailVerified": true,
		"credentials": []map[string]interface{}{
			{
				"type":      "password",
				"value":     password,
				"temporary": false,
			},
		},
	}

	body, _ := json.Marshal(userPayload)
	endpoint := k.baseURL + "/admin/realms/" + k.realm + "/users"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return "", middleware.NewInternalError(err)
	}

	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return "", middleware.NewInternalError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return "", middleware.NewEmailAlreadyUsedError(email)
	}

	if resp.StatusCode != http.StatusCreated {
		return "", middleware.NewInternalError(errors.New("failed to create keycloak user"))
	}

	// Extract ID from Location header: .../users/{id}
	location := resp.Header.Get("Location")
	if location == "" {
		// Fallback lookup if Location is missing
		return k.lookupUserIDByEmail(ctx, adminToken, email)
	}

	parts := strings.Split(location, "/")
	return parts[len(parts)-1], nil
}

func (k *KeycloakIdentityProvider) DeleteUser(ctx context.Context, keycloakSub string) error {
	adminToken, err := k.getAdminToken(ctx)
	if err != nil {
		return err
	}

	endpoint := k.baseURL + "/admin/realms/" + k.realm + "/users/" + keycloakSub
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return middleware.NewInternalError(err)
	}

	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return middleware.NewInternalError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return middleware.NewInternalError(errors.New("failed to delete keycloak user"))
	}

	return nil
}

func (k *KeycloakIdentityProvider) getAdminToken(ctx context.Context) (string, error) {
	form := url.Values{}
	form.Set("client_id", k.adminClientID)
	form.Set("client_secret", k.adminClientSecret)
	form.Set("grant_type", "client_credentials")

	endpoint := k.baseURL + "/realms/" + k.realm + "/protocol/openid-connect/token"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", middleware.NewInternalError(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return "", middleware.NewInternalError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", middleware.NewUnauthorizedError("failed to get keycloak admin token")
	}

	var tr keycloakTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", middleware.NewInternalError(err)
	}

	return tr.AccessToken, nil
}

func (k *KeycloakIdentityProvider) lookupUserIDByEmail(ctx context.Context, token, email string) (string, error) {
	endpoint := k.baseURL + "/admin/realms/" + k.realm + "/users?email=" + url.QueryEscape(email) + "&exact=true"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", middleware.NewInternalError(err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return "", middleware.NewInternalError(err)
	}
	defer resp.Body.Close()

	var users []struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return "", middleware.NewInternalError(err)
	}

	if len(users) == 0 {
		return "", middleware.NewInternalError(errors.New("user not found after creation"))
	}

	return users[0].ID, nil
}

func (k *KeycloakIdentityProvider) Refresh(ctx context.Context, refreshToken string) (*ports.IdentityLoginResult, error) {
	form := url.Values{}
	form.Set("client_id", k.clientID)
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)

	endpoint := k.baseURL + "/realms/" + k.realm + "/protocol/openid-connect/token"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, middleware.NewInternalError(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, middleware.NewInternalError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, middleware.NewUnauthorizedError("invalid or expired refresh token")
	}

	var tr keycloakTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
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

func (k *KeycloakIdentityProvider) Logout(ctx context.Context, refreshToken string) error {
	form := url.Values{}
	form.Set("client_id", k.clientID)
	form.Set("refresh_token", refreshToken)

	endpoint := k.baseURL + "/realms/" + k.realm + "/protocol/openid-connect/logout"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return middleware.NewInternalError(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return middleware.NewInternalError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return middleware.NewUnauthorizedError("failed to logout")
	}

	return nil
}