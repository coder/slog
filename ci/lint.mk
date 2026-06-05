lint: govet golint

govet:
	go vet ./...

golint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8 run .
