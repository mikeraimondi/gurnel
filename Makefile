.PHONY: build
build: generate
	mkdir -p dist
	go build -o dist/gurnel cmd/gurnel/main.go

.PHONY: generate
generate:
	go generate github.com/mikeraimondi/gurnel/internal/gurnel

.PHONY: lint
lint:
	golangci-lint run -c build/.golangci.yml

.PHONY: clean
clean:
	find . -type f -name '*_generated.go' -delete
	rm -rf dist

.PHONY: release
release: lint clean
	@:$(call check_defined, VERSION, version to release)
	go mod tidy
	git tag $(VERSION)

.PHONY: publish
publish: release
	goreleaser release --config=build/package/.goreleaser.yml
	make clean

check_defined = \
    $(strip $(foreach 1,$1, \
        $(call __check_defined,$1,$(strip $(value 2)))))
__check_defined = \
    $(if $(value $1),, \
      $(error Undefined $1$(if $2, ($2))))
