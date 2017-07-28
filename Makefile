.PHONY: test test-scheduler test-executor test-race bench scheduler executor build

test:
	@go test -timeout 1m -cover ./...

test-scheduler:
	@go test -timeout 1m -cover sprint/scheduler/...

test-executor:
	@go test -timeout 1m -cover sprint/executor/...

test-race:
	@go test -timeout 1m -race ./...

bench:
	@go test -timeout 1m -bench . ./...

scheduler: test-scheduler
	@go build -o sched sprint/scheduler/main

executor: test-executor
	@go build -o exec sprint/executor/main

build: scheduler executor
