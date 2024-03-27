package pkg

import (
	"github.com/kelseyhightower/envconfig"
)

// Config holds configuration from the env variables
type Config struct {
	Namespace                string `envconfig:"NAMESPACE" default:"argocd"`
	MatchUrl								 string `envconfig:"MATCH_URL" default:"https://dev.azure.com"`
}

// ParseConfig parses the configuration from env variables
func ParseConfig() (*Config, error) {
	c := new(Config)
	if err := envconfig.Process("config", c); err != nil {
		return c, err
	}
	return c, nil
}