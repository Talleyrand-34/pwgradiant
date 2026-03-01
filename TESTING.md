# Testing Documentation

This document describes the test suite for the Password Manager application.

## Overview

The test suite consists of 5 test files covering all major components:

1. **database_test.go** - Database initialization and schema creation
2. **models_test.go** - Data models and JSON serialization
3. **repository_test.go** - Data access layer and CRUD operations
4. **api_test.go** - REST API endpoints
5. **cli_test.go** - Command-line interface

## Running Tests

### Run All Tests

```bash
# Using Go test
go test ./...

# Or using make
make test-all

# With verbose output
go test -v ./...
```

### Run Specific Test Files

```bash
# Database tests
make test-database

# Models tests
make test-models

# Repository tests
make test-repository

# API tests
make test-api

# CLI tests
make test-cli
```

### Run Individual Tests

```bash
# Run a specific test function
go test -v -run TestAccountRepositoryCreate

# Run tests matching a pattern
go test -v -run TestAccount
```

### Test Coverage

```bash
# Generate coverage report
make test-coverage

# This creates coverage.html which you can open in a browser
```

## Test File Details

### database_test.go

Tests database initialization and schema creation:

- **TestInitDB**: Verifies database initialization with encryption
- **TestInitDBInvalidKey**: Tests that wrong encryption key fails
- **TestCreateSchema**: Validates all tables are created
- **TestCreateSchemaIdempotent**: Ensures schema can be created multiple times
- **TestCloseDB**: Tests database connection closing
- **TestForeignKeysEnabled**: Verifies foreign key constraints are enabled

### models_test.go

Tests data models and JSON serialization:

- **TestAccountJSONMarshaling**: Account serialization/deserialization
- **TestAccountNullableFields**: Nullable field handling
- **TestTagJSONMarshaling**: Tag model JSON operations
- **TestTOTPJSONMarshaling**: TOTP model JSON operations
- **TestAccountHistoryJSONMarshaling**: History model JSON operations
- **TestCurrentTimestamp**: Timestamp format validation
- **TestAccountWithTOTP**: Account with embedded TOTP
- **TestTagAssociationJSONMarshaling**: Tag association JSON
- **TestTOTPHistoryJSONMarshaling**: TOTP history JSON

### repository_test.go

Tests data access layer with comprehensive CRUD operations:

#### Account Repository Tests
- **TestAccountRepositoryCreate**: Create new accounts
- **TestAccountRepositoryGetByID**: Retrieve account by ID
- **TestAccountRepositoryGetByIDNotFound**: Handle non-existent accounts
- **TestAccountRepositoryGetAll**: List all accounts
- **TestAccountRepositoryUpdate**: Update existing accounts
- **TestAccountRepositoryDelete**: Delete accounts
- **TestAccountRepositorySearch**: Search functionality
- **TestAccountRepositoryAddTag**: Add tags to accounts
- **TestAccountRepositoryRemoveTag**: Remove tags from accounts
- **TestAccountRepositoryGetHistory**: Retrieve account history

#### Tag Repository Tests
- **TestTagRepositoryCreate**: Create tags
- **TestTagRepositoryGetByID**: Retrieve tag by ID
- **TestTagRepositoryGetByName**: Retrieve tag by name
- **TestTagRepositoryGetAll**: List all tags
- **TestTagRepositoryUpdate**: Update tags
- **TestTagRepositoryDelete**: Delete tags

#### TOTP Repository Tests
- **TestTOTPRepositoryCreate**: Create TOTP entries
- **TestTOTPRepositoryGetByAccountID**: Retrieve TOTP by account
- **TestTOTPRepositoryUpdate**: Update TOTP seeds
- **TestTOTPRepositoryDelete**: Delete TOTP entries
- **TestTOTPRepositoryGetHistory**: Retrieve TOTP history

#### Constraint Tests
- **TestForeignKeyConstraints**: Validate foreign key enforcement

### api_test.go

Tests REST API endpoints using httptest:

#### Health & Basic Tests
- **TestHealthCheck**: Health endpoint validation

#### Account API Tests
- **TestCreateAccount**: POST /api/v1/accounts
- **TestCreateAccountInvalidJSON**: Invalid JSON handling
- **TestListAccounts**: GET /api/v1/accounts
- **TestGetAccount**: GET /api/v1/accounts/:id
- **TestGetAccountNotFound**: 404 handling
- **TestUpdateAccount**: PUT /api/v1/accounts/:id
- **TestDeleteAccount**: DELETE /api/v1/accounts/:id
- **TestSearchAccounts**: GET /api/v1/accounts/search
- **TestGetAccountHistory**: GET /api/v1/accounts/:id/history
- **TestAddTagToAccount**: POST /api/v1/accounts/:id/tags/:tag_id
- **TestRemoveTagFromAccount**: DELETE /api/v1/accounts/:id/tags/:tag_id

