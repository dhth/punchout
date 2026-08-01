package issuecache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dhth/punchout/internal/config"
	"github.com/dhth/punchout/internal/domain"
)

type Store struct {
	filePath string
}

type Snapshot struct {
	Issues    []domain.Issue `json:"issues"`
	FetchedAt time.Time      `json:"fetched_at"`
}

func NewStore(userCacheDir string, installation config.JiraInstallation, jql string) (Store, error) {
	key, err := deriveKey(installation, jql)
	if err != nil {
		return Store{}, err
	}

	return Store{
		filePath: filepath.Join(userCacheDir, "punchout", "issues", key+".json"),
	}, nil
}

func (s Store) Load() (Snapshot, error) {
	contents, err := os.ReadFile(s.filePath)
	if err != nil {
		return Snapshot{}, fmt.Errorf("couldn't read issue cache file: %w", err)
	}

	var snapshot Snapshot
	if err := json.Unmarshal(contents, &snapshot); err != nil {
		return Snapshot{}, fmt.Errorf("couldn't decode issue cache: %w", err)
	}
	if snapshot.Issues == nil {
		return Snapshot{}, errors.New("issues are missing or null in the cache file")
	}
	if snapshot.FetchedAt.IsZero() {
		return Snapshot{}, errors.New("cache file has a zero fetched-at timestamp")
	}

	return snapshot, nil
}

func (s Store) Save(snapshot Snapshot) error {
	if snapshot.FetchedAt.IsZero() {
		return errors.New("fetched-at timestamp is zero")
	}

	if snapshot.Issues == nil {
		snapshot.Issues = []domain.Issue{}
	}

	encodedSnapshot, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("couldn't encode issue cache snapshot: %w", err)
	}

	cacheDir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(cacheDir, 0o700); err != nil {
		return fmt.Errorf("couldn't create issue cache directory: %w", err)
	}

	tempFile, err := os.CreateTemp(cacheDir, ".issues-*.tmp")
	if err != nil {
		return fmt.Errorf("couldn't create temporary issue cache file: %w", err)
	}
	tempFilePath := tempFile.Name()
	defer func() {
		_ = os.Remove(tempFilePath)
	}()

	if _, err := tempFile.Write(encodedSnapshot); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("couldn't write issue cache snapshot: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("couldn't close temporary issue cache file: %w", err)
	}
	if err := os.Rename(tempFilePath, s.filePath); err != nil {
		return fmt.Errorf("couldn't replace issue cache file: %w", err)
	}

	return nil
}
