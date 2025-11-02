<!-- OPENSPEC:START -->
# OpenSpec Instructions

These instructions are for AI assistants working in this project.

Always open `@/openspec/AGENTS.md` when the request:
- Mentions planning or proposals (words like proposal, spec, change, plan)
- Introduces new capabilities, breaking changes, architecture shifts, or big performance/security work
- Sounds ambiguous and you need the authoritative spec before coding

Use `@/openspec/AGENTS.md` to learn:
- How to create and apply change proposals
- Spec format and conventions
- Project structure and guidelines

Keep this managed block so 'openspec update' can refresh the instructions.

<!-- OPENSPEC:END -->

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
