package klconsul

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/consul/api"
	"github.com/lalamove/konfig"
	"github.com/lalamove/konfig/mocks"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	var testCases = []struct {
		name string
		run  func(t *testing.T, ctrl *gomock.Controller) *Loader
		err  bool
	}{
		{
			name: "key exists, no panic, no errors",
			run: func(t *testing.T, ctrl *gomock.Controller) *Loader {
				c, _ := api.NewClient(&api.Config{Address: "http://localhost"})

				var hl = New(&Config{
					Client: c,
					Keys: []Key{{
						Key: "foo",
					}},
				})

				var kvClient = mocks.NewMockKVClient(ctrl)
				kvClient.EXPECT().Get("foo", nil).Times(1).Return(
					&api.KVPair{
						Key:   "foo",
						Value: []byte(`bar`),
					},
					&api.QueryMeta{},
					nil,
				)

				hl.cfg.kvClient = kvClient

				var v = konfig.Values{}
				var err = hl.Load(v)

				if err != nil {
					t.Fatal(err)
				}
				require.Equal(t, "bar", v["foo"])
				return hl
			},
			err: false,
		},
		{
			name: "strict mode off, key doesn't exist, no panic, no error",
			run: func(t *testing.T, ctrl *gomock.Controller) *Loader {
				c, _ := api.NewClient(&api.Config{Address: "http://localhost"})

				var hl = New(&Config{
					Client:     c,
					StrictMode: false,
					Keys: []Key{{
						Key: "bar",
					}},
				})

				var kvClient = mocks.NewMockKVClient(ctrl)
				kvClient.EXPECT().Get("bar", nil).Return(
					nil,
					nil,
					nil,
				)
				hl.cfg.kvClient = kvClient

				var v = konfig.Values{}
				var err = hl.Load(v)

				if err != nil {
					t.Fatal(err)
				}

				var _, ok = v["bar"]

				require.False(t, ok)

				return hl
			},
			err: true,
		},
		{
			name: "strict mode on, key doesn't exist, no panic, error",
			run: func(t *testing.T, ctrl *gomock.Controller) *Loader {
				c, _ := api.NewClient(&api.Config{Address: "http://localhost"})

				var hl = New(&Config{
					Client:     c,
					StrictMode: true,
					Keys: []Key{{
						Key: "bar",
					}},
				})

				var kvClient = mocks.NewMockKVClient(ctrl)
				kvClient.EXPECT().Get("bar", nil).Return(
					nil,
					nil,
					nil,
				)
				hl.cfg.kvClient = kvClient

				var v = konfig.Values{}
				var err = hl.Load(v)

				require.NotNil(t, err, "err should not be nil as key was not found and strict mode is off")

				return hl
			},
			err: false,
		},
		{
			name: "multiple keys, no panic, no error",
			run: func(t *testing.T, ctrl *gomock.Controller) *Loader {
				c, _ := api.NewClient(&api.Config{Address: "http://localhost"})

				var hl = New(&Config{
					Client:     c,
					StrictMode: false,
					Keys: []Key{
						{
							Key: "key1",
						},
						{
							Key: "key2",
						}},
				})

				var kvClient = mocks.NewMockKVClient(ctrl)
				kvClient.EXPECT().Get("key1", nil).Return(
					&api.KVPair{
						Key:   "key1",
						Value: []byte(`test1`),
					},
					&api.QueryMeta{},
					nil,
				)

				kvClient.EXPECT().Get("key2", nil).Return(
					&api.KVPair{
						Key:   "key2",
						Value: []byte(`test2`),
					},
					&api.QueryMeta{},
					nil,
				)

				hl.cfg.kvClient = kvClient

				var v = konfig.Values{}
				var err = hl.Load(v)

				if err != nil {
					t.Fatal(err)
				}
				require.Equal(t, "test1", v["key1"])
				require.Equal(t, "test2", v["key2"])

				return hl
			},
			err: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				var ctrl = gomock.NewController(t)
				defer ctrl.Finish()

				konfig.Init(konfig.DefaultConfig())
				testCase.run(t, ctrl)
			},
		)
	}
}

func TestNew(t *testing.T) {
	t.Run(
		"no client panics",
		func(t *testing.T) {
			require.Panics(
				t,
				func() {
					New(&Config{})
				},
			)
		},
	)
}

func TestLoaderMethods(t *testing.T) {
	konfig.Init(konfig.DefaultConfig())
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	client, _ := api.NewClient(&api.Config{})

	var l = New(&Config{
		Name:       "consulloader",
		MaxRetry:   3,
		RetryDelay: 10 * time.Second,
		Client:     client,
		Keys:       []Key{{Key: "key1"}},
	})

	require.Equal(t, "consulloader", l.Name())
	require.Equal(t, 3, l.MaxRetry())
	require.Equal(t, 10*time.Second, l.RetryDelay())
}
