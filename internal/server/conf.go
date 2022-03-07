package server

const ConfKey = "server"

// Conf is required configuration for the server.
//+desc=Server configuration.
type Conf struct {
	// Credentials defines the admin user credentials
	Credentials *CredsConf `yaml:"credentials"`
	// MetricsEnabled defines whether the metrics endpoint is enabled.
	MetricsEnabled bool `yaml:"metricsEnabled" conf:",example=false"`
}

func (c *Conf) Key() string {
	return ConfKey
}

type CredsConf struct {
	Username string `yaml:"username" conf:",example=\"test\""`
	Password string `yaml:"password" conf:",example=\"test\""`
}
