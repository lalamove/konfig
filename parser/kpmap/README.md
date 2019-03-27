# Map Parser
kpmap package provides a function `PopFlatten` for populating a konfig.Store by flatteing a map[string]interface{}, which is used by json/toml/yaml parser.

# Usage
```
func main() {
	var m = map[string]interface{}{
		"test": map[string]interface{}{
			"foo": "bar",
		},
		"testIface": map[interface{}]interface{}{
			1: "bar",
			"testIface": map[interface{}]interface{}{
				"foo": "bar",
			},
			"test": map[string]interface{}{
				"foo": "bar",
			},
		},
	}
	var v = konfig.Values{}
	kpmap.PopFlatten(m, v)

	fmt.Println(v) // map[test.foo:bar testIface.1:bar testIface.test.foo:bar testIface.testIface.foo:bar]
}
```
