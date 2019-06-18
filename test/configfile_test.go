package test

import (
	"testing"
	"time"

	"github.com/lalamove/konfig"
	"github.com/lalamove/konfig/loader/klfile"
	"github.com/lalamove/konfig/parser/kpyaml"
	"github.com/stretchr/testify/require"
)

type DBConfig struct {
	MySQL MySQLConfig `konfig:"mysql"`
	Redis RedisConfig `konfig:"redis"`
}

type RedisConfig struct {
	H string `konfig:"host"`
}

type MySQLConfig struct {
	U           string `konfig:"username"`
	PW          string `konfig:"password"`
	MaxOpenConn int
}

type VaultConfig struct {
	Enable bool
	Server string
	Secret string `konfig:"dbSecret"`
}

type YAMLConfig struct {
	Debug    bool `konfig:"debug"`
	SQLDebug bool `konfig:"sqlDebug"`
	DB       DBConfig
	Port     string `konfig:"http.port"`
	Vault    VaultConfig
}

func TestYAMLFile(t *testing.T) {
	var expectedConfig = YAMLConfig{
		Debug:    true,
		SQLDebug: true,
		Port:     "8081",
		DB: DBConfig{
			MySQL: MySQLConfig{
				U:           "username",
				PW:          "password",
				MaxOpenConn: 10,
			},
			Redis: RedisConfig{
				H: "127.0.0.1",
			},
		},
		Vault: VaultConfig{
			Enable: true,
			Server: "http://127.0.0.1:8200",
			Secret: "/secret/db1",
		},
	}

	konfig.Init(&konfig.Config{
		NoExitOnError: true,
	})

	konfig.Bind(YAMLConfig{})

	konfig.RegisterLoader(
		klfile.New(&klfile.Config{
			Files: []klfile.File{
				{
					Path:   "./data/cfg.yml",
					Parser: kpyaml.Parser,
				},
			},
			MaxRetry:   2,
			RetryDelay: 1 * time.Second,
			Debug:      true,
		}),
	)

	if err := konfig.Load(); err != nil {
		t.Error(err)
	}

	require.Equal(t, expectedConfig, konfig.Value())
}
