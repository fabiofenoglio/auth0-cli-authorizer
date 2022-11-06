package auth0cliauthorizer

import (
	"context"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/browser"
	"github.com/pkg/errors"
)

type Authorizer interface {
	Authorize(ctx context.Context) (Authentication, error)
	Refresh(ctx context.Context, refreshToken string) (Authentication, error)
}

type DefaultImpl struct {
	domain                      *url.URL
	clientID                    string
	audience                    string
	prefillDeviceCode           bool
	requireOfflineAccess        bool
	autoOpenBrowser             bool
	httpClientCustomizer        HTTPClientCustomizer
	deviceConfirmPromptCallback DeviceConfirmPromptCallback
	logger                      Logger
}

var _ Authorizer = &DefaultImpl{}

func (a *DefaultImpl) Authorize(ctx context.Context) (Authentication, error) {
	if ctx.Err() != nil {
		return Authentication{}, ctx.Err()
	}

	deviceCodeResponse, err := a.getDeviceCode(ctx)
	if err != nil {
		return Authentication{}, errors.Wrap(err, "error fetching the device code")
	}

	if a.autoOpenBrowser {
		toOpen := deviceCodeResponse.VerificationUriComplete
		if !a.prefillDeviceCode {
			toOpen = deviceCodeResponse.VerificationUri
		}
		err = browser.OpenURL(toOpen)
		if err != nil {
			return Authentication{}, errors.Wrap(err, "error opening browser window")
		}
	}

	if a.deviceConfirmPromptCallback != nil {
		err = a.deviceConfirmPromptCallback(DeviceConfirmPrompt{
			DeviceCode:              deviceCodeResponse.DeviceCode,
			VerificationUri:         deviceCodeResponse.VerificationUri,
			VerificationUriComplete: deviceCodeResponse.VerificationUriComplete,
			ExpiresIn:               deviceCodeResponse.ExpiresIn,
		})
		if err != nil {
			return Authentication{}, err
		}
	}

	pollingInterval := time.Second*time.Duration(deviceCodeResponse.Interval) + time.Millisecond*500

	tokenResponse, err := a.pollForToken(ctx, deviceCodeResponse.DeviceCode, pollingInterval)
	if err != nil {
		return Authentication{}, errors.Wrap(err, "error waiting for authorization")
	}

	user, err := a.buildAuthentication(ctx, tokenResponse.AccessToken, tokenResponse.IdToken, tokenResponse.RefreshToken)
	if err != nil {
		return Authentication{}, errors.Wrap(err, "error building user")
	}

	return user, nil
}

func (a *DefaultImpl) Refresh(ctx context.Context, refreshToken string) (Authentication, error) {
	if ctx.Err() != nil {
		return Authentication{}, ctx.Err()
	}

	if refreshToken == "" {
		return Authentication{}, errors.New("missing refresh token")
	}

	refreshTokenResponse, err := a.getTokenFromRefreshToken(ctx, refreshToken)
	if err != nil {
		return Authentication{}, errors.Wrap(err, "error refreshing token")
	}

	newRefreshToken := refreshToken
	if refreshTokenResponse.RefreshToken != "" {
		newRefreshToken = refreshTokenResponse.RefreshToken
	}

	auth, err := a.buildAuthentication(ctx, refreshTokenResponse.AccessToken, refreshTokenResponse.IdToken, newRefreshToken)
	if err != nil {
		return Authentication{}, errors.Wrap(err, "error building user")
	}

	return auth, nil
}

func (a *DefaultImpl) buildAuthentication(ctx context.Context, accessToken, idToken, refreshToken string) (Authentication, error) {

	var accessTokenContent accessTokenContentDTO
	_, _, err := jwt.NewParser().ParseUnverified(accessToken, &accessTokenContent)
	if err != nil {
		return Authentication{}, errors.Wrap(err, "error decoding access token")
	}

	var userInfo commonUserInfoDTO

	if idToken != "" {
		var idTokenContent idTokenContentDTO
		_, _, err = jwt.NewParser().ParseUnverified(idToken, &idTokenContent)
		if err != nil {
			return Authentication{}, errors.Wrap(err, "error decoding identity token")
		}
		userInfo = idTokenContent.commonUserInfoDTO

	} else {
		userInfoResponse, err := a.getUserInfo(ctx, accessToken)
		if err != nil {
			return Authentication{}, errors.Wrap(err, "error fetching user profile")
		}
		userInfo = userInfoResponse.commonUserInfoDTO
	}

	var name string
	switch {
	case userInfo.Nickname != "":
		name = userInfo.Nickname
	case userInfo.Name != "":
		name = userInfo.Name
	case userInfo.Email != "":
		name = userInfo.Email
	default:
		name = accessTokenContent.Subject
	}

	expiresAt := time.Time{}
	if accessTokenContent.ExpiresAt != nil {
		expiresAt = (*accessTokenContent.ExpiresAt).Time
	}

	return Authentication{
		User: User{
			Name:        name,
			Email:       userInfo.Email,
			Permissions: accessTokenContent.Permissions,
		},
		Tokens: Tokens{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresAt:    expiresAt,
		},
	}, nil
}

func (a *DefaultImpl) pollForToken(ctx context.Context, deviceCode string, pollingInterval time.Duration) (tokenResponseDTO, error) {
	if pollingInterval < time.Second {
		pollingInterval = time.Second
	}

	a.logger.Debugf("will poll for an authorization token every %d ms", pollingInterval.Milliseconds())

	var token tokenResponseDTO
	var err error

	for {
		select {
		case <-time.After(pollingInterval):
		case <-ctx.Done():
			a.logger.Debug("context canceled, stopping polling")
			return tokenResponseDTO{}, ctx.Err()
		}

		token, err = a.getTokenFromDeviceCode(ctx, deviceCode)
		if err == nil {
			break
		}

		switch err {
		case errAuthorizationPending:
			a.logger.Debugf("still waiting ... (%v)", err)
		case errSlowDown:
			a.logger.Debugf("have to slow down (%v)", err)
			pollingInterval += time.Second
			a.logger.Debugf("polling every %d ms", pollingInterval.Milliseconds())
		default:
			return tokenResponseDTO{}, errors.Wrap(err, "error polling for verification status")
		}
	}

	return token, nil
}
