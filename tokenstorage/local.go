package tokenstorage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

var dir = "tokens"

// LocalTokenStore implements Storage. Tokens are found at ./tokens/<email>
type LocalTokenStore struct {
}

func NewLocalTokenStore() *LocalTokenStore {
	return &LocalTokenStore{}
}

func (l *LocalTokenStore) StoreToken(name string, token interface{}) error {
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, name)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("caching oauth token: %w", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

func (l *LocalTokenStore) RetrieveToken(name string, token interface{}) error {
	path := filepath.Join(dir, name)
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&token)
}

func (l *LocalTokenStore) ListEmails() ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var emails []string
	for _, entry := range entries {
		emails = append(emails, entry.Name())
	}
	return emails, nil
}

func (l *LocalTokenStore) RemoveToken(name string) error {
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, name)
	return os.Remove(path)
}
