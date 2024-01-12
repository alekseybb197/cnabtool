PROJECT=$(shell basename "$(PWD)")
APPVERS=0.1.0
GITCOMMIT=$(shell git rev-parse --short HEAD)
GOFLAGS=-trimpath -ldflags "-w -s -X 'cnabtool/cmd.Version=${APPVERS}' -X 'cnabtool/cmd.Commit=${GITCOMMIT}'"
GO111MODULE=on
CGO_ENABLED=0

default: build

.PHONY: build
build:
	go build ${GOFLAGS} -o bin/${PROJECT}

.PHONY: run
run:
	go run ${PROJECT}.go

.PHONY: dist
dist:
	# FreeBDS
	#GOOS=freebsd GOARCH=amd64 go build ${GOFLAGS} -o dist/${PROJECT}-freebsd-amd64
	# MacOS
	GOOS=darwin GOARCH=amd64 go build ${GOFLAGS} -o dist/${PROJECT}-darwin-amd64
	# Linux
	GOOS=linux GOARCH=amd64 go build ${GOFLAGS} -o dist/${PROJECT}-linux-amd64
	# Windows
	GOOS=windows GOARCH=amd64 go build ${GOFLAGS} -o dist/${PROJECT}-windows-amd64.exe

.PHONY: clean
clean:
	go clean
	rm -rf bin
	rm -rf dist

PHONY: fmt
fmt:
	gofumpt -w -s  .

PHONY: test
test:
	go test ./...

PHONY: lint
lint:
	golangci-lint run -c .golang-ci.yml
