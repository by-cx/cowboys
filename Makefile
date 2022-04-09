.PHONY: all
all: build

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: test
test: tidy
	go test -v cowboy/*.go
	go test -v universe/*.go
	# go test -v timetraveler/*.go
	go test -v common/*.go


.PHONY: build
build: tidy test
	go build -o bin/cowboy cowboy/*.go
