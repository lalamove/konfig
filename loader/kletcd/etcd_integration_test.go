// +build integration

package kletcd

import (
	"context"
	"testing"
	"time"

	"github.com/lalamove/konfig"
	"github.com/stretchr/testify/require"
	"go.etcd.io/etcd/clientv3"
)

func TestIntegrationLoad(t *testing.T) {
	c, _ := clientv3.New(clientv3.Config{
		Endpoints:   []string{"http://localhost:2379"},
		DialTimeout: 2 * time.Second,
	})

	l := New(&Config{
		Client: c,
		Keys: []Key{
			{
				Key: "foo",
			},
			{
				Key: "bar",
			},
		},
	})

	c.KV.Put(context.Background(), "foo", "bar")

	v := konfig.Values{}
	require.Nil(t, l.Load(v))

	require.Equal(t, "bar", v["foo"])
	require.Nil(t, v["bar"])
}
