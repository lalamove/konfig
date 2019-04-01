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

Loading value as string slice
```go
os.Setenv("APP_VAR1", "value1,value2") // will be loaded as string slice
os.Setenv("APP_VAR2", "value") // will be loaded as string

envLoader := klenv.New(&klenv.Config{
	Regexp: "^APP_.*",
	SliceSeparator: ",",
})

...

fmt.Printf("%+v\n", store.Get("VAR1")) // Output: []string{"value1","value2}
fmt.Printf("%+v\n", store.Get("VAR2")) // Output: value
```
