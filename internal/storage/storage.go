package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"vapeu/internal/models"
	"gopkg.in/yaml.v3"
)

type Storage struct {
	baseDir string
	mu      sync.RWMutex
}

func NewStorage(customDir string) (*Storage, error) {
	dir := customDir
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			dir = ".apiclient"
		} else {
			dir = filepath.Join(home, ".apiclient")
		}
	}

	dirs := []string{
		dir,
		filepath.Join(dir, "collections"),
		filepath.Join(dir, "environments"),
		filepath.Join(dir, "workspaces"),
		filepath.Join(dir, "cache"),
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", d, err)
		}
	}

	return &Storage{baseDir: dir}, nil
}

func (s *Storage) BaseDir() string {
	return s.baseDir
}

// Config Operations
func (s *Storage) LoadConfig() (models.Config, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cfg := models.NewDefaultConfig()
	path := filepath.Join(s.baseDir, "config.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Save default config
			_ = s.SaveConfig(cfg)
			return cfg, nil
		}
		return cfg, err
	}

	err = yaml.Unmarshal(data, &cfg)
	return cfg, err
}

func (s *Storage) SaveConfig(cfg models.Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.baseDir, "config.yaml")
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Collections Operations
func (s *Storage) LoadCollections() ([]models.Collection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := filepath.Join(s.baseDir, "collections")
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var collections []models.Collection
	for _, f := range files {
		if f.IsDir() || (filepath.Ext(f.Name()) != ".yaml" && filepath.Ext(f.Name()) != ".json") {
			continue
		}

		path := filepath.Join(dir, f.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var col models.Collection
		if filepath.Ext(f.Name()) == ".json" {
			err = json.Unmarshal(data, &col)
		} else {
			err = yaml.Unmarshal(data, &col)
		}

		if err == nil && col.ID != "" {
			collections = append(collections, col)
		}
	}

	return collections, nil
}

func (s *Storage) SaveCollection(col models.Collection) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if col.ID == "" {
		return fmt.Errorf("collection ID cannot be empty")
	}

	path := filepath.Join(s.baseDir, "collections", col.ID+".yaml")
	data, err := yaml.Marshal(col)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (s *Storage) DeleteCollection(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	pathYaml := filepath.Join(s.baseDir, "collections", id+".yaml")
	pathJson := filepath.Join(s.baseDir, "collections", id+".json")

	_ = os.Remove(pathYaml)
	_ = os.Remove(pathJson)
	return nil
}

// Environments Operations
func (s *Storage) LoadEnvironments() ([]models.Environment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := filepath.Join(s.baseDir, "environments")
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var envs []models.Environment
	for _, f := range files {
		if f.IsDir() || (filepath.Ext(f.Name()) != ".yaml" && filepath.Ext(f.Name()) != ".json") {
			continue
		}

		path := filepath.Join(dir, f.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var env models.Environment
		if filepath.Ext(f.Name()) == ".json" {
			_ = json.Unmarshal(data, &env)
		} else {
			_ = yaml.Unmarshal(data, &env)
		}

		if env.ID != "" {
			envs = append(envs, env)
		}
	}

	return envs, nil
}

func (s *Storage) SaveEnvironment(env models.Environment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if env.ID == "" {
		return fmt.Errorf("environment ID cannot be empty")
	}

	path := filepath.Join(s.baseDir, "environments", env.ID+".yaml")
	data, err := yaml.Marshal(env)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (s *Storage) DeleteEnvironment(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_ = os.Remove(filepath.Join(s.baseDir, "environments", id+".yaml"))
	_ = os.Remove(filepath.Join(s.baseDir, "environments", id+".json"))
	return nil
}

// History Operations
func (s *Storage) LoadHistory() ([]models.HistoryItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.baseDir, "history.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.HistoryItem{}, nil
		}
		return nil, err
	}

	var items []models.HistoryItem
	err = json.Unmarshal(data, &items)
	return items, err
}

func (s *Storage) SaveHistory(items []models.HistoryItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.baseDir, "history.json")
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (s *Storage) AddHistoryItem(item models.HistoryItem) error {
	items, err := s.LoadHistory()
	if err != nil {
		items = []models.HistoryItem{}
	}

	// Prepend item, keep maximum 200 history items
	items = append([]models.HistoryItem{item}, items...)
	if len(items) > 200 {
		items = items[:200]
	}

	return s.SaveHistory(items)
}

func (s *Storage) ClearHistory() error {
	return s.SaveHistory([]models.HistoryItem{})
}
