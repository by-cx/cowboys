IMAGE=foo5aiye
VERSION=v1

.PHONY: all
all: build

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: test
test: tidy
	go test -race -v cowboy/*.go
	go test -race -v universe/*.go
	go test -race -v timetraveler/*.go
	go test -race -v common/*.go

bin/cowboy-${VERSION}-amd64: tidy test
	go build -race -o bin/cowboy-${VERSION}-amd64 cowboy/*.go
bin/universe-${VERSION}-amd64: tidy test
	go build -race -o bin/universe-${VERSION}-amd64 universe/*.go
bin/timetraveler-${VERSION}-amd64: tidy test
	go build -race -o bin/timetraveler-${VERSION}-amd64 timetraveler/*.go

.PHONY: build
build: tidy test bin/cowboy-${VERSION}-amd64 bin/universe-${VERSION}-amd64 bin/timetraveler-${VERSION}-amd64

.PHONY: image
image:
	docker build -t creckx/${IMAGE}:${VERSION} .
	docker push creckx/${IMAGE}:${VERSION}
