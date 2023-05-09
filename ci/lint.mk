lint: govet golint

govet:
	go vet ./...

golint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run .
