all: clean setup logservice registryservice libraryservice

logservice:
	go build -o build/logservice ./cmd/logservice

registryservice:
	go build -o build/registryservice ./cmd/registryservice

libraryservice:
	go build -o build/libraryservice ./cmd/libraryservice

.PHONY: all setup

setup:
	mkdir -p build

clean:
	rm -rf build

gofmt:
	find . -name "*.go" | xargs -L1 go fmt
