package freebox

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type TokenStore struct {
	Path string
}

func (s TokenStore) Load() (string, error) {
	data, err := os.ReadFile(s.Path)
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func (s TokenStore) Save(token string) error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(s.Path, []byte(token+"\n"), 0o600)
}
