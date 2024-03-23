package pkg

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/go-logr/logr"
)

type AzureHelper struct {
	logger     logr.Logger
	credential *azidentity.DefaultAzureCredential
}

func NewAzureHelper(log logr.Logger) (*AzureHelper, error) {
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Error(err, "Failed to create default azure credential")
		return nil, err
	}
	return &AzureHelper{
		logger:     log,
		credential: credential,
	}, nil
}

func (a *AzureHelper) GetAzureDevOpsAccessToken() (*azcore.AccessToken, error) {
	ctx := context.Background()
	accessToken, err := a.credential.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{Scope},
	})
	if err != nil {
		a.logger.Error(err, "Failed to get access token")
		return nil, err
	}
	return &accessToken, nil
}
