package pkg

import (
	"time"

	"github.com/go-logr/logr"
	"github.com/golang-jwt/jwt/v5"
	v1 "k8s.io/api/core/v1"
)

type Coordinator struct {
	azureHelper      *AzureHelper
	kubernetesHelper *KubernetesHelper
	config           *Config
	log              logr.Logger
}

func NewCoordinator(log logr.Logger) (*Coordinator, error) {
	config, err := ParseConfig()
	if err != nil {
		return nil, err
	}
	kubernetesHelper, err := NewKubernetesHelper(log, config.Namespace)
	if err != nil {
		return nil, err
	}
	azureHelper, err := NewAzureHelper(log)
	if err != nil {
		return nil, err
	}
	return &Coordinator{
		azureHelper:      azureHelper,
		kubernetesHelper: kubernetesHelper,
		config:           config,
		log:              log,
	}, nil
}

func (c *Coordinator) EvaluateAccessTokenExpiration() error {
	urlsToEvaluate := c.config.MatchUrls
	var err error = nil
	if c.config.InClusterConfiguration {
		c.log.Info("Using in cluster configuration")
		urlsToEvaluate, err = c.kubernetesHelper.GetInClusterConfiguration(c.config.InClusterConfigMap)
		if err != nil {
			c.log.Error(err, "Failed to get in cluster configuration")
			return err
		}
	}
	c.log.Info("Urls to be evaluated", "urls", urlsToEvaluate)

	secrets, err := c.kubernetesHelper.SearchSecret(urlsToEvaluate)
	if err != nil {
		c.log.Error(err, "Failed to search secret")
		return err
	} else if secrets == nil || len(*secrets) == 0 {
		c.log.Info("No secret found")
		return nil
	}

	for _, secret := range *secrets {
		token, ok := c.evaluateTokenSecret(&secret)
		remainingTime, bufferTime := 0, 0

		if ok {
			current := time.Now()
			remainingTime = int(token.Claims.(jwt.MapClaims)["exp"].(float64)) - int(current.Unix())
			bufferTime = int((time.Minute * 5).Seconds())
		}

		if remainingTime < bufferTime || !ok {
			accessToken, err := c.azureHelper.GetAccessToken(string(secret.Data["url"]))
			if err != nil {
				c.log.Error(err, "Failed to get access token")
				return err
			}
			c.log.Info("Access token retrieved", "token", accessToken.Raw)
			err = c.kubernetesHelper.UpdateSecret(accessToken.Raw, &secret)
			if err != nil {
				c.log.Error(err, "Failed to update secret")
				return err
			}
			c.log.Info("Access token retrieved. Update expiration time", "expirationTime", time.Unix(int64(accessToken.Claims.(jwt.MapClaims)["exp"].(float64)), 0))
			return nil
		}
		c.log.Info("Access token is still valid", "remainingTime", remainingTime)
	}
	return nil
}

func (c *Coordinator) evaluateTokenSecret(secret *v1.Secret) (*jwt.Token, bool) {
	accessToken, ok := secret.Data["password"]
	if !ok {
		c.log.Info("No password found. Assuming it is missing completely.")
		return nil, false
	}

	token, _, err := new(jwt.Parser).ParseUnverified(string(accessToken), jwt.MapClaims{})
	if err != nil {
		c.log.Error(err, "Failed to parse access token. Assuming it is missing completely.")
		return nil, false
	}
	return token, true
}
