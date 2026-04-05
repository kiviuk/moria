.PHONY: build clean test

build:
	go build -o bin/moria ./cmd/moria

run: build
	./bin/moria

install:
	go install ./cmd/moria

test:
	go test ./...

clean:
	rm -rf bin/
