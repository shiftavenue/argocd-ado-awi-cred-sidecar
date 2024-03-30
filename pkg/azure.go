package pkg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/go-logr/logr"
	"github.com/golang-jwt/jwt/v5"
)

type AzureHelper struct {
	log        logr.Logger
	credential *azidentity.DefaultAzureCredential
}

func NewAzureHelper(log logr.Logger) (*AzureHelper, error) {
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Error(err, "Failed to create default azure credential")
		return nil, err
	}
	return &AzureHelper{
		log:        log,
		credential: credential,
	}, nil
}

func (a *AzureHelper) GetAccessToken(url string) (*jwt.Token, error) {
	if strings.Contains(url, "dev.azure.com") {
		return a.getAzureDevOpsAccessToken()
	} else if strings.Contains(url, "azurecr.io") {
		return a.getAzureContainerRegistryAccessToken(url)
	} else {
		a.log.Info("Cannot match url to scope", "url", url)
		return nil, errors.New("cannot match url to scope")
	}
}

func (a *AzureHelper) getAzureDevOpsAccessToken() (*jwt.Token, error) {
	accessToken, err := a.credential.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{AzureDevOpsScope},
	})
	if err != nil {
		a.log.Error(err, "Failed to get access token")
		return nil, err
	}
	token, _, err := new(jwt.Parser).ParseUnverified(string(accessToken.Token), jwt.MapClaims{})
	if err != nil {
		a.log.Error(err, "Failed to parse access token.")
		return nil, err
	}
	return token, nil
}

func (a *AzureHelper) getAzureContainerRegistryAccessToken(acrServiceName string) (*jwt.Token, error) {
	ctx := context.Background()
	aadToken, err := a.credential.GetToken(ctx, policy.TokenRequestOptions{Scopes: []string{AzureContainerRegistryScope}})
	if err != nil {
		panic(err)
	}
	aadTokenJWT, _, err := new(jwt.Parser).ParseUnverified(string(aadToken.Token), jwt.MapClaims{})
	if err != nil {
		a.log.Error(err, "Failed to parse AAD token.")
		return nil, err
	}
	tenantId := aadTokenJWT.Claims.(jwt.MapClaims)["tid"].(string)
	formData := url.Values{
		"grant_type":   {"access_token"},
		"service":      {acrServiceName},
		"tenant":       {tenantId},
		"access_token": {aadToken.Token},
	}
	jsonResponse, err := http.PostForm(fmt.Sprintf("https://%s/oauth2/exchange", acrServiceName), formData)
	if err != nil {
		panic(err)
	}
	var response map[string]interface{}
	json.NewDecoder(jsonResponse.Body).Decode(&response)
	a.log.Info("Response from ACR", "response", response)
	token, _, err := new(jwt.Parser).ParseUnverified(response["refresh_token"].(string), jwt.MapClaims{})
	if err != nil {
		a.log.Error(err, "Failed to parse access token.")
		return nil, err
	}
	return token, nil
}
