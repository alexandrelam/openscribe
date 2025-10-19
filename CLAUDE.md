Use the makefile to execute those commands

**Note:** OpenScribe only supports macOS Apple Silicon (M1/M2/M3/M4). Intel Macs are not supported.

```makefile
help:
	@echo "OpenScribe Makefile Commands:"
	@echo "  make build               - Build the binary"
	@echo "  make build-darwin-arm64  - Build for macOS ARM64 (Apple Silicon)"
	@echo "  make build-all           - Build for Apple Silicon"
	@echo "  make install             - Install the binary to GOPATH/bin"
	@echo "  make run                 - Build and run the application"
	@echo "  make clean               - Remove build artifacts"
	@echo "  make test                - Run tests"
	@echo "  make deps                - Download and tidy dependencies"
	@echo "  make fmt                 - Format code"
	@echo "  make lint                - Run linter"
	@echo "  make help                - Display this help message"
```
