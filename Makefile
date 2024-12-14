NAME    := sake
PACKAGE := github.com/alajmo/$(NAME)
DATE    := $(shell date +%FT%T%Z)
GIT     := $(shell [ -d .git ] && git rev-parse --short HEAD)
VERSION := v0.15.1

default: build

tidy:
	go get -u && go mod tidy

gofmt:
	go fmt ./cmd/***.go
	go fmt ./core/***.go
	go fmt ./core/dao/***.go
	go fmt ./core/run/***.go
	go fmt ./core/print/***.go
	go fmt ./test/integration/***.go

lint:
	golangci-lint run ./cmd/... ./core/... ./test/...
	deadcode .

benchmark:
	cd test && ./benchmark.sh

benchmark-save:
	cd test && ./benchmark.sh --save

test:
	# Unit tests
	go test -v ./core/...

	# Integration tests
	cd ./test && docker-compose up -d
	go test -v ./test/integration/... -count=5 -clean
	cd ./test && docker-compose down

unit-test:
	go test -v ./core/...

integration-test:
	go test -v ./test/integration/... -clean

update-golden-files:
	go test ./test/integration/... -update

mock-ssh:
	cd ./test && docker-compose up

mock-performance-ssh:
	cd ./test && docker-compose -f docker-compose-performance.yaml up

build:
	CGO_ENABLED=0 go build \
	-ldflags "-s -w -X '${PACKAGE}/cmd.version=${VERSION}' -X '${PACKAGE}/cmd.commit=${GIT}' -X '${PACKAGE}/cmd.date=${DATE}'" \
	-a -tags netgo -o dist/${NAME} main.go

build-all:
	goreleaser release --skip-publish --rm-dist --snapshot

build-and-link:
	go build \
		-ldflags "-w -X '${PACKAGE}/cmd.version=${VERSION}' -X '${PACKAGE}/cmd.commit=${GIT}' -X '${PACKAGE}/cmd.date=${DATE}'" \
		-a -tags netgo -o dist/${NAME} main.go
	cp ./dist/sake ~/.local/bin/sake

gen-man:
	go run -ldflags="-X 'github.com/alajmo/sake/cmd.buildMode=man' -X '${PACKAGE}/cmd.version=${VERSION}' -X '${PACKAGE}/cmd.commit=${GIT}' -X '${PACKAGE}/cmd.date=${DATE}'" ./main.go gen-docs

release:
	git tag ${VERSION} && git push origin ${VERSION}

clean:
	$(RM) -r dist target

.PHONY: tidy gofmt lint benchmark benchmark-save test unit-test integration-test update-golden-files mock-ssh build build-all build-and-link gen-man release clean
