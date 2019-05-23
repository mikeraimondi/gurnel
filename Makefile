.PHONY: build
build:
	go generate
	mkdir -p dist
	go build -o dist/gurnel

.PHONY: clean
clean:
	rm *_generated.go
	rm -rf dist
