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
	kubernetesHelper, err := NewKubernetesHelper(log)
	if err != nil {
		return nil, err
	}
	config, err := ParseConfig()
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
	secret, err := c.kubernetesHelper.SearchSecret(c.config.Namespace, c.config.MatchUrl)
	if err != nil {
		c.log.Error(err, "Failed to search secret")
		return err
	} else if secret == nil {
		c.log.Info("No secret found")
		return nil
	}

	token, ok := c.evaluateTokenSecret(secret)
	remainingTime, bufferTime := 0, 0

	if ok {
		current := time.Now()
		remainingTime = int(token.Claims.(jwt.MapClaims)["exp"].(float64)) - int(current.Unix())
		bufferTime = int((time.Minute * 5).Seconds())
	}

	if remainingTime < bufferTime || !ok {
		accessToken, err := c.azureHelper.GetAzureDevOpsAccessToken()
		if err != nil {
			c.log.Error(err, "Failed to get access token")
			return err
		}
		err = c.kubernetesHelper.UpdateSecret(accessToken.Token, secret)
		if err != nil {
			c.log.Error(err, "Failed to update secret")
			return err
		}
		c.log.Info("Access token retrieved. Update expiration time", "expirationTime", accessToken.ExpiresOn)
		return nil
	}
	c.log.Info("Access token is still valid", "remainingTime", remainingTime)
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
