package auth0cliauthorizer

import (
	"fmt"
)

var (
	errAuthorizationPending = &managedHTTPError{
		ErrorCode:        "authorization_pending",
		ErrorDescription: "still waiting for user authorization",
	}
	errSlowDown = &managedHTTPError{
		ErrorCode:        "slow_down",
		ErrorDescription: "too many requests",
	}
	errExpiredToken = &managedHTTPError{
		ErrorCode:        "expired_token",
		ErrorDescription: "token is expired",
	}
	errInvalidGrant = &managedHTTPError{
		ErrorCode:        "invalid_grant",
		ErrorDescription: "the grant request is invalid",
	}
	errAccessDenied = &managedHTTPError{
		ErrorCode:        "access_denied",
		ErrorDescription: "access denied",
	}
)

type managedHTTPError struct {
	ErrorCode        string `json:"error_code"`
	ErrorDescription string `json:"error_description"`
}

func (e *managedHTTPError) Error() string {
	return fmt.Sprintf("Error: %s (%s)", e.ErrorCode, e.ErrorDescription)
}
