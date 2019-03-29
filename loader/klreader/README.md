# Reader Loader
Loads config from an io.Reader

# Usage
```go
readerLoader := klreader.New(&klreader.Config{
    Parser: kpjson.Parser,
    Reader: strings.NewReader(`{"foo":"bar"}`),
})
``` 
