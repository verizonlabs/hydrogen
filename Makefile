.PHONY: test test-race bench scheduler executor

test:
	@go test -cover ./...

test-race:
	@go test -race ./...

bench:
	@go test -bench . ./...

scheduler:
	@go build -o sched sprint/scheduler/main

executor:
	@go build -o exec sprint/executor/main
