# Env Loader
Env loader loads environment variables in a konfig.Store

# Usage

Basic usage loading all environment variables
```go
envLoader := klenv.New(&klenv.Config{})
```

Loading specific variables
```go
envLoader := klenv.New(&klenv.Config{
	Vars: []string{
		"DEBUG",
		"PORT",
	},
})
```

Loading specific variables if key matches regexp
```go
envLoader := klenv.New(&klenv.Config{
	Regexp: "^APP_.*"
})
```

With a replacer and a Prefix for keys
```go
envLoader := klenv.New(&klenv.Config{
	Prefix: "config.",
	Replacer: nstrings.ReplacerToLower,
})
```
