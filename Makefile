# Makefile
.EXPORT_ALL_VARIABLES:

GO111MODULE=on
GOPROXY=https://proxy.golang.org,direct
GOSUMDB=sum.golang.org
GOPRIVATE=github.com/cssbruno/gocep

.PHONY: build update compose test

build:
	@echo "########## Building API ..."
	go build -ldflags="-s -w" -o gocep main.go
	#upx gocep
	@echo "build completed"
	@echo "\033[0;33m################ run #####################\033[0m"
	rm -f gocep

update:
	@echo "########## Updating dependencies ..."
	go get -u -t ./...
	go mod tidy
	go test ./...
	@echo "dependency update completed"

compose:
	@echo "########## Running deployment script ..."
	sh deploy.gocep.sh
	@echo "done"

test: 
	go test -race -v ./...
	go test -v -tags musl -covermode=atomic -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
