.PHONY: build
build:
	go generate
	mkdir -p dist
	go build -o dist/gurnel

.PHONY: clean
clean:
	-rm -f *_generated.go
	-rm -rf dist

.PHONY: release
release:
	@:$(call check_defined, VERSION, version to release)
	go mod tidy
	git tag $(VERSION)
	goreleaser

check_defined = \
    $(strip $(foreach 1,$1, \
        $(call __check_defined,$1,$(strip $(value 2)))))
__check_defined = \
    $(if $(value $1),, \
      $(error Undefined $1$(if $2, ($2))))
