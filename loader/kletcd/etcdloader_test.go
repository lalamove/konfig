package kletcd

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/coreos/etcd/mvcc/mvccpb"
	gomock "github.com/golang/mock/gomock"
	"github.com/lalamove/konfig"
	"github.com/lalamove/konfig/mocks"
	"github.com/lalamove/konfig/watcher/kwpoll"
	"github.com/stretchr/testify/require"
	"go.etcd.io/etcd/clientv3"
)

func TestEtcdLoader(t *testing.T) {
	t.Run(
		"basic no error",
		func(t *testing.T) {
			konfig.Init(konfig.DefaultConfig())

			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()

			var mockClient = mocks.NewMockKV(ctrl)
			var mockContexter = mocks.NewMockContexter(ctrl)

			var ctx, _ = context.WithTimeout(
				context.Background(),
				5*time.Second,
			)

			mockContexter.EXPECT().WithTimeout(
				context.Background(),
				5*time.Second,
			).Times(2).Return(ctx, context.CancelFunc(func() {}))

			mockClient.EXPECT().Get(ctx, "key1").Return(
				&clientv3.GetResponse{
					Kvs: []*mvccpb.KeyValue{
						&mvccpb.KeyValue{
							Key:   []byte(`key1`),
							Value: []byte(`bar`),
						},
					},
				},
				nil,
			)

			mockClient.EXPECT().Get(ctx, "key2").Return(
				&clientv3.GetResponse{
					Kvs: []*mvccpb.KeyValue{
						&mvccpb.KeyValue{
							Key:   []byte(`key2`),
							Value: []byte(`foo`),
						},
					},
				},
				nil,
			)

			var l = New(&Config{
				Client: mockClient,
				Keys: []Key{
					{Key: "key1"},
					{Key: "key2"},
				},
				Contexter: mockContexter,
			})

			var v = konfig.Values{}

			l.Load(v)

			require.Equal(t, "bar", v["key1"])
			require.Equal(t, "foo", v["key2"])
		},
	)

	t.Run(
		"basic no error multiple result in a key",
		func(t *testing.T) {
			konfig.Init(konfig.DefaultConfig())

			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()

			var mockClient = mocks.NewMockKV(ctrl)
			var mockContexter = mocks.NewMockContexter(ctrl)

			var ctx, _ = context.WithTimeout(
				context.Background(),
				5*time.Second,
			)

			mockContexter.EXPECT().
				WithTimeout(
					context.Background(),
					5*time.Second,
				).
				Times(1).
				Return(ctx, context.CancelFunc(func() {}))

			mockClient.EXPECT().Get(ctx, "key1").Return(
				&clientv3.GetResponse{
					Kvs: []*mvccpb.KeyValue{
						&mvccpb.KeyValue{
							Key:   []byte(`key1`),
							Value: []byte(`bar`),
						},
						&mvccpb.KeyValue{
							Key:   []byte(`key2`),
							Value: []byte(`foo`),
						},
					},
				},
				nil,
			)

			var l = New(&Config{
				Client: mockClient,
				Keys: []Key{
					{Key: "key1"},
				},
				Contexter: mockContexter,
			})

			var v = konfig.Values{}

			l.Load(v)

			require.Equal(t, "bar", v["key1"])
			require.Equal(t, "foo", v["key2"])
		},
	)

	t.Run(
		"with watcher no error multiple result in a key",
		func(t *testing.T) {
			konfig.Init(konfig.DefaultConfig())

			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()

			var mockClient = mocks.NewMockKV(ctrl)
			var mockContexter = mocks.NewMockContexter(ctrl)

			var ctx, _ = context.WithTimeout(
				context.Background(),
				5*time.Second,
			)

			mockContexter.EXPECT().
				WithTimeout(
					context.Background(),
					5*time.Second,
				).
				MinTimes(1).
				Return(ctx, context.CancelFunc(func() {}))

			mockClient.EXPECT().Get(ctx, "key1").Times(1).Return(
				&clientv3.GetResponse{
					Kvs: []*mvccpb.KeyValue{
						&mvccpb.KeyValue{
							Key:   []byte(`key1`),
							Value: []byte(`bar`),
						},
						&mvccpb.KeyValue{
							Key:   []byte(`key2`),
							Value: []byte(`foo`),
						},
					},
				},
				nil,
			)

			var l = New(&Config{
				Client: mockClient,
				Keys: []Key{
					{Key: "key1"},
				},
				Watch:     true,
				Rater:     kwpoll.Time(100 * time.Millisecond),
				Contexter: mockContexter,
				Debug:     true,
			})

			mockClient.EXPECT().Get(ctx, "key1").MinTimes(1).Return(
				&clientv3.GetResponse{
					Kvs: []*mvccpb.KeyValue{
						&mvccpb.KeyValue{
							Key:   []byte(`key1`),
							Value: []byte(`bar`),
						},
						&mvccpb.KeyValue{
							Key:   []byte(`key`),
							Value: []byte(`foo`),
						},
					},
				},
				nil,
			)

			err := l.Start()
			require.Nil(t, err)

			var timer = time.NewTimer(300 * time.Millisecond)
			var watched bool
			select {
			case <-timer.C:
				l.Close()
				require.True(t, watched)
			case <-l.Watch():
				watched = true
			}
		},
	)

	t.Run(
		"no error multiple result in a key replacer prefix",
		func(t *testing.T) {
			konfig.Init(konfig.DefaultConfig())

			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()

			var mockClient = mocks.NewMockKV(ctrl)
			var mockContexter = mocks.NewMockContexter(ctrl)

			var ctx, _ = context.WithTimeout(
				context.Background(),
				5*time.Second,
			)

			mockContexter.EXPECT().
				WithTimeout(
					context.Background(),
					5*time.Second,
				).
				Times(1).
				Return(ctx, context.CancelFunc(func() {}))

			mockClient.EXPECT().Get(ctx, "key1").Return(
				&clientv3.GetResponse{
					Kvs: []*mvccpb.KeyValue{
						&mvccpb.KeyValue{
							Key:   []byte(`key1`),
							Value: []byte(`bar`),
						},
						&mvccpb.KeyValue{
							Key:   []byte(`key2`),
							Value: []byte(`foo`),
						},
					},
				},
				nil,
			)

			var l = New(&Config{
				Client: mockClient,
				Keys: []Key{
					{Key: "key1"},
				},
				Prefix:    "pfx_",
				Replacer:  strings.NewReplacer("key", "yek"),
				Contexter: mockContexter,
			})

			var v = konfig.Values{}
			l.Load(v)

			require.Equal(t, "bar", v["pfx_yek1"])
			require.Equal(t, "foo", v["pfx_yek2"])
		},
	)

	t.Run(
		"no error multiple result in a key replacer prefix",
		func(t *testing.T) {
			konfig.Init(konfig.DefaultConfig())

			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()

			var mockClient = mocks.NewMockKV(ctrl)
			var mockContexter = mocks.NewMockContexter(ctrl)

			var ctx, _ = context.WithTimeout(
				context.Background(),
				5*time.Second,
			)

			mockContexter.EXPECT().
				WithTimeout(
					context.Background(),
					5*time.Second,
				).
				Times(1).
				Return(ctx, context.CancelFunc(func() {}))

			mockClient.EXPECT().Get(ctx, "key1").Return(
				nil,
				errors.New(""),
			)

			var l = New(&Config{
				Client:    mockClient,
				Keys:      []Key{{Key: "key1"}},
				Prefix:    "pfx_",
				Replacer:  strings.NewReplacer("key", "yek"),
				Contexter: mockContexter,
			})

			err := l.Load(konfig.Values{})

			require.NotNil(t, err)
		},
	)
}

func TestLoaderMethods(t *testing.T) {
	konfig.Init(konfig.DefaultConfig())
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	var mockClient = mocks.NewMockKV(ctrl)

	var l = New(&Config{
		Name:       "etcdloader",
		MaxRetry:   1,
		RetryDelay: 10 * time.Second,
		Client:     mockClient,
		Keys:       []Key{{Key: "key1"}},
	})

	require.Equal(t, "etcdloader", l.Name())
	require.Equal(t, 1, l.MaxRetry())
	require.Equal(t, 10*time.Second, l.RetryDelay())
}
