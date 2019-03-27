# Etcd Loader
Loads configs from Etcd into konfig.Store

# Usage

Basic usage loading keys and using result as string with watcher
```go
etcdLoader := kletcd.New(&kletc.Config{
	Client: etcdClient, // from go.etcd.io/etcd/clientv3 package
	Keys: []Key{
		{
			Key: "foo/bar",
		},
	},
	Watch: true,
})
```

Loading keys and JSON parser
```go
etcdLoader := kletcd.New(&kletc.Config{
	Client: etcdClient, // from go.etcd.io/etcd/clientv3 package
	Keys: []Key{
		{
			Key: "foo/bar",
			Parser: kpjson.Parser,
		},
	},
	Watch: true,
})
```
