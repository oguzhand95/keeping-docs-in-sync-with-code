package config

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"go.uber.org/config"
)

var conf = &configHolder{}

// GetSection populates a config section.
func GetSection(section Section) error {
	return conf.Get(section.Key(), section)
}

type configHolder struct {
	provider config.Provider
	mu       sync.RWMutex
}

func (ch *configHolder) Get(key string, out interface{}) error {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	if ch.provider == nil {
		return fmt.Errorf("config not loaded")
	}

	if err := ch.provider.Get(key).Populate(out); err != nil {
		return err
	}

	return nil
}

func (ch *configHolder) replaceProvider(provider config.Provider) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	ch.provider = provider
}

// Load loads the config file at the given path.
func Load(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat %s: %w", path, err)
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("config file path is a directory: %s", path)
	}

	return doLoad(config.File(path))
}

func doLoad(sources ...config.YAMLOption) error {
	opts := append(sources, config.Permissive()) //nolint:gocritic
	provider, err := config.NewYAML(opts...)
	if err != nil {
		if strings.Contains(err.Error(), "couldn't expand environment") {
			return fmt.Errorf("error loading configuration due to unknown environment variable. Config values containing '$' are interpreted as environment variables. Use '$$' to escape literal '$' values: [%w]", err)
		}
		return fmt.Errorf("failed to load config: %w", err)
	}

	conf.replaceProvider(provider)

	return nil
}
