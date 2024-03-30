# argocd-ado-awi-cred-sidecar
This project tries to use a sidecar container with workload identity to update Azure DevOps credentials for your ArgoCD deployment.

## How to use
1. Enable Workload Identity on your AKS cluster and create a UAMI with federated credential for the ArgoCD server.
2. Allow the UAMI to access the Azure DevOps API by adding it to you organisation and grant it ACR pull permissions on your Azure Container Registry.
3. Add the image of the sidecar container to your ArgoCD server deployment, like:
```yaml
- image: ghcr.io/shiftavenue/argocd-ado-awi-cred-sidecar
  name: argocd-ado-awi-cred-sidecar
  env:
  - name: POD_NAMESPACE
    valueFrom:
      fieldRef:
        apiVersion: v1
        fieldPath: metadata.namespace
```
4. Deploy a config map into the argocd namespace with the following content:
```yaml
apiVersion: v1
data:
  matchUrls: https://dev.azure.com,myacrname.azurecr.io
kind: ConfigMap
metadata:
  name: argocd-ado-awi-cred-sidecar
  namespace: argocd
```