# Poll Watcher
Poll Watcher is a konfig.Watcher that sends events every x time given in the konfig.

# Usage
```
import (
	"fmt"
	"os"
	"time"

	"github.com/lalamove/konfig"
	"github.com/lalamove/konfig/loader/klenv"
	"github.com/lalamove/konfig/watcher/kwpoll"
)

func main() {
	os.Setenv("foo", "bar")

	var l = klenv.New(&klenv.Config{
		Vars: []string{
			"foo",
		},
	})

	var v = konfig.Values{}
	l.Load(v)

	var w = kwpoll.New(&kwpoll.Config{
		Rater:     kwpoll.Time(100 * time.Millisecond),
		Loader:    l,
		Diff:      true,
		Debug:     true,
		InitValue: v,
	})
	w.Start()

	var timer = time.NewTimer(200 * time.Millisecond)
	var watched int

	os.Setenv("foo", "baz") // change the value

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

	fmt.Println(watched) // will output 1
}
```
