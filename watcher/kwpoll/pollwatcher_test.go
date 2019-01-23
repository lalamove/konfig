package kwpoll

import (
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/lalamove/konfig"
	"github.com/lalamove/konfig/mocks"
)

func TestWatcher(t *testing.T) {
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
}
