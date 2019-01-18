package klvault

import (
	"errors"
	"fmt"
	"os"
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
	defaultTTL = 45 * time.Minute
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
	// SecretKey is the URL to fetch the secret from (e.g. /v1/database/creds/mydb)
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
	Logger nlogger.Logger
	// Renew sets wether the vault loader should renew it self
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
		vl.cfg.Logger.Debug(
			"Loading vault config",
		)
	}
	// everytime we load we get a new token
	// maybe we could improve implementation to use a shorter ticker and check if config if different, if yes, reload it
	var token, ttl, err = vl.cfg.AuthProvider.Token()
	if err != nil {
		vl.cfg.Logger.Error(err.Error())

		return err
	}
	// we set the token in the client
	vl.cfg.Client.SetToken(token)

	var leaseDuration = int(ttl / time.Second)
	for _, secret := range vl.cfg.Secrets {
		// we fetch our secret
		var s *vault.Secret
		s, err = vl.logicalClient.Read(secret.Key)
		if err != nil {
			return err
		}

		if vl.cfg.Debug {
			vl.cfg.Logger.Debug(
				fmt.Sprintf("Got secret, expiring in: %d", s.LeaseDuration),
			)
		}

		// if the current secret lease is smaller than the previous smaller lease
		// or there is no previous lease
		if s.LeaseDuration != 0 && (leaseDuration == 0 || s.LeaseDuration < leaseDuration) {
			leaseDuration = s.LeaseDuration
		}

		// we set our data on the config store
		for k, v := range s.Data {
			var nK = secret.KeysPrefix + k
			if secret.Replacer != nil {
				nK = secret.Replacer.Replace(nK)
			}
			cs.Set(nK, v)
		}
	}

	// reset the ttl for renewal
	vl.resetTTL(ttl, time.Duration(leaseDuration)*time.Second)
	return nil
}

// Time returns the TTL of the vault loader
// It is used in the ticker watcher a source.
func (vl *Loader) Time() time.Duration {
	return vl.ttl
}

func (vl *Loader) resetTTL(tokenTTL, secretTTL time.Duration) {
	var ttl = tokenTTL
	if secretTTL < tokenTTL {
		ttl = secretTTL
	}
	ttl = (ttl * 75) / 100
	vl.mut.Lock()
	if ttl != vl.ttl {
		vl.ttl = ttl
	}
	vl.mut.Unlock()
}

func defaultLogger() nlogger.Logger {
	return nlogger.New(os.Stdout, "VAULT CONFIG | ")
}
