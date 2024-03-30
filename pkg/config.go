package pkg

import (
	"github.com/kelseyhightower/envconfig"
)

// Config holds configuration from the env variables
type Config struct {
	Namespace              string   `envconfig:"POD_NAMESPACE" default:"argocd"`
	MatchUrls              []string `envconfig:"MATCH_URLS" default:"https://dev.azure.com"`
	InClusterConfiguration bool     `envconfig:"IN_CLUSTER_CONFIG" default:"true"`
	InClusterConfigMap		 string   `envconfig:"IN_CLUSTER_CONFIG_MAP" default:"argocd-ado-awi-cred-sidecar"`
}

// ParseConfig parses the configuration from env variables
func ParseConfig() (*Config, error) {
	c := new(Config)
	if err := envconfig.Process("config", c); err != nil {
		return c, err
	}
	return c, nil
}