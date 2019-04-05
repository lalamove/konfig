package kwpoll

import (
	"errors"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/lalamove/konfig"
	"github.com/lalamove/konfig/mocks"
	"github.com/stretchr/testify/require"
)

func TestWatcher(t *testing.T) {
	t.Run(
		"panics not loader",
		func(t *testing.T) {
			require.Panics(t, func() {
				New(&Config{
					Diff: true,
				})
			})
		},
	)
	t.Run(
		"panics nil watcher",
		func(t *testing.T) {
			require.Panics(t, func() {
				var w *PollWatcher
				w.Start()
			})
		},
	)
	t.Run(
		"close error",
		func(t *testing.T) {
			var w = New(&Config{})
			require.Nil(t, w.Close())
			require.Equal(t, ErrAlreadyClosed, w.Close())
			require.NotNil(t, w.Close())
		},
	)
	t.Run(
		"basic watcher, no diff",
		func(t *testing.T) {
			var w = New(&Config{
				Rater: Time(100 * time.Millisecond),
				Debug: true,
			})
			w.Start()
			time.Sleep(200 * time.Millisecond)
			var timer = time.NewTimer(100 * time.Millisecond)
			select {
			case <-timer.C:
				t.Error("watcher should have ticked")
				w.Close()
			case <-w.Watch():
				w.Close()
				return
			}
		},
	)
	t.Run(
		"watcher diff",
		func(t *testing.T) {
			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()

			var g = mocks.NewMockLoader(ctrl)

			var v = konfig.Values{"foo": "bar"}

			gomock.InOrder(
				g.EXPECT().Load(konfig.Values{}).Times(1).Do(func(v konfig.Values) {
					v.Set("foo", "bar")
				}).Return(nil),
				g.EXPECT().Load(konfig.Values{}).Do(func(v konfig.Values) {
					v.Set("foo", "bar")
				}).Return(nil),
				g.EXPECT().Load(konfig.Values{}).Do(func(v konfig.Values) {
					v.Set("foo", "bar")
				}).Return(nil),
				g.EXPECT().Load(konfig.Values{}).Do(func(v konfig.Values) {
					v.Set("foo", "barr")
				}).Return(nil),
				g.EXPECT().Load(konfig.Values{}).Do(func(v konfig.Values) {
					v.Set("test", "barr")
				}).Return(nil),
				g.EXPECT().Load(konfig.Values{}).Do(func(v konfig.Values) {
				}).Return(nil),
			)

			var w = New(&Config{
				Rater:     Time(100 * time.Millisecond),
				Loader:    g,
				Diff:      true,
				Debug:     true,
				InitValue: v,
			})
			w.Start()

			var timer = time.NewTimer(620 * time.Millisecond)
			var watched int
		main:
			for {
				select {
				case <-timer.C:
					w.Close()
					break main
				case <-w.Watch():
					watched++
				}
			}

			require.Equal(t, 3, watched)
		},
	)

	t.Run(
		"watcher diff err",
		func(t *testing.T) {
			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()

			var g = mocks.NewMockLoader(ctrl)

			var v = konfig.Values{"foo": "bar"}

			gomock.InOrder(
				g.EXPECT().Load(konfig.Values{}).Times(1).Do(func(v konfig.Values) {
					v.Set("foo", "bar")
				}).Return(nil),
				g.EXPECT().Load(konfig.Values{}).Do(func(v konfig.Values) {
					v.Set("foo", "bar")
				}).Return(errors.New("Err")),
			)

			var w = New(&Config{
				Rater:     Time(100 * time.Millisecond),
				Loader:    g,
				Diff:      true,
				Debug:     true,
				InitValue: v,
			})
			w.Start()

			time.Sleep(400 * time.Millisecond)
			var timer = time.NewTimer(400 * time.Millisecond)
			var done bool
			var err error
			select {
			case <-timer.C:
				t.Error("watcher should have ticked")
				w.Close()
			case <-w.Watch():
				w.Close()
				return
			case <-w.Done():
				done = true
				err = w.Err()
				break
			}
			require.True(t, done)
			require.NotNil(t, err)
		},
	)
}
