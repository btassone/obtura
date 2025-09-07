# Obtura Testing Guide

This guide covers the testing infrastructure and practices for the Obtura framework.

## Quick Start

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test
make test-specific name=TestPluginRegistry

# Watch tests (requires gotestsum)
make test-watch
```

## Test Structure

```
test/
├── README.md           # This file
├── testutil/          # Test utilities and helpers
│   ├── helpers.go     # Common test helpers
│   ├── database.go    # Database test utilities
│   ├── auth.go        # Authentication test helpers
│   └── server.go      # HTTP server test utilities
└── integration/       # Integration tests
    └── full_flow_test.go

pkg/plugin/            # Plugin tests
├── registry_test.go
└── config_test.go

internal/              # Internal package tests
├── server/
│   └── server_test.go
├── middleware/
│   └── admin_test.go
├── database/
│   └── manager_test.go
└── models/
    └── user_test.go

plugins/auth/          # Plugin-specific tests
├── plugin_test.go
└── basic_provider_test.go
```

## Test Types

### Unit Tests
Standard Go tests for individual components:
```bash
make test-unit
```

### Integration Tests
Tests that verify complete workflows:
```bash
make test-integration
```

### Benchmarks
Performance tests:
```bash
make test-bench
```

## Test Utilities

### TestPlugin
Mock plugin implementation for testing:
```go
plugin := &testutil.TestPlugin{
    NameFunc: func() string { return "Test Plugin" },
}
```

### Database Testing
In-memory SQLite for fast tests:
```go
db := testutil.TestDB(t)
defer db.Close()
```

### Authentication Testing
JWT token generation for tests:
```go
jwt := testutil.DefaultTestJWT()
token, _ := jwt.GenerateAdminToken(1)
```

### HTTP Testing
Test server setup:
```go
ts := testutil.NewTestServer(t, handler)
resp := ts.Request(t, "GET", "/api/users", nil)
```

## Coverage

### Local Coverage
Generate HTML coverage report:
```bash
make test-coverage
# Opens coverage/coverage.html in browser
```

### CI Coverage
Tests run automatically on:
- Push to main/develop branches
- Pull requests

Coverage reports are uploaded to Codecov.

### Coverage Threshold
- Project: 70% minimum
- Patch: 80% minimum

## Writing Tests

### Naming Convention
- Test files: `*_test.go`
- Test functions: `Test<Function>_<Scenario>`
- Benchmarks: `Benchmark<Function>`

### Test Structure
```go
func TestUserRepository_Create(t *testing.T) {
    // Arrange
    db := setupTestDB(t)
    repo := NewUserRepository(db)
    
    // Act
    err := repo.Create(user)
    
    // Assert
    require.NoError(t, err)
    assert.Equal(t, expected, actual)
}
```

### Table-Driven Tests
```go
tests := []struct {
    name    string
    input   string
    want    string
    wantErr bool
}{
    {"valid input", "test", "TEST", false},
    {"empty input", "", "", true},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        got, err := Function(tt.input)
        if tt.wantErr {
            require.Error(t, err)
            return
        }
        require.NoError(t, err)
        assert.Equal(t, tt.want, got)
    })
}
```

## Mocking

### Interface Mocking
```go
type mockAuthProvider struct {
    authFunc func(credentials interface{}) (*User, error)
}

func (m *mockAuthProvider) Authenticate(creds interface{}) (*User, error) {
    if m.authFunc != nil {
        return m.authFunc(creds)
    }
    return nil, errors.New("not implemented")
}
```

### HTTP Mocking
```go
handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
})
```

## CI/CD Integration

### GitHub Actions
Tests run automatically with:
- Unit tests
- Integration tests
- Coverage reporting
- Multi-platform builds

### Running CI Tests Locally
```bash
make test-ci
```

## Best Practices

1. **Isolation**: Tests should not depend on external services
2. **Cleanup**: Always clean up test data and resources
3. **Parallel**: Use `t.Parallel()` for independent tests
4. **Assertions**: Use testify for clear assertions
5. **Coverage**: Aim for >80% coverage on new code
6. **Speed**: Keep unit tests under 100ms

## Debugging Tests

### Verbose Output
```bash
make test-verbose
```

### Run Single Test
```bash
go test -v -run TestSpecificFunction ./pkg/plugin
```

### Debug with Delve
```bash
dlv test ./pkg/plugin -- -test.run TestSpecificFunction
```

## Common Issues

### Import Cycles
- Move shared types to `internal/types`
- Use interfaces instead of concrete types

### Database Tests
- Use in-memory SQLite for speed
- Reset database between tests
- Use transactions for isolation

### Flaky Tests
- Avoid time-based assertions
- Use channels for synchronization
- Mock external dependencies

## Contributing

When adding new features:
1. Write tests first (TDD)
2. Ensure tests pass locally
3. Check coverage doesn't decrease
4. Update this documentation if needed