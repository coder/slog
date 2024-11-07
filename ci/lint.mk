lint: govet golint

govet:
	go vet ./...

golint:
	# Pin golang.org/x/tools, the go.mod of v0.25.0 is incompatible with Go 1.19.
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v0.24.0 run .
