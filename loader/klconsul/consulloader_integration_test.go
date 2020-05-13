// +build integration

package klconsul

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/sdk/testutil"
	"github.com/lalamove/konfig"
	"github.com/stretchr/testify/require"
)

func TestIntegrationLoad(t *testing.T) {
	srv, err := testutil.NewTestServerConfigT(t, func(c *testutil.TestServerConfig) {
		c.LogLevel = "err"
	})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Stop()

	var testCases = []struct {
		name  string
		setUp func(ctrl *gomock.Controller) *Loader
		err   bool
	}{
		{
			name: "key exists, no panic, no errors",
			setUp: func(ctrl *gomock.Controller) *Loader {
				c, _ := api.NewClient(&api.Config{Address: srv.HTTPAddr})

				var hl = New(&Config{
					Client: c,
					Keys: []Key{{
						Key: "foo",
					}},
				})

				srv.SetKV(t, "foo", []byte("bar"))

				kv, _, err := c.KV().Get("foo", nil)
				if err != nil {
					t.Fatal(err)
				}

				require.Equal(t, []byte("bar"), kv.Value)

				return hl
			},
			err: false,
		},
		{
			name: "strict mode on, key doesn't exist, no panic, error",
			setUp: func(ctrl *gomock.Controller) *Loader {
				c, _ := api.NewClient(&api.Config{Address: srv.HTTPAddr})

				var hl = New(&Config{
					Client:     c,
					StrictMode: true,
					Keys: []Key{{
						Key: "bar",
					}},
				})

				kv, _, err := c.KV().Get("bar", nil)
				if err != nil {
					t.Fatal(err)
				}

				require.Nil(t, kv)

				return hl
			},
			err: true,
		},
		{
			name: "strict mode off, key doesn't exist, no panic, error",
			setUp: func(ctrl *gomock.Controller) *Loader {
				c, _ := api.NewClient(&api.Config{Address: srv.HTTPAddr})

				var hl = New(&Config{
					Client:     c,
					StrictMode: false,
					Keys: []Key{{
						Key: "bar",
					}},
				})

				kv, _, err := c.KV().Get("bar", nil)
				if err != nil {
					t.Fatal(err)
				}

				require.Nil(t, kv)

				return hl
			},
			err: false,
		},
		{
			name: "multiple keys, no panic, no error",
			setUp: func(ctrl *gomock.Controller) *Loader {
				c, _ := api.NewClient(&api.Config{Address: srv.HTTPAddr})

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

				srv.SetKV(t, "key1", []byte("test"))
				srv.SetKV(t, "key2", []byte("test"))

				kv1, _, err := c.KV().Get("key1", nil)
				if err != nil {
					t.Fatal(err)
				}

				kv2, _, err := c.KV().Get("key2", nil)
				if err != nil {
					t.Fatal(err)
				}

				require.Equal(t, []byte("test"), kv1.Value)
				require.Equal(t, []byte("test"), kv2.Value)

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
				var hl = testCase.setUp(ctrl)

				var err = hl.Load(konfig.Values{})
				if testCase.err {
					require.NotNil(t, err, "err should not be nil")
					return
				}
				require.Nil(t, err, "err should be nil")
			},
		)
	}
}
