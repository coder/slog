lint: govet golint

govet:
	go vet -composites=false ./...

golint:
	golint -set_exit_status ./...
