package kwfile

import (
	"io/ioutil"
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
						Rate:  0, // using default rate
					})
				},
			)
		},
	)

	t.Run(
		"new with watcher",
		func(t *testing.T) {
			f, err := ioutil.TempFile("", "konfig")
			f.Write([]byte(`ABC`))
			require.Nil(t, err)

			defer os.Remove(f.Name())

			var n = New(&Config{
				Files: []string{f.Name()},
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
			case <-n.Done():
				break
			case <-n.Watch():
				watched = true
				break
			}

			require.True(t, watched)
			require.Nil(t, n.Err())
		},
	)

	t.Run(
		"close",
		func(t *testing.T) {
			f, err := ioutil.TempFile("", "konfig")
			f.Write([]byte(`ABC`))
			require.Nil(t, err)

			defer os.Remove(f.Name())

			var n = New(&Config{
				Files: []string{f.Name()},
				Rate:  100 * time.Millisecond,
				Debug: true,
			})

			require.Nil(t, n.Start())
			n.w.Wait()
			time.Sleep(100 * time.Millisecond)

			n.Close()
			<-n.Done()
		},
	)

	t.Run(
		"start panics",
		func(t *testing.T) {
			f, err := ioutil.TempFile("", "konfig")
			f.Write([]byte(`ABC`))
			require.Nil(t, err)

			defer os.Remove(f.Name())

			var n = New(&Config{
				Files: []string{f.Name()},
				Rate:  100 * time.Millisecond,
				Debug: true,
			})

			n.cfg.Rate = 0 // duration is less than 1ns

			n.Start()
		},
	)
}
