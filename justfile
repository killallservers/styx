# Styx development tasks
# Run with: just <task>
# List all tasks: just -l

@default:
    just --list

# Build the styx binary
build:
    go build -o bin/styx ./cmd/styx

# Run all tests
test:
    go test -v ./...

# Run tests with coverage
coverage:
    go test -cover ./...

# Format all Go code
fmt:
    go fmt ./...

# Run static analysis
vet:
    go vet ./...

# Run benchmarks (all packages)
bench:
    go test -bench=. -benchmem ./pkg/...

# Run benchmarks with comparison to baseline
bench-compare:
    @echo "Running benchmarks..."
    go test -bench=. -benchmem ./pkg/... | tee /tmp/bench.txt
    @echo ""
    @echo "✓ Benchmark results saved to /tmp/bench.txt"

# Remove build artifacts
clean:
    rm -f styx
    go clean

# Download and tidy dependencies
deps:
    go mod tidy
    go mod download

# Show help for styx
help:
    ./styx --help

# Show version
version:
    ./styx --version

# Install styx install command help
help-install:
    ./styx install --help

# Run install command (stub)
install:
    ./styx install

# Full test + lint pipeline
check: fmt vet test
    @echo "✓ All checks passed"

# Build and run tests
dev: build test fmt
    @echo "✓ Development build ready"

# Generate bash completion script
completion-bash: build
    ./styx completion bash

# Generate zsh completion script
completion-zsh: build
    ./styx completion zsh

# Generate fish completion script
completion-fish: build
    ./styx completion fish

# Show how to install shell completion
install-completion: build
    @echo "To install shell completion:"
    @echo ""
    @echo "Bash:"
    @echo "  styx completion bash >> ~/.bashrc"
    @echo "  source ~/.bashrc"
    @echo ""
    @echo "Zsh:"
    @echo "  styx completion zsh >> ~/.zshrc"
    @echo "  exec zsh"
    @echo ""
    @echo "Fish:"
    @echo "  mkdir -p ~/.config/fish/conf.d"
    @echo "  styx completion fish > ~/.config/fish/conf.d/styx.fish"

# Install man pages to /usr/local/man
install-man:
    @mkdir -p /usr/local/man/man1
    @cp man/styx.1 /usr/local/man/man1/
    @cp man/styx-*.1 /usr/local/man/man1/
    @echo "✓ Man pages installed to /usr/local/man/man1"
    @echo ""
    @echo "View with:"
    @echo "  man styx"
    @echo "  man styx-env"
    @echo "  man styx-init"
