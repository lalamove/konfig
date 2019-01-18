# Vault Loader
Loads config values from a vault secrets

# Usage

Basic usage with Kubernetes auth provider and renewal
```go
vaultLoader := klvault.New(&klvault.Config{
    Secrets: []klvault.Secret{
        {
            Key: "/database/creds/db"
        },
    },
    Client: vaultClient, // from github.com/hashicorp/vault/api
    AuthProvider: k8s.New(&k8s.Config{
        Client: vaultClient,
        K8sTokenPath: "/var/run/secrets/kubernetes.io/serviceaccount/token",
    }),
    Renew: true, 
})
```
