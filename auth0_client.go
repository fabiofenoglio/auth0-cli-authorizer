package auth0cliauthorizer

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

const (
	defaultScopes      = "profile email openid"
	scopeOfflineAccess = "offline_access"
)

func (a *DefaultImpl) getDeviceCode(ctx context.Context) (deviceCodeResponseDTO, error) {
	a.logger.Debug("requesting a device code")

	scopes := defaultScopes
	if a.requireOfflineAccess {
		scopes += " " + scopeOfflineAccess
	}

	data := url.Values{}
	data.Set("client_id", a.clientID)
	data.Set("audience", a.audience)
	data.Set("scope", scopes)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		a.relativeURL("oauth/device/code"),
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return deviceCodeResponseDTO{}, errors.Wrap(err, "error creating HTTP request")
	}

	var deserialized deviceCodeResponseDTO
	if err = a.doWithClient(req, &deserialized); err != nil {
		return deviceCodeResponseDTO{}, err
	}

	if deserialized.DeviceCode == "" {
		return deviceCodeResponseDTO{}, errors.New("received an empty device code")
	}
	if deserialized.VerificationUriComplete == "" {
		return deviceCodeResponseDTO{}, errors.New("received an empty verification URI")
	}

	return deserialized, nil
}

func (a *DefaultImpl) getTokenFromDeviceCode(ctx context.Context, deviceCode string) (tokenResponseDTO, error) {

	data := url.Values{}
	data.Set("client_id", a.clientID)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	data.Set("device_code", deviceCode)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		a.relativeURL("oauth/token"),
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return tokenResponseDTO{}, errors.Wrap(err, "error creating HTTP request")
	}

	var deserialized tokenResponseDTO
	if err = a.doWithClient(req, &deserialized); err != nil {
		return tokenResponseDTO{}, err
	}

	if deserialized.AccessToken == "" {
		return tokenResponseDTO{}, errors.New("received an empty access token")
	}
	if deserialized.TokenType == "" {
		return tokenResponseDTO{}, errors.New("received an empty token type")
	}

	return deserialized, nil
}

func (a *DefaultImpl) getUserInfo(ctx context.Context, token string) (userInfoResponseDTO, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		a.relativeURL("userinfo"),
		nil,
	)
	if err != nil {
		return userInfoResponseDTO{}, errors.Wrap(err, "error creating HTTP request")
	}

	req.Header.Set("Authorization", "Bearer "+token)

	var deserialized userInfoResponseDTO
	if err = a.doWithClient(req, &deserialized); err != nil {
		return userInfoResponseDTO{}, err
	}

	if deserialized.Email == "" {
		return userInfoResponseDTO{}, errors.New("received an empty email")
	}

	return deserialized, nil
}

func (a *DefaultImpl) getTokenFromRefreshToken(ctx context.Context, refreshToken string) (refreshTokenResponseDTO, error) {

	data := url.Values{}
	data.Set("client_id", a.clientID)
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		a.relativeURL("oauth/token"),
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return refreshTokenResponseDTO{}, errors.Wrap(err, "error creating HTTP request")
	}

	var deserialized refreshTokenResponseDTO
	if err = a.doWithClient(req, &deserialized); err != nil {
		return refreshTokenResponseDTO{}, err
	}

	if deserialized.AccessToken == "" {
		return refreshTokenResponseDTO{}, errors.New("received an empty access token")
	}
	if deserialized.TokenType == "" {
		return refreshTokenResponseDTO{}, errors.New("received an empty token type")
	}

	return deserialized, nil
}
