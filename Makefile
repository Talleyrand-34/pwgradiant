.PHONY: all build run clean test fmt install-deps server init

# Binary name
BINARY_NAME=pwdmgr
DB_PATH=passwords.db
ENCRYPTION_KEY=change-this-key-in-production

all: build

# Install dependencies
install-deps:
	go mod download
	go mod tidy

# Build the application
build: install-deps
	go build -o $(BINARY_NAME) .

# Build for multiple platforms
build-all: install-deps
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux .
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-macos .
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME).exe .

# Run the application (CLI help)
run: build
	./$(BINARY_NAME) --help

# Initialize the database
init: build
	./$(BINARY_NAME) init --database $(DB_PATH) --key "$(ENCRYPTION_KEY)"

# Start the API server
server: build
	./$(BINARY_NAME) server --database $(DB_PATH) --key "$(ENCRYPTION_KEY)" --port 8080

# Add a sample account
add-sample: build
	./$(BINARY_NAME) account add \
		--database $(DB_PATH) \
		--key "$(ENCRYPTION_KEY)" \
		--email "admin@example.com" \
		--user "admin" \
		--password "SecurePassword123!" \
		--url "https://example.com" \
		--notes "Sample account"

# List accounts
list: build
	./$(BINARY_NAME) account list --database $(DB_PATH) --key "$(ENCRYPTION_KEY)"

# Run tests
test:
	go test -v ./...

# Format code
fmt:
	go fmt ./...

# Clean build artifacts
clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-linux
	rm -f $(BINARY_NAME)-macos
	rm -f $(BINARY_NAME).exe

# Clean database (be careful!)
clean-db:
	rm -f $(DB_PATH)

# Full clean (build artifacts and database)
clean-all: clean clean-db

# Help
help:
	@echo "Available targets:"
	@echo "  make build         - Build the application"
	@echo "  make build-all     - Build for multiple platforms"
	@echo "  make init          - Initialize the database"
	@echo "  make server        - Start the API server"
	@echo "  make add-sample    - Add a sample account"
	@echo "  make list          - List all accounts"
	@echo "  make test          - Run tests"
	@echo "  make fmt           - Format code"
	@echo "  make clean         - Remove build artifacts"
	@echo "  make clean-db      - Remove database"
	@echo "  make clean-all     - Remove everything"
	@echo ""
	@echo "Environment variables:"
	@echo "  DB_PATH           - Database file path (default: passwords.db)"
	@echo "  ENCRYPTION_KEY    - Database encryption key (default: change-this-key-in-production)"
