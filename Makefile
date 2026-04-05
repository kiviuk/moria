.PHONY: build clean test cover

build:
	go build -o bin/moria ./cmd/moria

run: build
	./bin/moria

install:
	go install ./cmd/moria

test:
	go test ./...

cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

clean:
	rm -rf bin/
