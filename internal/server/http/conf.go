package http

import "github.com/oguzhand95/keeping-docs-in-sync-with-code/internal/server"

const confKey = server.ConfKey + ".http"

// Conf is optional configuration for the http service.
//+desc=HTTP configuration.
type Conf struct {
	// HTTPAddr is the dedicated HTTP address.
	HTTPAddr string `yaml:"httpAddr" conf:"required,example=\":8080\""`
}

func (c *Conf) Key() string {
	return confKey
}
