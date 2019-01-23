package k8s

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/francoispqt/gojay"
	vault "github.com/hashicorp/vault/api"
	"github.com/lalamove/konfig/loader/klvault"
	"github.com/lalamove/nui/nfs"
)

var _ klvault.AuthProvider = (*VaultAuth)(nil)

const (
	loginPath                 = "/auth/kubernetes/login"
	k8sTokenKeyNamespace      = "kubernetes.io/serviceaccount/namespace"
	k8sTokenKeyServiceAccount = "kubernetes.io/serviceaccount/service-account.name"
)

var (
	errNoAuth         = errors.New("No authentication in login response")
	errNoClient       = errors.New("No client provided")
	errMalformedToken = errors.New("K8s token is malformed")
	fileSystem        = nfs.OSFileSystem{}
)

// VaultAuth is the structure representing a vault authentication provider
type VaultAuth struct {
	cfg           *Config
	k8sToken      string
	role          string
	logicalClient klvault.LogicalClient
}

// Config is the config of a VaultAuth provider
type Config struct {
	// Client is the vault client
	Client *vault.Client
	// K8sTokenPath is the path to the kubernetes service account jwt
	K8sTokenPath string
	// Role is the role string
	Role string
	// RoleFunc is a function to build the role
	RoleFunc func(string) (string, error)
	// FileSystem is the file system to use
	// If no value provided it uses the os file system
	FileSystem nfs.FileSystem
}

// New creates a new K8sVaultauth with the given config cfg.
func New(cfg *Config) *VaultAuth {
	// if no vault client
	if cfg.Client == nil {
		panic(errNoClient)
	}
	// if no file system use the default file system,
	if cfg.FileSystem == nil {
		cfg.FileSystem = fileSystem
	}

	var k8sVault = &VaultAuth{
		cfg:           cfg,
		logicalClient: cfg.Client.Logical(),
	}

	// load the k8s token
	var token string
	var err error
	if token, err = k8sVault.readK8sToken(); err != nil {
		panic(err)
	}
	k8sVault.k8sToken = token

	var role string
	// if role is in config, use it
	if k8sVault.cfg.Role != "" {
		role = k8sVault.cfg.Role
	} else if k8sVault.cfg.RoleFunc != nil {
		// if we have a role func run it
		if role, err = k8sVault.cfg.RoleFunc(token); err != nil {
			panic(err)
		}
	} else {
		// use the default role func
		if role, err = k8sVault.buildRole(token); err != nil {
			panic(err)
		}
	}
	k8sVault.role = role

	return k8sVault
}

// Token returns a vault token or an error if it encountered one.
// {"jwt": "'"$KUBE_TOKEN"'", "role": "{{ SERVICE_ACCOUNT_NAME }}"}
func (k *VaultAuth) Token() (string, time.Duration, error) {
	var s, err = k.logicalClient.Write(
		loginPath,
		map[string]interface{}{
			"jwt":  k.k8sToken,
			"role": k.role,
		},
	)
	if err != nil {
		return "", 0, err
	}
	// if we don't have auth return an error
	if s.Auth == nil {
		return "", 0, errNoAuth
	}
	// return the client token
	return s.Auth.ClientToken, time.Duration(s.Auth.LeaseDuration) * time.Second, nil
}

func (k *VaultAuth) readK8sToken() (string, error) {
	var f io.ReadCloser
	var err error
	if f, err = k.cfg.FileSystem.Open(k.cfg.K8sTokenPath); err != nil {
		return "", err
	}

	var b []byte
	b, err = ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(b), nil

}

func (k *VaultAuth) buildRole(k8sToken string) (string, error) {
	// the token is a JWT, we split it by dots and take what's at index 1
	var tokenSpl = strings.Split(k8sToken, ".")
	if len(tokenSpl) != 3 {
		return "", errMalformedToken
	}

	var b64TokenData = tokenSpl[1]

	var tokenData, err = base64.RawStdEncoding.DecodeString(b64TokenData)
	if err != nil {
		return "", err
	}

	var dec = gojay.BorrowDecoder(bytes.NewReader(tokenData))
	defer dec.Release()

	var namespace string
	var role string

	err = dec.Decode(gojay.DecodeObjectFunc(func(dec *gojay.Decoder, k string) error {
		switch k {
		case k8sTokenKeyNamespace:
			return dec.String(&namespace)
		case k8sTokenKeyServiceAccount:
			return dec.String(&role)
		}
		return nil
	}))

	if err != nil {
		return "", err
	}
	return namespace + "-" + role, nil
}
