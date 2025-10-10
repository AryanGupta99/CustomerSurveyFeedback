# Makefile for Customer Survey Application

# Define the Go binary name
BINARY_NAME=customer-survey

# Define the Go build command
build:
	go build -o $(BINARY_NAME) ./cmd/survey/main.go

# Define the clean command to remove the binary
clean:
	rm -f $(BINARY_NAME)

# Define the run command to build and run the application
run: build
	./$(BINARY_NAME)

# Define the install command to install the application
install: build
	cp $(BINARY_NAME) /usr/local/bin/

# Define the help command to display available commands
help:
	@echo "Makefile commands:"
	@echo "  build   - Build the application"
	@echo "  clean   - Remove the built binary"
	@echo "  run     - Build and run the application"
	@echo "  install  - Install the application"
	@echo "  help    - Display this help message"