package auth0cliauthorizer

import (
	"encoding/json"
	"os"
	"path"

	"github.com/pkg/errors"
)

type store interface {
	Save(authentication Authentication) error
	Load() (*Authentication, error)
}

type fileSystemStore struct {
	tenant   string
	basePath string
	logger   *loggerWrapper
}

var _ store = &fileSystemStore{}

func newFileSystemStore(tenant, basePath string, logger *loggerWrapper) (*fileSystemStore, error) {
	if tenant == "" || basePath == "" {
		return nil, errors.New("missing tenant or basePath")
	}

	s := &fileSystemStore{
		tenant:   tenant,
		basePath: basePath,
		logger:   logger,
	}

	p := s.fullPath()
	pPath := path.Dir(p)

	err := os.MkdirAll(pPath, os.ModePerm)
	if err != nil {
		return nil, errors.Errorf("could not create directory %s", pPath)
	}

	return s, nil
}

func (f *fileSystemStore) fullPath() string {
	return path.Join(f.basePath, "auth0-cli-auth", f.tenant+".json")
}

func (f *fileSystemStore) Save(authentication Authentication) error {
	if f.tenant == "" || f.basePath == "" {
		return errors.New("missing tenant or basePath")
	}

	serialized, err := json.MarshalIndent(authentication, "", "  ")
	if err != nil {
		return errors.Wrap(err, "error serializing authentication")
	}

	p := f.fullPath()
	f.logger.Debugf("saving authentication to %s", p)

	err = os.WriteFile(p, serialized, 0644)
	if err != nil {
		return errors.Wrap(err, "error writing to file")
	}
	f.logger.Debugf("saved authentication to %s", p)

	return nil
}

func (f *fileSystemStore) Load() (*Authentication, error) {
	if f.tenant == "" || f.basePath == "" {
		return nil, errors.New("missing tenant or basePath")
	}

	p := f.fullPath()
	if !checkFileExists(p) {
		f.logger.Debugf("no authentication available from %s", p)
		return nil, nil
	}

	f.logger.Debugf("loading authentication from %s", p)

	serialized, err := os.ReadFile(p)
	if err != nil {
		return nil, errors.Wrap(err, "error reading from file")
	}

	var deserialized Authentication
	err = json.Unmarshal(serialized, &deserialized)
	if err != nil {
		return nil, errors.Wrap(err, "error deserializing authentication")
	}

	f.logger.Debugf("loaded authentication from %s", p)

	return &deserialized, nil
}

type appDataStore struct {
	underlying store
	// TODO
}

var _ store = &appDataStore{}

func newAppDataStore(tenant string, logger *loggerWrapper) (*appDataStore, error) {
	if tenant == "" {
		return nil, errors.New("missing tenant or basePath")
	}

	path, err := os.UserCacheDir()
	if err != nil {
		return nil, errors.Wrap(err, "error detecting the user cache dir")
	}

	logger.Debugf("selected %s as app data folder", path)

	underlying, err := newFileSystemStore(tenant, path, logger)
	if err != nil {
		return nil, errors.Wrap(err, "error building the underlying file system store")
	}

	return &appDataStore{
		underlying: underlying,
	}, nil
}

func (a appDataStore) Save(authentication Authentication) error {
	return a.underlying.Save(authentication)
}

func (a appDataStore) Load() (*Authentication, error) {
	return a.underlying.Load()
}

func checkFileExists(filePath string) bool {
	_, error := os.Stat(filePath)
	return !errors.Is(error, os.ErrNotExist)
}