#### Tag API Tests
- **TestCreateTag**: POST /api/v1/tags
- **TestListTags**: GET /api/v1/tags
- **TestGetTag**: GET /api/v1/tags/:id
- **TestUpdateTag**: PUT /api/v1/tags/:id
- **TestDeleteTag**: DELETE /api/v1/tags/:id

#### TOTP API Tests
- **TestCreateTOTP**: POST /api/v1/totp
- **TestGetTOTPByAccount**: GET /api/v1/totp/account/:account_id
- **TestUpdateTOTP**: PUT /api/v1/totp/:id
- **TestDeleteTOTP**: DELETE /api/v1/totp/:id

### cli_test.go

Tests command-line interface structure and commands:

#### Core Tests
- **TestInitDatabase**: CLI database initialization
- **TestPrintJSON**: JSON output formatting
- **TestPrintJSONWithAccount**: Account JSON output
- **TestPrintJSONInvalidData**: Error handling

#### Command Structure Tests
- **TestRootCmdExists**: Root command validation
- **TestServerCmdExists**: Server command
- **TestInitCmdExists**: Init command
- **TestAccountCmdExists**: Account command group
- **TestTagCmdExists**: Tag command group

#### Account Command Tests
- **TestAccountAddCmdExists**: account add subcommand
- **TestAccountListCmdExists**: account list subcommand
- **TestAccountGetCmdExists**: account get subcommand
- **TestAccountUpdateCmdExists**: account update subcommand
- **TestAccountDeleteCmdExists**: account delete subcommand
- **TestAccountSearchCmdExists**: account search subcommand
- **TestAccountHistoryCmdExists**: account history subcommand

#### Tag Command Tests
- **TestTagAddCmdExists**: tag add subcommand
- **TestTagListCmdExists**: tag list subcommand
- **TestTagDeleteCmdExists**: tag delete subcommand

#### Flag Tests
- **TestGlobalFlagsExist**: Global flags (--database, --key)
- **TestServerFlagsExist**: Server-specific flags
- **TestAccountAddFlagsExist**: Account add flags
- **TestAccountUpdateFlagsExist**: Account update flags

#### Hierarchy Tests
- **TestCommandHierarchy**: Command tree structure
- **TestAccountSubcommands**: Account subcommand validation
- **TestTagSubcommands**: Tag subcommand validation

## Test Coverage Goals

The test suite aims for:

- **Unit Test Coverage**: 80%+ for all core functions
- **Integration Test Coverage**: All API endpoints tested
- **Edge Case Coverage**: Error handling, null values, invalid inputs

## Test Database

Tests use temporary SQLite databases with the pattern `test_*.db`. These are:

- Created at the start of each test suite
- Encrypted with test keys
- Automatically cleaned up after tests complete

## Best Practices

### Writing New Tests

1. **Use setup/cleanup functions**:
```go
func TestSomething(t *testing.T) {
    cleanup := setupTestDB(t)
    defer cleanup()
    
    // Your test code here
}
```

2. **Test error cases**:
```go
// Test success case
// Test failure case
// Test edge cases
```

3. **Use meaningful test names**:
```go
func TestAccountRepositoryCreateWithNullableFields(t *testing.T)
```

4. **Assert with clear error messages**:
```go
if result != expected {
    t.Errorf("Expected %v, got %v", expected, result)
}
```

### Running Tests During Development

```bash
# Watch mode (requires entr or similar)
find . -name "*.go" | entr -c go test -v ./...

# Quick validation
go test ./... && echo "✓ All tests passed"

# Before committing
make test-coverage
```

## Continuous Integration

Tests should be run in CI/CD pipelines:

```yaml
# Example GitHub Actions
- name: Run Tests
  run: |
    go test -v -race -coverprofile=coverage.txt ./...
    
- name: Upload Coverage
  uses: codecov/codecov-action@v3
```

## Common Issues

### Issue: Database Locked

**Solution**: Ensure tests properly close database connections:
```go
defer CloseDB()
```

### Issue: Tests Fail with "wrong key"

**Solution**: Clean up test databases before running:
```bash
make test-clean
```

### Issue: Port Already in Use (API tests)

**Solution**: API tests use httptest which doesn't bind to real ports, so this shouldn't occur.

## Performance Testing

For performance testing:

```bash
# Benchmark tests
go test -bench=. -benchmem

# Profile CPU usage
go test -cpuprofile cpu.prof
go tool pprof cpu.prof
```

## Test Metrics

Current test metrics (as of last update):

- **Total Tests**: 70+
- **Test Files**: 5
- **Coverage**: Target 80%+
- **Execution Time**: < 5 seconds (all tests)

## Contributing

When adding new features:

1. Write tests first (TDD approach recommended)
2. Ensure all existing tests pass
3. Add integration tests for API changes
4. Update this documentation if needed
5. Run `make test-coverage` before submitting PR

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Testify Framework](https://github.com/stretchr/testify) (if added)
