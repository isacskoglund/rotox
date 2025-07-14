protoc:
	buf generate --include-imports

.PHONY: test
test:
	go test -v -count=1 ./...	