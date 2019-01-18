package kletcd

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/lalamove/konfig"
	"github.com/lalamove/konfig/mocks"
	"github.com/lalamove/konfig/watcher/kwpoll"
	"github.com/coreos/etcd/mvcc/mvccpb"
	gomock "github.com/golang/mock/gomock"
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
			})

			l.cfg.contexter = mockContexter

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
			})

			l.cfg.contexter = mockContexter

			var v = konfig.Values{}

			l.Load(v)

			require.Equal(t, "bar", v["key1"])
			require.Equal(t, "foo", v["key2"])
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
				Prefix:   "pfx_",
				Replacer: strings.NewReplacer("key", "yek"),
			})

			l.cfg.contexter = mockContexter

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
				Client:   mockClient,
				Keys:     []Key{{Key: "key1"}},
				Prefix:   "pfx_",
				Replacer: strings.NewReplacer("key", "yek"),
			})

			l.cfg.contexter = mockContexter

			err := l.Load(konfig.Values{})

			require.NotNil(t, err)
		},
	)

}

func TestGetter(t *testing.T) {
	t.Run(
		"multiple keys no error",
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
				Client:   mockClient,
				Keys:     []Key{{Key: "key1"}},
				Prefix:   "pfx_",
				Replacer: strings.NewReplacer("key", "yek"),
			})

			l.cfg.contexter = mockContexter

			r, err := l.Get()

			require.Nil(t, err)

			require.Equal(
				t,
				map[string]map[string][]byte{
					"key1": map[string][]byte{
						"pfx_yek1": []byte("bar"),
						"pfx_yek2": []byte("foo"),
					},
				},
				r,
			)
		},
	)

	t.Run(
		"loader with watcher and parser on keys",
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

			mockClient.EXPECT().Get(ctx, "key1").Times(2).Return(
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
				Client:    mockClient,
				Keys:      []Key{{Key: "key1"}},
				Prefix:    "pfx_",
				Replacer:  strings.NewReplacer("key", "yek"),
				Watch:     true,
				Rater:     kwpoll.Time(1 * time.Second),
				contexter: mockContexter,
			})

			r, err := l.Get()

			require.Nil(t, err)

			require.Equal(
				t,
				map[string]map[string][]byte{
					"key1": map[string][]byte{
						"pfx_yek1": []byte("bar"),
						"pfx_yek2": []byte("foo"),
					},
				},
				r,
			)

		},
	)
}
