# Integration Tests

This directory contains integration tests for the Telegram Bot Go implementation.

## Overview

Integration tests verify that multiple components work together correctly. Unlike unit tests that test individual functions in isolation, integration tests:

- Test complete workflows end-to-end
- Use real database connections (SQLite)
- Simulate actual Telegram webhook requests
- Verify interactions between components

## Test Categories

### 1. End-to-End Tests (Task 24.1 - Optional)

End-to-end tests simulate complete user interactions:

- **Message Processing Flow**: Send a webhook request → Process message → Verify response
- **Command Execution**: Test each command (/start, /help, /new, etc.) in a realistic environment
- **AI Conversation**: Test complete conversation flow with mock AI responses

**Example Structure**:
```go
func TestEndToEndMessageFlow(t *testing.T) {
    // Setup: Create test database, config, and HTTP server
    // Action: Send webhook POST request with Telegram update
    // Verify: Check response, database state, and side effects
}
```

### 2. Database Integration Tests (Task 24.2 - Optional)

Database integration tests use real SQLite databases:

- **Real Database Operations**: Test with actual SQLite instead of mocks
- **Concurrent Access**: Verify thread-safety with multiple goroutines
- **Data Persistence**: Verify data survives across connections
- **Cleanup Operations**: Test TTL expiration and cleanup routines

**Example Structure**:
```go
func TestDatabaseConcurrentAccess(t *testing.T) {
    // Setup: Create temporary SQLite database
    // Action: Launch multiple goroutines performing reads/writes
    // Verify: No data corruption, all operations succeed
    // Cleanup: Remove temporary database
}
```

### 3. Go Version Improvements Integration Tests (Task 13 - Optional)

Integration tests for the new features added in go-version-improvements spec:

#### 3.1 End-to-End User Interaction Tests (Task 13.1 - Optional)

Tests complete user interaction flows with the new permission system:

- **Regular User with ENABLE_USER_SETTING=true**: Verify users can modify their own configuration
- **Regular User with ENABLE_USER_SETTING=false**: Verify users cannot see or access config commands
- **Admin User with ENABLE_USER_SETTING=false**: Verify admins can modify configuration
- **Configuration Modification**: Test permission control for config changes
- **Database Switching**: Test with SQLite, MySQL, and PostgreSQL

**Requirements**: 4.1, 4.2, 4.3, 4.4, 5.8

**Example Structure**:
```go
func TestEndToEndUserInteraction(t *testing.T) {
    // Setup: Create test environment with different ENABLE_USER_SETTING values
    // Test 1: Regular user with ENABLE_USER_SETTING=true can modify config
    // Test 2: Regular user with ENABLE_USER_SETTING=false cannot see config commands
    // Test 3: Admin user can always modify config
    // Test 4: Test database operations with different backends
}
```

#### 3.2 Backward Compatibility Tests (Task 13.2 - Optional)

Tests backward compatibility with existing configurations:

- **Default ENABLE_USER_SETTING**: Verify default behavior when not set (should be true)
- **DB_PATH Fallback**: Verify DB_PATH is used when DSN is not set
- **Old SQLite Database**: Verify old database format can be read
- **Configuration Migration**: Test migration from old configuration format

**Requirements**: 7.1, 7.2, 7.4

**Example Structure**:
```go
func TestBackwardCompatibility(t *testing.T) {
    // Test 1: Unset ENABLE_USER_SETTING defaults to true
    // Test 2: Unset DSN uses DB_PATH for SQLite
    // Test 3: DSN takes priority over DB_PATH when both set
    // Test 4: Old SQLite database can be read and used
}
```

## Running Integration Tests

Integration tests are marked as optional in the task list. To run them when implemented:

```bash
# Run all tests including integration tests
go test ./internal/integration/...

# Run with verbose output
go test -v ./internal/integration/...

# Skip integration tests (default behavior)
go test -short ./...

# Run only integration tests with build tag
go test -tags=integration ./internal/integration/...
```

## Implementation Status

- ✅ Integration test framework created
- ⏭️ Task 24.1: End-to-end tests (Optional - Not implemented)
- ⏭️ Task 24.2: Database integration tests (Optional - Not implemented)
- ✅ Task 13: Go version improvements integration tests framework created
- ⏭️ Task 13.1: End-to-end user interaction tests (Optional - Not implemented)
- ⏭️ Task 13.2: Backward compatibility tests (Optional - Not implemented)

## Future Implementation

When implementing these optional tests, consider:

### End-to-End Test Requirements

1. **Test Server Setup**:
   - Create HTTP test server
   - Initialize all components (storage, config, handlers)
   - Use test-specific configuration

2. **Mock External Services**:
   - Mock Telegram API responses
   - Mock AI provider APIs (OpenAI, Azure, etc.)
   - Use httptest for HTTP mocking

3. **Test Data**:
   - Generate realistic Telegram updates
   - Create test user configurations
   - Prepare sample conversation histories

4. **Assertions**:
   - Verify HTTP response codes
   - Check database state after operations
   - Validate message content and format

### Database Integration Test Requirements

1. **Temporary Databases**:
   - Create unique database file per test
   - Clean up after test completion
   - Use t.TempDir() for isolation

2. **Concurrency Testing**:
   - Use sync.WaitGroup for goroutine coordination
   - Test with realistic concurrent load
   - Verify no race conditions (use -race flag)

3. **Performance Testing**:
   - Measure query performance
   - Test with large datasets
   - Verify cleanup efficiency

4. **Error Scenarios**:
   - Test database connection failures
   - Test disk space issues
   - Test corrupted data handling

## Test Helpers

When implementing, create helper functions:

```go
// setupTestDB creates a temporary SQLite database for testing
func setupTestDB(t *testing.T) *storage.SQLiteStorage

// createTestUpdate generates a Telegram update for testing
func createTestUpdate(chatID int64, text string) *tgbotapi.Update

// mockAIProvider creates a mock AI provider for testing
func mockAIProvider() agent.ChatAgent

// assertDatabaseState verifies expected database contents
func assertDatabaseState(t *testing.T, db storage.Storage, expected map[string]interface{})
```

## References

- Task 24: 编写集成测试
- Task 24.1: 编写端到端测试 (Optional)
- Task 24.2: 编写数据库集成测试 (Optional)
- Design Document: Testing Strategy section
- Requirements: All functional requirements
