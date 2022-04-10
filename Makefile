IMAGE=foo5aiye
VERSION=v1

.PHONY: all
all: build

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: test
test: tidy
	go test -v cowboy/*.go
	go test -v universe/*.go
	go test -v timetraveler/*.go
	go test -v common/*.go

bin/cowboy-${VERSION}-amd64: tidy test
	go build -o bin/cowboy-${VERSION}-amd64 cowboy/*.go
bin/universe-${VERSION}-amd64: tidy test
	go build -o bin/universe-${VERSION}-amd64 universe/*.go
bin/timetraveler-${VERSION}-amd64: tidy test
	go build -o bin/timetraveler-${VERSION}-amd64 timetraveler/*.go

.PHONY: build
build: tidy test bin/cowboy-${VERSION}-amd64 bin/universe-${VERSION}-amd64 bin/timetraveler-${VERSION}-amd64

.PHONY: image
image:
	docker build -t creckx/${IMAGE}:${VERSION} .
	docker push creckx/${IMAGE}:${VERSION}
