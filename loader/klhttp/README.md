# HTTP Loader
Loads config from a source over HTTP

# Usage

Basic usage with a json source and a poll watcher
```go
httpLoader := klhttp.New(&klhttp.Config{
	Sources: []Source{
		{
			URL: "https://konfig.io/config.json",
			Method: "GET",
			Parser: kpjson.Parser,
		},
	},
	Watch: true,
	Rater: kwpoll.Time(10 * time.Second), // Rater is the rater for the poll watcher
})
```
