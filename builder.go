package auth0cliauthorizer

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

type Option interface {
	apply(target *DefaultImpl) error
}

func New(domain, clientID, audience string, options ...Option) (*DefaultImpl, error) {
	if domain == "" {
		return nil, errors.New("missing domain")
	}
	if clientID == "" {
		return nil, errors.New("missing clientID")
	}
	if audience == "" {
		return nil, errors.New("missing audience")
	}

	domainURL, err := url.Parse(domain)
	if err != nil {
		return nil, errors.Wrap(err, "domain is not a valid URL")
	}

	if !domainURL.IsAbs() {
		return nil, errors.Wrap(err, "domain is not an absolute URL")
	}

	v := &DefaultImpl{
		domain:               domainURL,
		clientID:             clientID,
		audience:             audience,
		prefillDeviceCode:    true,
		autoOpenBrowser:      true,
		requireOfflineAccess: true,
		logger: &loggerWrapper{
			underlying: &consoleLogger{},
		},
	}

	for _, option := range options {
		if err = option.apply(v); err != nil {
			return nil, err
		}
	}

	if !v.autoOpenBrowser && v.deviceConfirmPromptCallback == nil {
		return nil, errors.New("autoOpenBrowser is disabled and no deviceConfirmPromptCallback was provided")
	}

	if v.storeBuilder != nil {
		h := md5.New()
		h.Write([]byte(domain + "|" + clientID + "|" + audience))
		hash := hex.EncodeToString(h.Sum(nil))

		storeImpl, err := v.storeBuilder(hash, v.logger)
		if err != nil {
			return nil, errors.Wrap(err, "error building the store")
		}
		v.store = storeImpl
	}

	return v, nil
}

type optionLogger struct {
	value Logger
}

func WithLogger(l Logger) Option {
	return &optionLogger{l}
}

func (o *optionLogger) apply(target *DefaultImpl) error {
	if o.value == nil {
		o.value = &noOpLogger{}
	}
	target.logger = &loggerWrapper{
		underlying: o.value,
	}
	return nil
}

type HTTPClientCustomizer func(c *http.Client)

type optionHttpClientCustomizer struct {
	value HTTPClientCustomizer
}

func WithHTTPClientCustomizer(customizer HTTPClientCustomizer) Option {
	return &optionHttpClientCustomizer{customizer}
}

func (o *optionHttpClientCustomizer) apply(target *DefaultImpl) error {
	target.httpClientCustomizer = o.value
	return nil
}

type optionPrefillDeviceCode struct {
	value bool
}

func WithPrefillDeviceCode(prefillDeviceCode bool) Option {
	return &optionPrefillDeviceCode{prefillDeviceCode}
}

func (o *optionPrefillDeviceCode) apply(target *DefaultImpl) error {
	target.prefillDeviceCode = o.value
	return nil
}

type DeviceConfirmPromptCallback func(DeviceConfirmPrompt) error

type optionDeviceConfirmPromptCallback struct {
	value DeviceConfirmPromptCallback
}

func WithDeviceConfirmPromptCallback(callback DeviceConfirmPromptCallback) Option {
	return &optionDeviceConfirmPromptCallback{callback}
}

func (o *optionDeviceConfirmPromptCallback) apply(target *DefaultImpl) error {
	target.deviceConfirmPromptCallback = o.value
	return nil
}

type optionAutoOpenBrowser struct {
	value bool
}

func WithAutoOpenBrowser(autoOpenBrowser bool) Option {
	return &optionAutoOpenBrowser{autoOpenBrowser}
}

func (o *optionAutoOpenBrowser) apply(target *DefaultImpl) error {
	target.autoOpenBrowser = o.value
	return nil
}

type optionRequireOfflineAccess struct {
	value bool
}

func WithRequireOfflineAccess(requireOfflineAccess bool) Option {
	return &optionRequireOfflineAccess{requireOfflineAccess}
}

func (o *optionRequireOfflineAccess) apply(target *DefaultImpl) error {
	target.requireOfflineAccess = o.value
	return nil
}

type storeBuilder func(hash string, logger *loggerWrapper) (store, error)

type optionStore struct {
	storeBuilder storeBuilder
	minDuration  time.Duration
}

func WithAppDataStore(minDuration time.Duration) Option {
	return &optionStore{
		minDuration: minDuration,
		storeBuilder: func(hash string, logger *loggerWrapper) (store, error) {
			return newAppDataStore(hash, logger)
		},
	}
}

func (o *optionStore) apply(target *DefaultImpl) error {
	target.storeRestoreMinDuration = o.minDuration
	target.storeBuilder = o.storeBuilder
	return nil
}
