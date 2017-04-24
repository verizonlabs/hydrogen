.PHONY: test test-scheduler test-executor test-race bench scheduler executor build

test:
	@go test -cover ./...

test-scheduler:
	@go test -cover sprint/scheduler/...

test-executor:
	@go test -cover sprint/executor/...

test-race:
	@go test -race ./...

bench:
	@go test -bench . ./...

scheduler: test-scheduler
	@go build -o sched sprint/scheduler/main

executor: test-executor
	@go build -o exec sprint/executor/main

build: scheduler executor
