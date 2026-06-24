BINARY  := bolt
VERSION := 0.1.0
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION)"

PLATFORMS := \
	darwin/amd64 \
	darwin/arm64 \
	linux/amd64 \
	linux/arm64 \
	windows/amd64

.PHONY: all build install uninstall release clean test lint help

all: build

build:
	go build $(LDFLAGS) -o $(BINARY)$(if $(filter windows,$(OS)),.exe,) .

install: build
	./$(BINARY) install

uninstall:
	./$(BINARY) uninstall

test:
	go test ./...

lint:
	go vet ./...

release: clean
	$(foreach P,$(PLATFORMS), \
		$(eval OS   := $(word 1,$(subst /, ,$(P)))) \
		$(eval ARCH := $(word 2,$(subst /, ,$(P)))) \
		$(eval EXT  := $(if $(filter windows,$(OS)),.exe,)) \
		$(eval OUT  := dist/$(BINARY)_$(VERSION)_$(OS)_$(ARCH)$(EXT)) \
		GOOS=$(OS) GOARCH=$(ARCH) go build $(LDFLAGS) -o $(OUT) . ; \
	)
	cd dist && \
		tar -czf $(BINARY)_$(VERSION)_darwin_amd64.tar.gz  $(BINARY)_$(VERSION)_darwin_amd64  && \
		tar -czf $(BINARY)_$(VERSION)_darwin_arm64.tar.gz  $(BINARY)_$(VERSION)_darwin_arm64  && \
		tar -czf $(BINARY)_$(VERSION)_linux_amd64.tar.gz   $(BINARY)_$(VERSION)_linux_amd64   && \
		tar -czf $(BINARY)_$(VERSION)_linux_arm64.tar.gz   $(BINARY)_$(VERSION)_linux_arm64   && \
		zip      $(BINARY)_$(VERSION)_windows_amd64.zip    $(BINARY)_$(VERSION)_windows_amd64.exe
	@echo "release $(VERSION) ready in dist/"

clean:
	rm -rf dist/ $(BINARY) $(BINARY).exe

help:
	@echo "usage: make <target>"
	@echo ""
	@echo "  build      build for current platform"
	@echo "  install    build then install to PATH via bolt install"
	@echo "  uninstall  remove bolt from PATH via bolt uninstall"
	@echo "  release    cross-compile all platforms into dist/"
	@echo "  test       run tests"
	@echo "  lint       run go vet"
	@echo "  clean      remove build artifacts"
