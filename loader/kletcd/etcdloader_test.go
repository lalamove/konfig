package kletcd

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/golang/mock/gomock"
	"github.com/lalamove/konfig"
	"github.com/lalamove/konfig/mocks"
	"github.com/lalamove/konfig/parser"
	"github.com/lalamove/konfig/watcher/kwpoll"
	"github.com/stretchr/testify/require"
	"go.etcd.io/etcd/clientv3"
)

func newClient() *clientv3.Client {
	c, _ := clientv3.New(clientv3.Config{
		Endpoints:   []string{"http://254.0.0.1:12345"},
		DialTimeout: 2 * time.Second,
	})
	return c
}

func TestEtcdLoader(t *testing.T) {
	t.Run(
		"basic no error",
		func(t *testing.T) {
			konfig.Init(konfig.DefaultConfig())

			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()

			var mockClient = mocks.NewMockKV(ctrl)
			var mockContexter = mocks.NewMockContexter(ctrl)

			var ctx, cancel = context.WithTimeout(
				context.Background(),
				5*time.Second,
			)
			defer cancel()

			mockContexter.EXPECT().WithTimeout(
				context.Background(),
				5*time.Second,
			).Times(2).Return(ctx, context.CancelFunc(func() {}))

			mockClient.EXPECT().Get(ctx, "key1").Return(
				&clientv3.GetResponse{
					Kvs: []*mvccpb.KeyValue{
						{
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
						{
							Key:   []byte(`key2`),
							Value: []byte(`foo`),
						},
					},
				},
				nil,
			)

			var l = New(&Config{
				Client:   newClient(),
				kvClient: mockClient,
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

			var ctx, cancel = context.WithTimeout(
				context.Background(),
				5*time.Second,
			)
			defer cancel()

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
						{
							Key:   []byte(`key1`),
							Value: []byte(`bar`),
						},
						{
							Key:   []byte(`key2`),
							Value: []byte(`foo`),
						},
					},
				},
				nil,
			)

			var l = New(&Config{
				Client:   newClient(),
				kvClient: mockClient,
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

			var timeout = time.Second

			var ctx, cancel = context.WithTimeout(
				context.Background(),
				timeout,
			)
			defer cancel()

			mockContexter.EXPECT().
				WithTimeout(
					context.Background(),
					timeout,
				).
				MinTimes(1).
				Return(ctx, context.CancelFunc(func() {}))

			mockClient.EXPECT().Get(ctx, "key1").Times(1).Return(
				&clientv3.GetResponse{
					Kvs: []*mvccpb.KeyValue{
						{
							Key:   []byte(`key1`),
							Value: []byte(`bar`),
						},
						{
							Key:   []byte(`key2`),
							Value: []byte(`foo`),
						},
					},
				},
				nil,
			)

			var l = New(&Config{
				Client:   newClient(),
				kvClient: mockClient,
				Keys: []Key{
					{Key: "key1"},
				},
				Watch:     true,
				Rater:     kwpoll.Time(100 * time.Millisecond),
				Contexter: mockContexter,
				Debug:     true,
				Timeout:   timeout,
			})

			mockClient.EXPECT().Get(ctx, "key1").MinTimes(1).Return(
				&clientv3.GetResponse{
					Kvs: []*mvccpb.KeyValue{
						{
							Key:   []byte(`key1`),
							Value: []byte(`bar`),
						},
						{
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
			for {
				select {
				case <-timer.C:
					l.Close()
					require.True(t, watched)
					return
				case <-l.Watch():
					watched = true
				}
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

			var ctx, cancel = context.WithTimeout(
				context.Background(),
				5*time.Second,
			)
			defer cancel()

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
						{
							Key:   []byte(`key1`),
							Value: []byte(`bar`),
						},
						{
							Key:   []byte(`key2`),
							Value: []byte(`foo`),
						},
					},
				},
				nil,
			)

			var l = New(&Config{
				Client:   newClient(),
				kvClient: mockClient,
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

			var ctx, cancel = context.WithTimeout(
				context.Background(),
				5*time.Second,
			)
			defer cancel()

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
				Client:    newClient(),
				kvClient:  mockClient,
				Keys:      []Key{{Key: "key1"}},
				Prefix:    "pfx_",
				Replacer:  strings.NewReplacer("key", "yek"),
				Contexter: mockContexter,
			})

			err := l.Load(konfig.Values{})

			require.NotNil(t, err)
		},
	)

	t.Run(
		"parse value fail",
		func(t *testing.T) {
			konfig.Init(konfig.DefaultConfig())

			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()

			var mockClient = mocks.NewMockKV(ctrl)
			var mockContexter = mocks.NewMockContexter(ctrl)

			var ctx, cancel = context.WithTimeout(
				context.Background(),
				5*time.Second,
			)
			defer cancel()

			mockContexter.EXPECT().WithTimeout(
				context.Background(),
				5*time.Second,
			).Times(1).Return(ctx, context.CancelFunc(func() {}))

			mockClient.EXPECT().Get(ctx, "key1").Return(
				&clientv3.GetResponse{
					Kvs: []*mvccpb.KeyValue{
						{
							Key:   []byte(`key1`),
							Value: []byte(`bar`),
						},
					},
				},
				nil,
			)

			var l = New(&Config{
				Client:   newClient(),
				kvClient: mockClient,
				Keys: []Key{
					{Key: "key1", Parser: parser.NopParser{Err: errors.New("parse fail")}},
					{Key: "key2"},
				},
				Contexter: mockContexter,
			})

			var v = konfig.Values{}
			var err = l.Load(v)
			require.Error(t, err)
		},
	)
}

func TestLoaderMethods(t *testing.T) {
	konfig.Init(konfig.DefaultConfig())
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	var mockClient = mocks.NewMockKV(ctrl)

	var l = New(&Config{
		Name:          "etcdloader",
		StopOnFailure: true,
		MaxRetry:      1,
		RetryDelay:    10 * time.Second,
		Client:        newClient(),
		kvClient:      mockClient,
		Keys:          []Key{{Key: "key1"}},
	})

	require.True(t, l.StopOnFailure())
	require.Equal(t, "etcdloader", l.Name())
	require.Equal(t, 1, l.MaxRetry())
	require.Equal(t, 10*time.Second, l.RetryDelay())
}
