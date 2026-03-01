# Password Manager with SQLCipher

A secure password manager using SQLCipher encryption, with both REST API and CLI interfaces.

## Features

- ✅ Encrypted SQLite database using SQLCipher (AES-256)
- ✅ REST API with JSON responses
- ✅ Command-line interface (CLI) using Cobra
- ✅ Account management (CRUD operations)
- ✅ Tag system for organizing accounts
- ✅ Full history tracking for accounts
- ✅ Search functionality
- ✅ Go structs mapped to database with JSON export

## Installation

### Prerequisites

- Go 1.21 or later
- SQLCipher library

#### Install SQLCipher (Ubuntu/Debian)
```bash
sudo apt-get install libsqlcipher-dev
```

#### Install SQLCipher (macOS)
```bash
brew install sqlcipher
```

### Build the application

```bash
go mod download
go build -o pwdmgr
```

## Usage

### Initialize the Database

```bash
./pwdmgr init --key "your-secret-encryption-key"
```

### CLI Commands

All commands require the `--key` flag with your encryption key.

#### Account Management

**Add an account:**
```bash
./pwdmgr account add \
  --key "your-secret-key" \
  --email "john@example.com" \
  --user "john_doe" \
  --password "SecurePass123!" \
  --url "https://example.com" \
  --notes "My example account"
```

**List all accounts:**
```bash
./pwdmgr account list --key "your-secret-key"
```

**Get a specific account:**
```bash
./pwdmgr account get 1 --key "your-secret-key"
```

**Update an account:**
```bash
./pwdmgr account update 1 \
  --key "your-secret-key" \
  --password "NewSecurePass456!"
```

**Delete an account:**
```bash
./pwdmgr account delete 1 --key "your-secret-key"
```

**Search accounts:**
```bash
./pwdmgr account search "example" --key "your-secret-key"
```

**View account history:**
```bash
./pwdmgr account history 1 --key "your-secret-key"
```

#### Tag Management

**Add a tag:**
```bash
./pwdmgr tag add "work" --key "your-secret-key"
```

**List all tags:**
```bash
./pwdmgr tag list --key "your-secret-key"
```

**Delete a tag:**
```bash
./pwdmgr tag delete 1 --key "your-secret-key"
```

#### Tag Association with Accounts

**Add a tag to an account (by ID):**
```bash
./pwdmgr account add-tag 1 2 --key "your-secret-key"
# Adds tag ID 2 to account ID 1
```

**Add a tag to an account (by name - easier!):**
```bash
./pwdmgr account tag 1 work --key "your-secret-key"
# Adds 'work' tag to account ID 1 (creates tag if doesn't exist)
```

**Remove a tag from an account (by ID):**
```bash
./pwdmgr account remove-tag 1 2 --key "your-secret-key"
# Removes tag ID 2 from account ID 1
```

**Remove a tag from an account (by name - easier!):**
```bash
./pwdmgr account untag 1 work --key "your-secret-key"
# Removes 'work' tag from account ID 1
```

**List tags for an account:**
```bash
./pwdmgr account list-tags 1 --key "your-secret-key"
# Shows all tags associated with account ID 1
```

**List all accounts with a specific tag:**
```bash
./pwdmgr account by-tag work --key "your-secret-key"
# Shows all accounts tagged with 'work'
```

#### TOTP Management

**Add TOTP to an account:**
```bash
./pwdmgr totp add 1 JBSWY3DPEHPK3PXP --key "your-secret-key"
# Adds TOTP seed to account ID 1
```

**Generate TOTP code:**
```bash
./pwdmgr totp generate 1 --key "your-secret-key"
# Generates current TOTP code for account ID 1
```

**Verify TOTP code:**
```bash
./pwdmgr totp verify 1 123456 --key "your-secret-key"
# Verifies code 123456 for account ID 1
```

**Delete TOTP from account:**
```bash
./pwdmgr totp delete 1 --key "your-secret-key"
# Removes TOTP from account ID 1
```

### REST API

**Start the API server:**
```bash
./pwdmgr server --key "your-secret-key" --port 8080
```

#### API Endpoints

