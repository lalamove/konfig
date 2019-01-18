package kwpoll

import (
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
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

			var g = mocks.NewMockGetter(ctrl)

			gomock.InOrder(
				g.EXPECT().Get().Times(1).Return(nil, nil),
				g.EXPECT().Get().Times(1).Return(1, nil),
			)

			var w = New(&Config{
				Rater:  Time(100 * time.Millisecond),
				Getter: g,
				Diff:   true,
				Debug:  true,
			})
			w.Start()

			time.Sleep(200 * time.Millisecond)
			var timer = time.NewTimer(200 * time.Millisecond)
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
