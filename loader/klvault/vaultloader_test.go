package klvault

import (
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	vault "github.com/hashicorp/vault/api"
	"github.com/lalamove/konfig"
	"github.com/lalamove/konfig/mocks"
	"github.com/stretchr/testify/require"
)

func TestVaultLoader(t *testing.T) {
	var testCases = []struct {
		name    string
		setUp   func(ctrl *gomock.Controller) *Loader
		asserts func(t *testing.T, vl *Loader, cfg konfig.Values)
		err     bool
	}{
		{
			name: "BasicNoError",
			setUp: func(ctrl *gomock.Controller) *Loader {
				var aP = mocks.NewMockAuthProvider(ctrl)
				aP.EXPECT().Token().Return(
					"DUMMYTOKEN",
					1*time.Hour,
					nil,
				)

				var c, _ = vault.NewClient(vault.DefaultConfig())

				var vl = New(&Config{
					Client: c,
					Secrets: []Secret{
						{Key: "/dummy/secret/path"},
						{Key: "/dummy/secret/path2"},
						{Key: "/dummy/secret/data/path3"},
						{Key: "/dummy/secret/data/path3?version=1"},
					},
					AuthProvider: aP,
					Debug:        true,
				})

				var lC = mocks.NewMockLogicalClient(ctrl)
				vl.logicalClient = lC
				lC.EXPECT().ReadWithData("dummy/secret/path", map[string][]string{}).Return(
					&vault.Secret{
						Data: map[string]interface{}{
							"FOO": "BAR",
						},
						LeaseDuration: int(2 * time.Hour / time.Second),
					},
					nil,
				)

				lC.EXPECT().ReadWithData("dummy/secret/path2", map[string][]string{}).Return(
					&vault.Secret{
						Data: map[string]interface{}{
							"BAR": "FOO",
						},
						LeaseDuration: int(1 * time.Hour / time.Second),
					},
					nil,
				)
				lC.EXPECT().ReadWithData("dummy/secret/data/path3", map[string][]string{}).Return(
					&vault.Secret{
						Data: map[string]interface{}{
							"data": map[string]interface{}{
								"VERSIONEDFOO": "FOO2",
							},
							"metadata": map[string]interface{}{
								"created_time":  "2018-03-22T02:24:06.945319214Z",
								"deletion_time": "",
								"destroyed":     false,
								"version":       json.Number("1"),
							},
						},
						LeaseDuration: int(1 * time.Hour / time.Second),
					},
					nil,
				)
				lC.EXPECT().ReadWithData("dummy/secret/data/path3", map[string][]string{"version": {"1"}}).Return(
					&vault.Secret{
						Data: map[string]interface{}{
							"data": map[string]interface{}{
								"OLDFOO": "FOO1",
							},
							"metadata": map[string]interface{}{
								"created_time":  "2018-03-22T02:24:06.945319214Z",
								"deletion_time": "",
								"destroyed":     false,
								"version":       json.Number("1"),
							},
						},
						LeaseDuration: int(1 * time.Hour / time.Second),
					},
					nil,
				)

				return vl
			},
			asserts: func(t *testing.T, vl *Loader, cfg konfig.Values) {
				require.Equal(
					t,
					45*time.Minute,
					vl.Time(),
				)
				require.Equal(
					t,
					"BAR",
					cfg["FOO"],
				)
				require.Equal(
					t,
					"FOO",
					cfg["BAR"],
				)
				require.Equal(
					t,
					"FOO2",
					cfg["VERSIONEDFOO"],
				)
				require.Equal(
					t,
					"FOO1",
					cfg["OLDFOO"],
				)
				require.Equal(
					t,
					defaultName,
					vl.Name(),
				)
			},
		},
		{
			name: "ErrorOnAuthProvider",
			err:  true,
			setUp: func(ctrl *gomock.Controller) *Loader {
				var aP = mocks.NewMockAuthProvider(ctrl)
				aP.EXPECT().Token().Return(
					"",
					time.Duration(0),
					errors.New(""),
				)

				var c, _ = vault.NewClient(vault.DefaultConfig())

				var vl = New(&Config{
					Client:       c,
					Secrets:      []Secret{{Key: "/dummy/secret/path"}},
					AuthProvider: aP,
				})
				return vl
			},
			asserts: func(t *testing.T, vl *Loader, cfg konfig.Values) {},
		},
		{
			name: "ErrorFetchingSecret",
			err:  true,
			setUp: func(ctrl *gomock.Controller) *Loader {
				var aP = mocks.NewMockAuthProvider(ctrl)
				aP.EXPECT().Token().Return(
					"DUMMYTOKEN",
					1*time.Hour,
					nil,
				)

				var c, _ = vault.NewClient(vault.DefaultConfig())

				var vl = New(&Config{
					Client:       c,
					Secrets:      []Secret{{Key: "/dummy/secret/path"}},
					AuthProvider: aP,
				})

				var lC = mocks.NewMockLogicalClient(ctrl)
				vl.logicalClient = lC
				lC.EXPECT().ReadWithData("dummy/secret/path", map[string][]string{}).Return(
					nil,
					errors.New(""),
				)

				return vl
			},
			asserts: func(t *testing.T, vl *Loader, cfg konfig.Values) {},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()
			var vl = testCase.setUp(ctrl)
			konfig.Init(&konfig.Config{})
			var c = konfig.Values{}

			var err = vl.Load(c)

			if testCase.err {
				require.NotNil(t, err, "err should not be nil")
				return
			}

			require.Nil(t, err, "err should be nil")
			testCase.asserts(t, vl, c)
		})
	}
}

