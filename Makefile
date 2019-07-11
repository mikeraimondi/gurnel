.PHONY: build
build:
	go mod download
	mkdir -p dist
	./scripts/bindata.sh
	go build -o dist/gurnel cmd/gurnel/main.go
	./scripts/bindata_debug.sh

.PHONY: lint
lint:
	golangci-lint run -c build/.golangci.yml

.PHONY: clean
clean:
	rm -rf dist

.PHONY: release
release: lint clean
	@:$(call check_defined, VERSION, version to release)
	go mod tidy
	git tag $(VERSION)

.PHONY: publish
publish: release
	go mod download
	./scripts/bindata.sh
	goreleaser release --config=build/package/.goreleaser.yml
	make clean
	./scripts/bindata_debug.sh

check_defined = \
    $(strip $(foreach 1,$1, \
        $(call __check_defined,$1,$(strip $(value 2)))))
__check_defined = \
    $(if $(value $1),, \
      $(error Undefined $1$(if $2, ($2))))
