lint: govet golint

govet:
	go vet ./...

golint:
	# golangci-lint newer than v1.55.2 is not compatible with Go 1.20 when using go run.
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2 run .
