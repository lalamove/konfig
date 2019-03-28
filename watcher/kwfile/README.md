# File Watcher
File Watcher watches over a file given in the config.

# Usage
```
import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/lalamove/konfig/watcher/kwfile"
)

func main() {
	f, _ := ioutil.TempFile("", "konfig")
	f.Write([]byte(`ABC`))

	defer os.Remove(f.Name())

	var n = kwfile.New(&kwfile.Config{
		Files: []string{f.Name()},
		Rate:  100 * time.Millisecond,
		Debug: true,
	})

	n.Start()

	time.Sleep(100 * time.Millisecond)

	f.Write([]byte(`12345`))

	var timer = time.NewTimer(200 * time.Millisecond)
	select {
	case now := <-timer.C:
		fmt.Println(now)
		break
	case <-n.Done():
		fmt.Println("done!")
		break
	case <-n.Watch():
		fmt.Println("file changed!") // will see this log
		break
	}
}
```
