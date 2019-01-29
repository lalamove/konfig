test:
	go test $(shell go list ./... | grep -v /examples/ ) -covermode=count

test-race:
	go test -race $(shell go list ./... | grep -v /examples/ )

coverage:
	GO111MODULE=off go test ./... -cover -covermode=count -coverprofile=cover.out; GO111MODULE=off go tool cover -func cover.out;

coverage-html:
	GO111MODULE=off go test ./... -cover -covermode=count -coverprofile=cover.out; GO111MODULE=off go tool cover -html=cover.out;

benchmarks:
	cd benchmarks && go test -bench . && cd ../

lint: 
	golint -set_exit_status $(shell (go list ./... | grep -v /vendor/))

mocks:
	mockgen -source ./loader.go -package mocks > ./mocks/loader_mock.go
	mockgen -source ./watcher.go -package mocks > ./mocks/watcher_mock.go
	mockgen -source ./loader.go -package konfig > ./loader_mock_test.go
	mockgen -source ./watcher.go -package konfig > ./watcher_mock_test.go
	mockgen -source ./loader/klvault/authprovider.go -package mocks > ./mocks/authprovider_mock.go	
	mockgen -source ./loader/klvault/vaultloader.go -package mocks LogicalClient > ./mocks/logicalclient_mock.go	
	mockgen -source ./parser/parser.go -package mocks Parser > ./mocks/parser_mock.go
	mockgen -source ./loader/klhttp/httploader.go -package mocks Client > ./mocks/client_mock.go
	mockgen -package mocks go.etcd.io/etcd/clientv3 KV > ./mocks/kv_mock.go
	mockgen -package mocks github.com/lalamove/nui/ncontext Contexter > ./mocks/contexter_mock.go
	mockgen -source ./parser/parser.go -package mocks Parser > ./mocks/parser_mock.go
	mockgen -source ./loader/klconsul/consulloader.go -package mocks ConsulKV > ./mocks/consulkv_mock.go

.PHONY: test test-race coverage coverage-html lint benchmarks mocks
