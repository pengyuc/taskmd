.PHONY: build run clean install install-full install-dev install-dev-full test e2e lint lint-fix tidy check web-build web-embed build-full sync-spec mcpb mcpb-all

# Build the binary
build:
	CGO_ENABLED=0 go build -ldflags "-s -w -extldflags '-static'" -o taskmd ./cmd/taskmd

# Run the application
run:
	CGO_ENABLED=0 go run ./cmd/taskmd

# Clean build artifacts
clean:
	rm -f taskmd
	rm -rf internal/web/static/dist

# Install the binary to $GOPATH/bin
install:
	CGO_ENABLED=0 go install -ldflags "-s -w -extldflags '-static'" ./cmd/taskmd

# Install with embedded web assets
install-full: web-embed
	CGO_ENABLED=0 go install -tags embed_web -ldflags "-s -w -extldflags '-static'" ./cmd/taskmd

# ldflags for injecting git info into dev builds
DEV_COMMIT := $(shell git rev-parse --short HEAD)
DEV_DIRTY  := $(shell git diff --quiet 2>/dev/null && echo false || echo true)
DEV_LDFLAGS := -s -w -extldflags '-static' \
               -X 'github.com/driangle/taskmd/apps/cli/internal/cli.GitCommit=$(DEV_COMMIT)' \
               -X 'github.com/driangle/taskmd/apps/cli/internal/cli.GitDirty=$(DEV_DIRTY)'

# Install as taskmd-dev (keeps Homebrew version available)
install-dev:
	CGO_ENABLED=0 go build -ldflags "$(DEV_LDFLAGS)" -o ~/bin/taskmd-dev ./cmd/taskmd
	@echo "Installed development binary to ~/bin/taskmd-dev"
	@echo "Usage: taskmd-dev [command]"

# Install as taskmd-dev with embedded web assets
install-dev-full: web-embed
	CGO_ENABLED=0 go build -tags embed_web -ldflags "$(DEV_LDFLAGS)" -o ~/bin/taskmd-dev ./cmd/taskmd
	@echo "Installed development binary (with web) to ~/bin/taskmd-dev"
	@echo "Usage: taskmd-dev [command]"

# Run unit and integration tests (excludes e2e)
test:
	go test ./...

# Run end-to-end tests (builds binary and tests as subprocess)
e2e:
	go test -tags e2e -count=1 -v ./internal/e2e/...

GOLANGCI_LINT := $(shell command -v golangci-lint 2>/dev/null || echo $(shell go env GOPATH)/bin/golangci-lint)

# Run linter
lint:
	$(GOLANGCI_LINT) run

# Run linter with auto-fix
lint-fix:
	$(GOLANGCI_LINT) run --fix

# Run go mod tidy to clean up dependencies
tidy:
	go mod tidy

# Run all checks (test, lint, vet)
check: test lint
	go vet ./...

# Build the web SPA
web-build:
	cd ../../apps/web && pnpm install && pnpm build

# Copy built SPA into Go-embeddable location
web-embed: web-build
	rm -rf internal/web/static/dist
	mkdir -p internal/web/static
	cp -r ../../apps/web/dist internal/web/static/dist

# Build Go binary with embedded web assets
build-full: web-embed
	CGO_ENABLED=0 go build -tags embed_web -ldflags "-s -w -extldflags '-static'" -o taskmd ./cmd/taskmd

# Sync spec copies from the canonical source (docs/taskmd_specification.md).
# Run this after editing the spec.
sync-spec:
	cp ../../docs/taskmd_specification.md internal/cli/templates/TASKMD_SPEC.md
	cp ../../docs/taskmd_specification.md ../../apps/docs/reference/specification.md
	sed 's|\./taskmd_specification\.md|./specification.md|g' ../../docs/taskmd_operations.md > ../../apps/docs/reference/taskmd_operations.md
	@echo "Synced spec to embedded template and docs site"

# Build MCPB bundle for current platform
mcpb: build
	../../scripts/build-mcpb.sh ./taskmd "$$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo 0.0.0-dev)" $$(go env GOOS) $$(go env GOARCH) .
	@echo "MCPB bundle created in current directory"

# Build MCPB bundles for all platforms
mcpb-all:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w -extldflags '-static'" -o taskmd-darwin-amd64 ./cmd/taskmd
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w -extldflags '-static'" -o taskmd-darwin-arm64 ./cmd/taskmd
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -extldflags '-static'" -o taskmd-linux-amd64 ./cmd/taskmd
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-s -w -extldflags '-static'" -o taskmd-linux-arm64 ./cmd/taskmd
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -extldflags '-static'" -o taskmd-windows-amd64.exe ./cmd/taskmd
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -ldflags "-s -w -extldflags '-static'" -o taskmd-windows-arm64.exe ./cmd/taskmd
	../../scripts/build-mcpb.sh ./taskmd-darwin-amd64 "$$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo 0.0.0-dev)" darwin amd64 .
	../../scripts/build-mcpb.sh ./taskmd-darwin-arm64 "$$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo 0.0.0-dev)" darwin arm64 .
	../../scripts/build-mcpb.sh ./taskmd-linux-amd64 "$$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo 0.0.0-dev)" linux amd64 .
	../../scripts/build-mcpb.sh ./taskmd-linux-arm64 "$$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo 0.0.0-dev)" linux arm64 .
	../../scripts/build-mcpb.sh ./taskmd-windows-amd64.exe "$$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo 0.0.0-dev)" windows amd64 .
	../../scripts/build-mcpb.sh ./taskmd-windows-arm64.exe "$$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo 0.0.0-dev)" windows arm64 .
	@echo "MCPB bundles created for all platforms"

# Build for multiple platforms
build-all:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w -extldflags '-static'" -o taskmd-darwin-amd64 ./cmd/taskmd
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w -extldflags '-static'" -o taskmd-darwin-arm64 ./cmd/taskmd
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -extldflags '-static'" -o taskmd-linux-amd64 ./cmd/taskmd
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -extldflags '-static'" -o taskmd-windows-amd64.exe ./cmd/taskmd
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -ldflags "-s -w -extldflags '-static'" -o taskmd-windows-arm64.exe ./cmd/taskmd
