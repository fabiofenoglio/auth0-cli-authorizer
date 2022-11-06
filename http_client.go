package auth0cliauthorizer

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const headerContentType = "content-type"
const contentTypeForPOST = "application/x-www-form-urlencoded"

func (a *DefaultImpl) getHTTPClient() *http.Client {
	var netTransport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	var netClient = &http.Client{
		Timeout:   time.Second * 15,
		Transport: netTransport,
	}

	if a.httpClientCustomizer != nil {
		a.httpClientCustomizer(netClient)
	}

	return netClient
}

func (a *DefaultImpl) relativeURL(relative string) string {
	url, err := url.Parse(a.domain.String())
	if err != nil {
		panic(err)
	}
	url.Path = path.Join(url.Path, relative)
	return url.String()
}

func (a *DefaultImpl) onRequest(req *http.Request) {
	a.logger.Debugf("sending HTTP %s request to %s", req.Method, req.URL.String())
}

func (a *DefaultImpl) onResponse(req *http.Request, res *http.Response) {
	a.logger.Debugf("received status %d - %s from HTTP %s request to %s", res.StatusCode, res.Status, req.Method, req.URL.String())
}

func (a *DefaultImpl) doWithClient(req *http.Request, target interface{}) error {
	client := a.getHTTPClient()

	if req.Header.Get(headerContentType) == "" && (req.Method == http.MethodPost || req.Method == http.MethodPut) {
		req.Header.Add(headerContentType, contentTypeForPOST)
	}

	a.onRequest(req)

	res, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "error sending HTTP request")
	}
	defer res.Body.Close()

	a.onResponse(req, res)

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.Wrap(err, "error reading response")
	}

	if res.StatusCode >= 300 {
		return parseError(res, body)
	}

	if err = json.Unmarshal(body, target); err != nil {
		return errors.Wrap(err, "error decoding response")
	}

	return nil
}

func parseError(res *http.Response, body []byte) error {
	generic := errors.Errorf("error %d (%s)", res.StatusCode, res.Status)

	if !strings.Contains(strings.ToLower(res.Header.Get(headerContentType)), "/json") {
		return generic
	}

	var deserializedError errorResponseDTO
	if err := json.Unmarshal(body, &deserializedError); err != nil {
		return generic
	}

	if deserializedError.ErrorCode == "" {
		return generic
	}

	switch deserializedError.ErrorCode {
	case errAuthorizationPending.ErrorCode:
		return errAuthorizationPending
	case errSlowDown.ErrorCode:
		return errSlowDown
	case errExpiredToken.ErrorCode:
		return errExpiredToken
	case errInvalidGrant.ErrorCode:
		return errInvalidGrant
	case errAccessDenied.ErrorCode:
		return errAccessDenied
	default:
		return &managedHTTPError{
			ErrorCode:        deserializedError.ErrorCode,
			ErrorDescription: deserializedError.ErrorDescription,
		}
	}
}
