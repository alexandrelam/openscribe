Use the makefile to execute those commands 

```makefile
help:
	@echo "OpenScribe Makefile Commands:"
	@echo "  make build    - Build the binary"
	@echo "  make install  - Install the binary to GOPATH/bin"
	@echo "  make run      - Build and run the application"
	@echo "  make clean    - Remove build artifacts"
	@echo "  make test     - Run tests"
	@echo "  make deps     - Download and tidy dependencies"
	@echo "  make fmt      - Format code"
	@echo "  make help     - Display this help message"
```