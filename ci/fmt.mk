fmt: modtidy gofmt prettier
ifdef CI
	if [[ $$(git ls-files --other --modified --exclude-standard) != "" ]]; then
	  echo "Files need generation or are formatted incorrectly:"
	  git -c color.ui=always status | grep --color=no '\e\[31m'
	  echo "Please run the following locally:"
	  echo "  make fmt"
	  exit 1
	fi
endif

modtidy: gen
	go mod tidy

gofmt: gen
	# gofumpt v0.7.0 requires Go 1.22 or later.
	go run mvdan.cc/gofumpt@v0.6.0 -w .

prettier:
	npx prettier --write --print-width=120 --no-semi --trailing-comma=all --loglevel=warn $$(git ls-files "*.yml")

gen:
	go generate ./...
