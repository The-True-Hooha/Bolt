BINARY  := bolt
VERSION := 0.1.0
LDFLAGS := -ldflags="-s -w -X github.com/The-True-Hooha/Bolt/internal/cmd.version=$(VERSION)"

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

release:
	@echo "releases are built via GitHub Actions on git tag push"
	@echo "  git tag v$(VERSION) && git push origin v$(VERSION)"

clean:
	go clean
	$(if $(filter Windows_NT,$(OS)),powershell -Command "Remove-Item -Recurse -Force -ErrorAction SilentlyContinue dist,$(BINARY).exe",rm -rf dist/ $(BINARY))

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
