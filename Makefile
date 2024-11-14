server:
	go run main.go

test:
	go test -v -cover -short -race ./...

clear_cache:
	go clean -testcache

.PHONY: server test clear_cache