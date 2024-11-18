server:
	go run main.go

test:
	go test -v -cover -short -race ./...

load_test:
	rm -rf reports
	mkdir reports
	go test -v -cover -race -run TestLoad ./...

clear_cache:
	go clean -testcache

.PHONY: server test clear_cache