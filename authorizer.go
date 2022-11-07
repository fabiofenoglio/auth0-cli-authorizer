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
	storeBuilder                storeBuilder
	storeRestoreMinDuration     time.Duration
	store                       store
	logger                      Logger
}

var _ Authorizer = &DefaultImpl{}

func (a *DefaultImpl) Authorize(ctx context.Context) (Authentication, error) {
	if ctx.Err() != nil {
		return Authentication{}, ctx.Err()
	}

	if a.store != nil {
		loaded, err := a.loadFromStore(ctx)
		if err != nil {
			a.logger.Warningf("failed to load authentication from store: %v", err)
		} else if loaded == nil {
			a.logger.Debug("no authentication available from store")
		} else {
			a.logger.Debug("loaded cached authentication from store")
			return *loaded, nil
		}
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

	authentication, err := a.buildAuthentication(ctx, tokenResponse.AccessToken, tokenResponse.IdToken, tokenResponse.RefreshToken)
	if err != nil {
		return Authentication{}, errors.Wrap(err, "error building authentication")
	}

	if a.store != nil {
		err = a.store.Save(authentication)
		if err != nil {
			a.logger.Errorf("error storing authentication in store: %v", err)
		}
	}

	return authentication, nil
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

	authentication, err := a.buildAuthentication(ctx, refreshTokenResponse.AccessToken, refreshTokenResponse.IdToken, newRefreshToken)
	if err != nil {
		return Authentication{}, errors.Wrap(err, "error building authentication")
	}

	err = a.store.Save(authentication)
	if err != nil {
		a.logger.Errorf("error saving the refreshed authentication: %v", err)
	}

	return authentication, nil
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

func (a *DefaultImpl) loadFromStore(ctx context.Context) (*Authentication, error) {
	if a.store == nil {
		return nil, errors.New("missing store implementation")
	}

	loaded, err := a.store.Load()
	if err != nil {
		return nil, err
	}
	if loaded == nil {
		return nil, nil
	}

	if loaded.Tokens.ExpiresAt.IsZero() {
		return nil, errors.New("restored tokens have an unknown expiration date")
	}

	expiresIn := time.Until(loaded.Tokens.ExpiresAt)

	if expiresIn >= a.storeRestoreMinDuration {
		a.logger.Debug("cache hit for store")
		return loaded, nil
	}

	a.logger.Debugf("cached authentication in store is expired (valid for %v, under threshold of %v)",
		expiresIn, a.storeRestoreMinDuration)

	if loaded.Tokens.RefreshToken == "" {
		return nil, errors.New("restored tokens could not be refreshed as no refresh token is available")
	}

	refreshed, err := a.Refresh(ctx, loaded.Tokens.RefreshToken)
	if err != nil {
		return nil, errors.Wrap(err, "error attempting to refresh the token")
	}

	a.logger.Debug("cached authentication was refreshed successfully")

	return &refreshed, nil
}
