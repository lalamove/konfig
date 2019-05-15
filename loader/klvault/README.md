# Vault Loader
Loads config values from a vault secrets engine

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

It is possible to pass additional params to the vault secrets engine in the following manner:

`Key: "/aws/creds/example-role?ttl=20m"`

KV Secrets Engine - Version 2 (Versioned KV Store) is also supported by the loader, key from the versioned KV store can be accessed as follows:

`Key: "/secret/data/my-versioned-key"`

This will return the latest version of the key, a particular version of the secret can be accessed as follows:

`Key: "/secret/data/my-versioned-key?version=1"`
