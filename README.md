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
4. Add credential templates in the ArgoCD UI under repositories or add secret manifests for your required repositories. The username and password fields can be set, but will be overridden by the sidecar container afterwards.
5. Deploy a config map into the argocd namespace with the following content:
```yaml
apiVersion: v1
data:
  matchUrls: https://dev.azure.com,myacrname.azurecr.io
kind: ConfigMap
metadata:
  name: argocd-ado-awi-cred-sidecar
  namespace: argocd
```
> **_NOTE:_**  The matchUrls field is a comma separated list of URLs that should be overridden by the sidecar container. The URLs can be the base URL of Azure DevOps and/or the Azure Container Registry used for helm charts. The sidecar container will only override the credentials for the URLs in this list. The example above will override the credentials for **Azure DevOps** and the **Azure Container Registry with the name myacrname**.