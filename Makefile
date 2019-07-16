.PHONY: build
build: compile post

.PHONY: pre
pre:
	go mod download
	./scripts/bindata.sh

.PHONY: compile
compile: pre
	mkdir -p dist
	go build -o dist/gurnel cmd/gurnel/main.go

.PHONY: post
post:
	./scripts/bindata_debug.sh

.PHONY: lint
lint:
	golangci-lint run -c build/.golangci.yml

.PHONY: test
test:
	go test ./internal/gurnel/... -v -race

.PHONY: clean
clean:
	rm -rf dist

.PHONY: release
release: test lint clean
	@:$(call check_defined, VERSION, version to release)
	go mod tidy
	git tag $(VERSION)
	goreleaser release --config=build/package/.goreleaser.yml

.PHONY: publish
publish: pre release post clean

check_defined = \
    $(strip $(foreach 1,$1, \
        $(call __check_defined,$1,$(strip $(value 2)))))
__check_defined = \
    $(if $(value $1),, \
      $(error Undefined $1$(if $2, ($2))))
