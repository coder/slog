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
	go run mvdan.cc/gofumpt@latest -w .

prettier:
	npx prettier --write --print-width=120 --no-semi --trailing-comma=all --loglevel=warn $$(git ls-files "*.yml")

gen:
	go generate ./...
