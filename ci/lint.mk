lint: govet golint

govet:
	go vet ./...

golint:
	golint -set_exit_status ./...
