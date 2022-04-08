.PHONY: tidy
tidy:
	go mod tidy

.PHONY: test
test: tidy
	go test cowboy/*.go
	go test time/*.go
	go test timetraveler/*.go


.PHONY: build
build: tidy
	go build -o bin/cowboy cowboy/*.go
