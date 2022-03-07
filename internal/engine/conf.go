package engine

const confKey = "engine"

// Conf is required configuration for the engine.
//+desc=Engine configuration.
type Conf struct {
	// CacheSize defines the size of the cache in terms of number of policies.
	CacheSize int `yaml:"cacheSize" conf:",example=100"`
}

func (c *Conf) Key() string {
	return confKey
}
