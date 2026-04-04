.PHONY: build clean test

build:
	go build -o bin/pwdgen ./cmd/pwdgen

run: build
	./bin/pwdgen

install:
	go install ./cmd

test:
	go test ./...

clean:
	rm -rf bin/
