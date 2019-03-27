# Consul loader
Loads config from consul KV.


# Usage

Basic usage loading keys and using result as string with watcher
```go
etcdLoader := klconsul.New(&klconsul.Config{
	Client: consulClient, // from github.com/hashicorp/consul/api package
	Keys: []Key{
		{
			Key: "foo",
		},
	},
	Watch: true,
})
```

Loading keys and JSON parser
```go
consulLoader := klconsul.New(&klconsul.Config{
	Client: consulClient, // from github.com/hashicorp/consul/api package
	Keys: []Key{
		{
			Key: "foo",
			Parser: kpjson.Parser,
		},
	},
	Watch: true,
})
```

# Strict mode
If strict mode is enabled, a key defined in the config but missing in consul will trigger an error.
