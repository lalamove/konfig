package konfig

import (
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestWatcher(t *testing.T) {
	t.Run("Start returns nil",
		func(t *testing.T) {
			var w = NopWatcher{}
			require.Nil(t, w.Start())

			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()

			var mockL = NewMockLoader(ctrl)

			var c = New(&Config{})

			c.RegisterLoaderWatcher(
				NewLoaderWatcher(mockL, w),
			)

			c.Watch()
			require.Nil(t, w.Err())
			time.Sleep(100 * time.Millisecond)
		},
	)
}
