package agentfilestorage

import (
	"io"
	"os"
	"path/filepath"
)

type LocalAgentFileStore struct {
}

func NewLocalAgentFileStorage() *LocalAgentFileStore {
	return &LocalAgentFileStore{}
}

var (
	filename = "AGENT.md"
	dir      = "agentfiles"
)

func (l *LocalAgentFileStore) RetrieveFile(name string) (string, error) {
	path := filepath.Join(dir, name, filename)
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