func TestResetTTL(t *testing.T) {
	var testCases = []struct {
		name        string
		tokenTTL    time.Duration
		secretTTL   time.Duration
		expectedTTL time.Duration
	}{
		{
			name:        "token TTL is smaller than secret TTL",
			tokenTTL:    1 * time.Hour,
			secretTTL:   2 * time.Hour,
			expectedTTL: 45 * time.Minute,
		},
		{
			name:        "token TTL is smaller than secret TTL",
			tokenTTL:    1 * time.Hour,
			secretTTL:   30 * time.Minute,
			expectedTTL: 1350 * time.Second,
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				var vl = &Loader{
					mut: &sync.Mutex{},
				}
				vl.resetTTL(75, testCase.tokenTTL, testCase.secretTTL)
				require.Equal(t, testCase.expectedTTL, vl.Time())
			},
		)
	}
}

func TestNew(t *testing.T) {
	t.Run(
		"no secret key panics",
		func(t *testing.T) {
			require.Panics(
				t,
				func() {
					New(&Config{})
				},
			)
		},
	)

	t.Run(
		"no auth provider panics",
		func(t *testing.T) {
			require.Panics(
				t,
				func() {
					New(&Config{
						Secrets: []Secret{{Key: "/dummy/secret/path"}},
					})
				},
			)
		},
	)

	t.Run(
		"no vault client panics",
		func(t *testing.T) {
			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()
			var aP = mocks.NewMockAuthProvider(ctrl)
			require.Panics(
				t,
				func() {
					New(&Config{
						Secrets:      []Secret{{Key: "/dummy/secret/path"}},
						AuthProvider: aP,
					})
				},
			)
		},
	)

	t.Run(
		"no panic, no renewal",
		func(t *testing.T) {
			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()
			var aP = mocks.NewMockAuthProvider(ctrl)
			var c, _ = vault.NewClient(
				vault.DefaultConfig(),
			)
			var vl = New(&Config{

				Secrets:      []Secret{{Key: "/dummy/secret/path"}},
				AuthProvider: aP,
				Client:       c,
			})

			require.Nil(t, vl.PollWatcher)
		},
	)

	t.Run(
		"no panic, with renewal",
		func(t *testing.T) {
			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()
			var aP = mocks.NewMockAuthProvider(ctrl)
			var c, _ = vault.NewClient(
				vault.DefaultConfig(),
			)
			var vl = New(&Config{
				Secrets:      []Secret{{Key: "/dummy/secretr/path"}},
				AuthProvider: aP,
				Client:       c,
				Renew:        true,
			})

			require.NotNil(t, vl.PollWatcher)
		},
	)
}

func TestMaxRetryRetryDelay(t *testing.T) {
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()
	var aP = mocks.NewMockAuthProvider(ctrl)
	var c, _ = vault.NewClient(
		vault.DefaultConfig(),
	)
	var vl = New(&Config{
		Secrets:       []Secret{{Key: "/dummy/secretr/path"}},
		AuthProvider:  aP,
		Client:        c,
		Renew:         true,
		StopOnFailure: true,
		MaxRetry:      1,
		RetryDelay:    1 * time.Second,
	})

	require.True(t, vl.StopOnFailure())
	require.Equal(t, 1, vl.MaxRetry())
	require.Equal(t, 1*time.Second, vl.RetryDelay())
}
