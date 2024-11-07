lint: govet golint

govet:
	go vet ./...

golint:
	# golangci-lint v1.60.1 is not compatible with Go 1.20.
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.60.0 run .
