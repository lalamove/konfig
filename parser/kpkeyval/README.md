# KV Parser
KV parser is a key value parser to parse an io.Reader's content of key/values with a configurable separator and add it into a konfig.Store.

Ex:
```
bar=foo
foo=bar
```
Will add the following key/value to the config
```
"foo" => "bar"
"bar" => "foo"
```

# Usage
```
func main() {
	var v = konfig.Values{}
	var p = kpkeyval.New(&kpkeyval.Config{})

	p.Parse(strings.NewReader(
		"bar=foo\nfoo=bar",
	), v)

	fmt.Println(v) // map[bar:foo foo:bar]
}
```