**Health Check:**
```bash
curl http://localhost:8080/health
```

**Accounts:**

- `POST /api/v1/accounts` - Create account
- `GET /api/v1/accounts` - List all accounts
- `GET /api/v1/accounts/:id` - Get account by ID
- `PUT /api/v1/accounts/:id` - Update account
- `DELETE /api/v1/accounts/:id` - Delete account
- `GET /api/v1/accounts/search?q=query` - Search accounts
- `GET /api/v1/accounts/:id/history` - Get account history
- `POST /api/v1/accounts/:id/tags/:tag_id` - Add tag to account
- `DELETE /api/v1/accounts/:id/tags/:tag_id` - Remove tag from account

**Tags:**

- `POST /api/v1/tags` - Create tag
- `GET /api/v1/tags` - List all tags
- `GET /api/v1/tags/:id` - Get tag by ID
- `PUT /api/v1/tags/:id` - Update tag
- `DELETE /api/v1/tags/:id` - Delete tag

**TOTP:**

- `POST /api/v1/totp` - Create TOTP for account
- `GET /api/v1/totp/account/:account_id` - Get TOTP by account ID
- `PUT /api/v1/totp/:id` - Update TOTP
- `DELETE /api/v1/totp/:id` - Delete TOTP
- `GET /api/v1/totp/:id/history` - Get TOTP history

#### API Examples

**Create an account:**
```bash
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "user": "john_doe",
    "password": "SecurePass123!",
    "url": "https://example.com",
    "notes": "My example account"
  }'
```

**Update an account with change reason:**
```bash
curl -X PUT "http://localhost:8080/api/v1/accounts/1?change_reason=Password%20expired" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "user": "john_doe",
    "password": "NewSecurePass456!",
    "url": "https://example.com",
    "notes": "Updated password"
  }'
```

**Get all accounts:**
```bash
curl http://localhost:8080/api/v1/accounts
```

**Search accounts:**
```bash
curl "http://localhost:8080/api/v1/accounts/search?q=example"
```

**Create a tag:**
```bash
curl -X POST http://localhost:8080/api/v1/tags \
  -H "Content-Type: application/json" \
  -d '{"name": "work"}'
```

**Add tag to account:**
```bash
# By tag ID
curl -X POST http://localhost:8080/api/v1/accounts/1/tags/1

# By tag name (easier - creates tag if doesn't exist)
curl -X POST http://localhost:8080/api/v1/accounts/1/tag/work
```

**Remove tag from account:**
```bash
# By tag ID
curl -X DELETE http://localhost:8080/api/v1/accounts/1/tags/1

# By tag name
curl -X DELETE http://localhost:8080/api/v1/accounts/1/tag/work
```

**Get accounts by tag:**
```bash
curl http://localhost:8080/api/v1/accounts/by-tag/work
```

**Create TOTP for account:**
```bash
curl -X POST http://localhost:8080/api/v1/totp \
  -H "Content-Type: application/json" \
  -d '{
    "account_id": 1,
    "totp_seed": "JBSWY3DPEHPK3PXP"
  }'
```

**Get TOTP for account:**
```bash
curl http://localhost:8080/api/v1/totp/account/1
```

## Database Schema

The application uses the following tables based on your exact specification:

### Main Tables

**account** - Password accounts
- `id` (INTEGER PRIMARY KEY) - Auto-incrementing account ID
- `email` (TEXT, nullable) - Email address
- `user` (TEXT, required) - Username
- `password` (TEXT, required) - Password
- `url` (TEXT, nullable) - Associated URL
- `notes` (TEXT, nullable) - Additional notes
- `expire` (TEXT, required, default: current datetime) - Expiration timestamp

**tag** - Tags for organizing accounts
- `id` (INTEGER PRIMARY KEY) - Auto-incrementing tag ID
- `name` (TEXT, unique, required) - Tag name

**tag_association** - Links accounts to tags
- `account_id` (INTEGER, required) - Foreign key to account
- `tag_id` (INTEGER, required) - Foreign key to tag
- Primary key: (account_id, tag_id)

