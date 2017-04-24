.PHONY: test test-race bench

test:
	@go test -cover ./...

test-race:
	@go test -race ./...

bench:
	@go test -bench . ./...
