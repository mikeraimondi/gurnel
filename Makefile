.PHONY: check
check: tidy test lint

.PHONY: build
build: pre
	mkdir -p dist
	go build -o dist/gurnel cmd/gurnel/main.go

.PHONY: pre
pre:
	go mod download
	./scripts/bindata.sh

.PHONY: lint
lint:
	golangci-lint run -c build/.golangci.yml

.PHONY: test
test:
	go test ./internal/gurnel/... -v -race

.PHONY: clean
clean:
	rm -rf dist

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: release
release: check clean
	@:$(call check_defined, VERSION, version to release)
	git tag $(VERSION)
	goreleaser release --config=build/package/.goreleaser.yml

.PHONY: publish
publish: pre release clean

check_defined = \
    $(strip $(foreach 1,$1, \
        $(call __check_defined,$1,$(strip $(value 2)))))
__check_defined = \
    $(if $(value $1),, \
      $(error Undefined $1$(if $2, ($2))))