**totp** - TOTP seeds for 2FA (linked to accounts)
- `id` (INTEGER PRIMARY KEY) - Auto-incrementing TOTP ID
- `account_id` (INTEGER, unique, required) - Foreign key to account (one TOTP per account)
- `totp_seed` (TEXT, required) - TOTP secret seed

### History Tables

**account_history** - Historical versions of accounts
- `history_id` (INTEGER PRIMARY KEY) - Auto-incrementing history ID
- `account_id` (INTEGER, nullable) - Reference to account
- `email` (TEXT, nullable) - Historical email
- `user` (TEXT, nullable) - Historical username
- `password` (TEXT, nullable) - Historical password
- `url` (TEXT, nullable) - Historical URL
- `notes` (TEXT, nullable) - Historical notes
- `expire` (TEXT, nullable) - Historical expiration
- `valid_from` (TEXT, required, default: current datetime) - Start of validity
- `valid_to` (TEXT, required, default: current datetime) - End of validity
- `change_reason` (TEXT, nullable) - Reason for the change

**tag_association_history** - Historical tag associations
- `account_id` (INTEGER, required) - Account ID
- `tag_id` (INTEGER, required) - Tag ID
- `valid_from` (TEXT, required, default: current datetime) - Start of validity
- `valid_to` (TEXT, required, default: current datetime) - End of validity
- Primary key: (account_id, tag_id, valid_from)

**totp_history** - Historical TOTP seeds
- `history_id` (INTEGER PRIMARY KEY) - Auto-incrementing history ID
- `totp_id` (INTEGER, required) - Reference to TOTP
- `totp_seed` (TEXT, required) - Historical TOTP seed
- `valid_from` (TEXT, required, default: current datetime) - Start of validity
- `valid_to` (TEXT, required, default: current datetime) - End of validity

### Relationships

- `account.id` ← `tag_association.account_id` (one-to-many)
- `tag.id` ← `tag_association.tag_id` (one-to-many)
- `account.id` ← `totp.account_id` (one-to-one)
- `account.id` ← `account_history.account_id` (one-to-many)
- `account.id` ← `tag_association_history.account_id` (one-to-many)
- `tag.id` ← `tag_association_history.tag_id` (one-to-many)
- `totp.id` ← `totp_history.totp_id` (one-to-many)

All tables support full history tracking with `valid_from` and `valid_to` timestamps for audit purposes.

## Security Notes

1. **Keep your encryption key secure!** Without it, your database cannot be decrypted.
2. Store the encryption key in environment variables, not in code:
   ```bash
   export PWDMGR_KEY="your-secret-key"
   ./pwdmgr account list --key "$PWDMGR_KEY"
   ```
3. The database file is encrypted with AES-256 encryption via SQLCipher.
4. Consider using a key derivation function (KDF) for user passwords.
5. In production, add authentication/authorization to the REST API.

## Project Structure

```
.
├── main.go          # Entry point
├── models.go        # Data models with JSON tags
├── database.go      # Database connection and schema
├── repository.go    # Data access layer
├── api.go          # REST API handlers
├── cli.go          # CLI commands
├── go.mod          # Go module definition
└── README.md       # This file
```

## Development

### Running Tests

The project includes comprehensive test coverage (80%+) across all components.

**Run all tests:**
```bash
make test
```

**Run tests with coverage:**
```bash
make test-coverage
```

**Run specific test suites:**
```bash
make test-database      # Database tests
make test-repository    # Repository tests
make test-api          # API tests
make test-integration  # Integration tests
```

**Use the test runner script:**
```bash
./run_tests.sh
```

**Generate HTML coverage report:**
```bash
make test-coverage-html
# Opens coverage.html in your browser
```

See [TESTING.md](TESTING.md) for comprehensive testing documentation.

### Code Quality

**Run tests:**
```bash
go test ./...
```

**Format code:**
```bash
go fmt ./...
```

**Run with race detection:**
```bash
make test-race
```

**Build for different platforms:**
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o pwdmgr-linux

# macOS
GOOS=darwin GOARCH=amd64 go build -o pwdmgr-macos

# Windows
GOOS=windows GOARCH=amd64 go build -o pwdmgr.exe
```

## License

MIT License
