# File Loader
File loader loads config from files

# Usage

Basic usage with files and json parser and a watcher
```go
fileLoader := klfile.New(&klfile.Config{
	Files: []File{
		{
			Path: "./config.json",
			Parser: kpjson.Parser,
		},
	},
	Watch: true,
	Rate: 1 * time.Second, // Rate for the polling watching the file changes
})
```

Simplified syntax:
```go
fileLoader := klfile.
	NewFileLoader("config-files", kpjson.Parser, "file1.json", "file2.json").
	WithWatcher()
```
