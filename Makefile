NAME    := sake
PACKAGE := github.com/alajmo/$(NAME)
DATE    := $(shell date +%FT%T%Z)
GIT     := $(shell [ -d .git ] && git rev-parse --short HEAD)
VERSION := v0.1.6

default: build

tidy:
	go get -u && go mod tidy

gofmt:
	go fmt ./cmd/***.go
	go fmt ./core/***.go
	go fmt ./core/dao/***.go
	go fmt ./core/run/***.go
	go fmt ./core/print/***.go

lint:
	golangci-lint run ./cmd/... ./core/... ./test/...

test:
	go test ./core/dao/***
	cd ./test && docker-compose up -d
	go test -v ./test/integration/... -count=5 -clean
	cd ./test && docker-compose down

unit-test:
	go test ./core/dao/***

integration-test:
	go test -v ./test/integration/... -clean

update-golden-files:
	go test ./test/integration/... -update

mock-ssh:
	cd ./test && docker-compose up

build:
	CGO_ENABLED=0 go build \
	-ldflags "-s -w -X '${PACKAGE}/cmd.version=${VERSION}' -X '${PACKAGE}/cmd.commit=${GIT}' -X '${PACKAGE}/cmd.date=${DATE}'" \
	-a -tags netgo -o dist/${NAME} main.go

build-all:
	goreleaser release --skip-publish --rm-dist --snapshot

build-man:
	go run -ldflags="-X 'github.com/alajmo/sake/cmd.buildMode=man' -X '${PACKAGE}/cmd.version=${VERSION}' -X '${PACKAGE}/cmd.commit=${GIT}' -X '${PACKAGE}/cmd.date=${DATE}'" ./main.go gen-docs

build-and-link:
	go build \
		-ldflags "-w -X '${PACKAGE}/cmd.version=${VERSION}' -X '${PACKAGE}/cmd.commit=${GIT}' -X '${PACKAGE}/cmd.date=${DATE}'" \
		-a -tags netgo -o dist/${NAME} main.go
	cp ./dist/sake ~/.local/bin/sake

release:
	git tag ${VERSION} && git push origin ${VERSION}

clean:
	$(RM) -r dist target

.PHONY: tidy gofmt lint test unit-test integration-test update-golden-files mock-ssh build build-all build-man build-and-link release clean
