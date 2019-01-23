package k8s

import (
	"encoding/base64"
	"errors"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	vault "github.com/hashicorp/vault/api"
	"github.com/lalamove/konfig/mocks"
	"github.com/lalamove/nui/nfs"
	"github.com/stretchr/testify/require"
)

func TestNewK8sAuth(t *testing.T) {
	t.Run(
		"new no error, uses default role builder",
		func(t *testing.T) {
			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()

			var fs = nfs.NewMockFileSystem(ctrl)

			fs.EXPECT().
				Open("test").
				Return(
					ioutil.NopCloser(strings.NewReader(
						"12345."+base64.RawStdEncoding.EncodeToString([]byte(
							`{"kubernetes.io/serviceaccount/namespace":"dev","kubernetes.io/serviceaccount/service-account.name":"vault-config-loader"}`,
						))+".12345")),
					nil,
				)

			var c, _ = vault.NewClient(vault.DefaultConfig())

			var k8sAuth = New(&Config{
				K8sTokenPath: "test",
				Client:       c,
				FileSystem:   fs,
			})

			require.Equal(t, "dev-vault-config-loader", k8sAuth.role)
		},
	)

	t.Run(
		"new no error, uses config role",
		func(t *testing.T) {
			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()

			var fs = nfs.NewMockFileSystem(ctrl)

			fs.EXPECT().
				Open("test").
				Return(
					ioutil.NopCloser(strings.NewReader(
						"12345."+base64.RawStdEncoding.EncodeToString([]byte(
							`{"kubernetes.io/serviceaccount/namespace":"dev","kubernetes.io/serviceaccount/service-account.name":"vault-config-loader"}`,
						))+".12345")),
					nil,
				)

			var c, _ = vault.NewClient(vault.DefaultConfig())

			var k8sAuth = New(&Config{
				K8sTokenPath: "test",
				Client:       c,
				FileSystem:   fs,
				Role:         "foobar",
			})

			require.Equal(t, "foobar", k8sAuth.role)
		},
	)

	t.Run(
		"new no error, uses config role error invalid base64",
		func(t *testing.T) {
			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()

			var fs = nfs.NewMockFileSystem(ctrl)

			fs.EXPECT().
				Open("test").
				Return(
					ioutil.NopCloser(strings.NewReader(
						"12345.!##%$.12345")),
					nil,
				)

			var c, _ = vault.NewClient(vault.DefaultConfig())

			require.Panics(
				t,
				func() {
					New(&Config{
						K8sTokenPath: "test",
						Client:       c,
						FileSystem:   fs,
					})
				},
			)
		},
	)

	t.Run(
		"new no error, uses config role func",
		func(t *testing.T) {
			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()

			var fs = nfs.NewMockFileSystem(ctrl)

			fs.EXPECT().
				Open("test").
				Return(
					ioutil.NopCloser(strings.NewReader(
						"12345."+base64.RawStdEncoding.EncodeToString([]byte(
							`{"kubernetes.io/serviceaccount/namespace":"dev","kubernetes.io/serviceaccount/service-account.name":"vault-config-loader"}`,
						))+".12345")),
					nil,
				)

			var c, _ = vault.NewClient(vault.DefaultConfig())

			var k8sAuth = New(&Config{
				K8sTokenPath: "test",
				Client:       c,
				FileSystem:   fs,
				RoleFunc: func(string) (string, error) {
					return "foobar", nil
				},
			})

			require.Equal(t, "foobar", k8sAuth.role)
		},
	)

	t.Run(
		"new no error, uses config role func with error",
		func(t *testing.T) {
			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()

			var fs = nfs.NewMockFileSystem(ctrl)

			fs.EXPECT().
				Open("test").
				Return(
					ioutil.NopCloser(strings.NewReader(
						"12345."+base64.RawStdEncoding.EncodeToString([]byte(
							`{"kubernetes.io/serviceaccount/namespace":"dev","kubernetes.io/serviceaccount/service-account.name":"vault-config-loader"}`,
						))+".12345")),
					nil,
				)

			var c, _ = vault.NewClient(vault.DefaultConfig())

			require.Panics(t, func() {
				New(&Config{
					K8sTokenPath: "test",
					Client:       c,
					FileSystem:   fs,
					RoleFunc: func(string) (string, error) {
						return "", errors.New("err")
					},
				})
			})
		},
	)

	t.Run(
		"new panics no client",
		func(t *testing.T) {
			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()

			require.Panics(t, func() {
				New(&Config{
					K8sTokenPath: "test",
					RoleFunc: func(string) (string, error) {
						return "foobar", nil
					},
				})
			})
		},
	)
}

func TestBuildRole(t *testing.T) {
	var testCases = []struct {
		name         string
		token        string
		expectedRole string
		err          bool
	}{
		{
			token: "12345." + base64.RawStdEncoding.EncodeToString([]byte(
				`{"kubernetes.io/serviceaccount/namespace":"dev","kubernetes.io/serviceaccount/service-account.name":"vault-config-loader"}`,
			)) + ".ABCD",
			expectedRole: "dev-vault-config-loader",
		},
		{
			token: "ABCDE",
			err:   true,
		},
		{
			token: "12345." + base64.RawStdEncoding.EncodeToString([]byte(
				`{"kubernetes.io/serviceaccount/namespace":"dev","kubernetes.io/serviceaccount/service-account.name":"vault-config-loader"}`,
			)) + ".ABCD",
			expectedRole: "dev-vault-config-loader",
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				var k8sAuth = &VaultAuth{}
				var s, err = k8sAuth.buildRole(testCase.token)
				if testCase.err {
					require.NotNil(t, err, "err should not be nil")
					return
				}
				require.Nil(t, err, "err should be nil")
				require.Equal(t, testCase.expectedRole, s)
			},
		)
	}
}

func TestToken(t *testing.T) {
	t.Run(
		"no error",
		func(t *testing.T) {
			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()

			var logicalClient = mocks.NewMockLogicalClient(ctrl)
			logicalClient.EXPECT().Write(
				loginPath,
				map[string]interface{}{
					"jwt":  "123",
					"role": "role",
				},
			).Times(1).Return(&vault.Secret{
				Auth: &vault.SecretAuth{
					ClientToken:   "123",
					LeaseDuration: 3600,
				},
			}, nil)

			var k = &VaultAuth{
				k8sToken:      "123",
				role:          "role",
				logicalClient: logicalClient,
			}

			var token, d, err = k.Token()
			require.Equal(t, "123", token)
			require.Equal(t, 3600*time.Second, d)
			require.Nil(t, err)
		},
	)

	t.Run(
		"error when calling vault",
		func(t *testing.T) {
			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()

			var logicalClient = mocks.NewMockLogicalClient(ctrl)
			logicalClient.EXPECT().Write(
				loginPath,
				map[string]interface{}{
					"jwt":  "123",
					"role": "role",
				},
			).Times(1).Return(
				nil,
				errors.New("err"),
			)

			var k = &VaultAuth{
				k8sToken:      "123",
				role:          "role",
				logicalClient: logicalClient,
			}

			var _, _, err = k.Token()
			require.NotNil(t, err)
		},
	)
}
