package klvault

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	vault "github.com/hashicorp/vault/api"
	"github.com/lalamove/konfig"
	"github.com/lalamove/konfig/watcher/kwpoll"
	"github.com/lalamove/nui/nlogger"
	"github.com/lalamove/nui/nstrings"
)

var _ konfig.Loader = (*Loader)(nil)

var (
	defaultTTL      = 45 * time.Minute
	defaultTTLRatio = 75
	// ErrNoClient is the error thrown when trying to create a Loader without vault.Client
	ErrNoClient = errors.New("No vault client provided")
	// ErrNoAuthProvider is the error thrown when trying to create a Loader without an AuthProvider
	ErrNoAuthProvider = errors.New("No auth provider given")
	// ErrNoSecretKey is the error thrown when trying to create a Loader without a SecretKey
	ErrNoSecretKey = errors.New("No secret key given")
)

const defaultName = "vault"

// LogicalClient is a interface for the vault logical client
type LogicalClient interface {
	Read(key string) (*vault.Secret, error)
	Write(key string, data map[string]interface{}) (*vault.Secret, error)
	ReadWithData(key string, data map[string][]string) (*vault.Secret, error)
}

// Secret is a secret to load
type Secret struct {
	// Key is the URL to fetch the secret from (e.g. /v1/database/creds/mydb)
	Key string
	// KeysPrefix sets a prefix to be prepended to all keys in the config store
	KeysPrefix string
	// Replacer transforms vault secret's keys
	Replacer nstrings.Replacer
}

// Config is the config for the Loader
type Config struct {
	// Name is the name of the loader
	Name string
	// StopOnFailure tells whether a failure to load configs should closed the config and all registered closers
	StopOnFailure bool
	// Secrets is the list of secrets to load
	Secrets []Secret
	// AuthProvider is the vault auth provider
	AuthProvider AuthProvider
	// Client is the vault client for the vault loader
	Client *vault.Client
	// MaxRetry is the maximum number of times the load method can be retried
	MaxRetry int
	// RetryDelay is the time between each retry
	RetryDelay time.Duration
	// Debug enables debug mode
	Debug bool
	// Logger is the logger used for debug logs
	Logger nlogger.Provider
	// TTLRatio is the factor to multiply the key's TTL by to deduce the moment
	// the Loader should ask vault for new credentials. Default value is 75.
	// Example: ttl = 1h, ttl * 75 / 100 = 45m, the loader will refresh key after 45m
	TTLRatio int
	// Renew sets whether the vault loader should renew it self
	Renew bool
}

// Loader is the structure representing a Loader
type Loader struct {
	*kwpoll.PollWatcher
	cfg           *Config
	logicalClient LogicalClient
	mut           *sync.Mutex
	ttl           time.Duration
}

// New creates a new Loader with the given config
func New(cfg *Config) *Loader {
	if cfg.Secrets == nil || len(cfg.Secrets) == 0 {
		panic(ErrNoSecretKey)
	}
	if cfg.AuthProvider == nil {
		panic(ErrNoAuthProvider)
	}
	if cfg.Client == nil {
		panic(ErrNoClient)
	}
	if cfg.Logger == nil {
		cfg.Logger = defaultLogger()
	}
	if cfg.Name == "" {
		cfg.Name = defaultName
	}
	if cfg.TTLRatio == 0 {
		cfg.TTLRatio = defaultTTLRatio
	}
	var vl = &Loader{
		cfg:           cfg,
		logicalClient: cfg.Client.Logical(),
		mut:           &sync.Mutex{},
		ttl:           defaultTTL,
	}

	var pw *kwpoll.PollWatcher
	if cfg.Renew {
		pw = kwpoll.New(
			&kwpoll.Config{
				Debug:  cfg.Debug,
				Logger: cfg.Logger,
				Rater:  vl,
			},
		)
	}
	vl.PollWatcher = pw

	return vl
}

// Name returns the name of the loader
func (vl *Loader) Name() string { return vl.cfg.Name }

// MaxRetry is the maximum number of times the load method can be retried
func (vl *Loader) MaxRetry() int {
	return vl.cfg.MaxRetry
}

// RetryDelay is the delay between each retry
func (vl *Loader) RetryDelay() time.Duration {
	return vl.cfg.RetryDelay
}

// Load implements konfig.Loader interface.
// It fetches a token from the auth provider and sets the token in the vault client.
// Then it loads the secret and assigns it values to the konfig.Store.
func (vl *Loader) Load(cs konfig.Values) error {
	if vl.cfg.Debug {
		vl.cfg.Logger.Get().Debug(
			"Loading vault config",
		)
	}
	// everytime we load we get a new token
	// maybe we could improve implementation to use a shorter ticker and check if config if different, if yes, reload it
	var token, ttl, err = vl.cfg.AuthProvider.Token()
	if err != nil {
		vl.cfg.Logger.Get().Error(err.Error())

		return err
	}
	// we set the token in the client
	vl.cfg.Client.SetToken(token)

	var leaseDuration = int(ttl / time.Second)
	for _, secret := range vl.cfg.Secrets {
		// we fetch our secret
		var s *vault.Secret
		var sData map[string]interface{}

		k := strings.TrimSpace(secret.Key)
		k = strings.Trim(k, "/")
		if k == "" {
			return err
		}

		p, err := url.Parse(k)
		if err != nil {
			return err
		}
		s, err = vl.logicalClient.ReadWithData(p.Path, p.Query())
		if err != nil {
			return err
		}
		// checking for KV V2 for vault secret store
		// confirming version exists on metadata and it is an int
		if m, ok := s.Data["metadata"].(map[string]interface{}); ok {
			kvData, dataOK := s.Data["data"].(map[string]interface{})
			_, versionJSONNumberOK := m["version"].(json.Number)
			if versionJSONNumberOK && dataOK {
				sData = kvData
			}
		} else {
			sData = s.Data
		}
		if vl.cfg.Debug {
			vl.cfg.Logger.Get().Debug(
				fmt.Sprintf("Got secret, expiring in: %d", s.LeaseDuration),
			)
		}

		// if the current secret lease is smaller than the previous smaller lease
		// or there is no previous lease
		if s.LeaseDuration != 0 && (leaseDuration == 0 || s.LeaseDuration < leaseDuration) {
			leaseDuration = s.LeaseDuration
		}
		// we set our data on the config store
		for k, v := range sData {
			var nK = secret.KeysPrefix + k
			if secret.Replacer != nil {
				nK = secret.Replacer.Replace(nK)
			}
			cs.Set(nK, v)
		}
	}

	// reset the ttl for renewal
	vl.resetTTL(vl.cfg.TTLRatio, ttl, time.Duration(leaseDuration)*time.Second)
	return nil
}

// Time returns the TTL of the vault loader
// It is used in the ticker watcher a source.
func (vl *Loader) Time() time.Duration {
	return vl.ttl
}

// StopOnFailure returns whether a load failure should stop the config and the registered closers
func (vl *Loader) StopOnFailure() bool {
	return vl.cfg.StopOnFailure
}

func (vl *Loader) resetTTL(ttlFac int, tokenTTL, secretTTL time.Duration) {
	var ttl = tokenTTL
	if secretTTL < tokenTTL {
		ttl = secretTTL
	}
	ttl = ttl * time.Duration(ttlFac) / 100
	vl.mut.Lock()
	if ttl != vl.ttl {
		vl.ttl = ttl
	}
	vl.mut.Unlock()
}

func defaultLogger() nlogger.Provider {
	return nlogger.NewProvider(nlogger.New(os.Stdout, "VAULT CONFIG | "))
}
