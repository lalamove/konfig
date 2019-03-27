# Flag Loader
Loads config values from command line flags

# Usage

Basic usage with command line FlagSet
```go
flagLoader := klflag.New(&klflag.Config{})
```

With a nstrings.Replacer for keys
```go
flagLoader := klflag.New(&klflag.Config{
	Replacer: strings.NewReplacer(".", "-")
})
```
