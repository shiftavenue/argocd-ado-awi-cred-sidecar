package pkg

const (
	// LabelSelector is the label selector for the secret
	LabelSelector = "argocd.argoproj.io/secret-type in (repo-creds, repository)"
	// Scope for the Azure DevOps access token
	AzureDevOpsScope = "499b84ac-1321-427f-aa17-267ca6975798/.default"
	// Scope for Azure COntainer Registry access token
	AzureContainerRegistryScope = "https://management.azure.com/.default"
	// Default username for the access token
	DefaultUsername = "00000000-0000-0000-0000-000000000000"
)
