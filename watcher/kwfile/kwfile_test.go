package kwfile

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWatcher(t *testing.T) {
	t.Run(
		"new panics",
		func(t *testing.T) {
			require.Panics(
				t,
				func() {
					New(&Config{
						Files: []string{"donotexist"},
						Rate:  10 * time.Second,
					})
				},
			)
		},
	)

	t.Run(
		"new with watcher",
		func(t *testing.T) {
			f, err := os.OpenFile("./test", os.O_RDWR|os.O_CREATE, 0755)
			f.Write([]byte(`ABC`))
			require.Nil(t, err)

			defer os.Remove(f.Name())

			var n = New(&Config{
				Files: []string{"./test"},
				Rate:  100 * time.Millisecond,
				Debug: true,
			})

			require.Nil(t, n.Start())

			n.w.Wait()

			time.Sleep(100 * time.Millisecond)

			_, err = f.Write([]byte(`1233454`))
			require.Nil(t, err)

			var timer = time.NewTimer(200 * time.Millisecond)
			var watched bool
			select {
			case <-timer.C:
				break
			case <-n.Watch():
				watched = true
				break
			}

			require.True(t, watched)
		},
	)
}
