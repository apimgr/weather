# Weather Service Tests

Comprehensive test suite for the Weather Service including unit tests, integration tests, and end-to-end tests.

## Quick Start

```bash
# Run all tests
./tests/run-tests.sh

# Run with coverage report
./tests/run-tests.sh --coverage

# Run with verbose output
./tests/run-tests.sh --verbose

# Run benchmarks
./tests/run-tests.sh --bench

# Start test server for manual testing
./tests/test-server.sh
```

## Test Structure

```
tests/
├── README.md                           # This file
├── run-tests.sh                        # Main test runner
├── test-server.sh                      # Isolated test server
├── unit/                               # Unit tests
│   ├── services/
│   │   ├── location_enhancer_test.go  # Location service tests
│   │   └── weather_service_test.go    # Weather service tests
│   └── handlers/
│       └── auth_test.go               # Authentication tests
├── integration/                        # Integration tests
│   └── api_test.go                    # API endpoint tests
└── e2e/                               # End-to-end tests
    └── setup_flow_test.go             # Complete setup flow

```

## Test Scripts

### `run-tests.sh` - Main Test Runner

Runs the complete test suite with options for coverage and verbosity.

**Usage:**
```bash
# Run all tests
./tests/run-tests.sh

# With coverage report (generates coverage.html)
./tests/run-tests.sh --coverage

# With verbose output
./tests/run-tests.sh -v

# Run benchmarks
./tests/run-tests.sh --bench

# Combine options
./tests/run-tests.sh -c -v
```

**Options:**
- `-c, --coverage` - Generate coverage report
- `-v, --verbose` - Verbose test output
- `-b, --bench` - Run benchmarks

### `test-server.sh` - Isolated Test Server

Runs the weather service in an isolated temporary directory.

**Features:**
- Isolated temp directory per run (`/tmp/weather-test-<pid>`)
- Auto-cleanup on exit (Ctrl+C)
- No repo pollution
- Real-time log following

**Usage:**
```bash
# Basic usage (auto port)
./tests/test-server.sh

# Custom port
PORT=3053 ./tests/test-server.sh

# Keep temp directory for debugging
KEEP_TEMP=1 PORT=3053 ./tests/test-server.sh
```

**Environment Variables:**
- `PORT` - Server port (default: 3053)
- `KEEP_TEMP` - Set to `1` to keep temp directory

## Running Specific Tests

### Unit Tests

```bash
# All unit tests
go test ./tests/unit/...

# Service tests only
go test ./tests/unit/services/

# Handler tests only
go test ./tests/unit/handlers/

# Specific test
go test ./tests/unit/services/ -run TestLocationEnhancer_FindCityByID

# With coverage
go test -cover ./tests/unit/...
```

### Integration Tests

```bash
# All integration tests
go test ./tests/integration/...

# Specific integration test
go test ./tests/integration/ -run TestAPI_Weather_Coordinates

# With verbose output
go test -v ./tests/integration/...
```

### End-to-End Tests

```bash
# All e2e tests
go test ./tests/e2e/...

# Setup flow test
go test ./tests/e2e/ -run TestCompleteSetupFlow

# With timeout
go test -timeout 30s ./tests/e2e/...
```

## Manual Testing

### Start Test Server

```bash
# Start server
./tests/test-server.sh

# In another terminal, test endpoints:
curl http://localhost:3053/healthz
curl http://localhost:3053/api/v1/weather?lat=40.7128&lon=-74.0060
curl http://localhost:3053/api/v1/weather?city_id=5128581
curl "http://localhost:3053/api/v1/weather?lat=40.7128&lon=-74.0060&nearest=true"

# Stop server (Ctrl+C in first terminal)
```

## Coverage Reports

```bash
# Generate coverage report
./tests/run-tests.sh --coverage

# View in browser
open coverage.html  # macOS
xdg-open coverage.html  # Linux
start coverage.html  # Windows
```

## Benchmarking

```bash
# Run benchmarks
./tests/run-tests.sh --bench

# Or directly with Go
go test -bench=. -benchmem ./tests/...

# Specific benchmark
go test -bench=BenchmarkWeatherAPI ./tests/integration/
```

## Continuous Integration

The test suite is designed to work with CI/CD pipelines:

```yaml
# Example: GitHub Actions
- name: Run tests
  run: ./tests/run-tests.sh --coverage

- name: Upload coverage
  uses: codecov/codecov-action@v3
  with:
    files: ./coverage.out
```

## Writing New Tests

### Unit Test Template

```go
package mypackage_test

import (
    "testing"
)

func TestMyFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "test", "expected", false},
        {"invalid input", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFunction(tt.input)

            if (err != nil) != tt.wantErr {
                t.Errorf("MyFunction() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if got != tt.want {
                t.Errorf("MyFunction() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Best Practices

1. **Isolation** - Each test should be independent
2. **Cleanup** - Always clean up resources (defer, t.Cleanup)
3. **Table-driven** - Use table-driven tests for multiple cases
4. **Descriptive names** - Test names should describe what they test
5. **Fast tests** - Keep unit tests fast (<100ms each)
6. **No external deps** - Unit tests should not require network/DB

## Test Data

- All test data stored in `/tmp/weather-test-<pid>`
- Automatically cleaned up on exit
- Use `KEEP_TEMP=1` to inspect after test run
- Never commit test databases or temp files

## Troubleshooting

### Tests failing with database errors
```bash
# Ensure database schema is up to date
go test ./tests/... -v
```

### Port already in use
```bash
# Change port for test server
PORT=3054 ./tests/test-server.sh
```

### Coverage report not generating
```bash
# Ensure you have write permissions
./tests/run-tests.sh --coverage
ls -la coverage.out coverage.html
```

## Notes

- Test server uses isolated temp directories
- No pollution of your working directory
- All tests run in-memory databases
- Coverage reports saved to `coverage.html`
- Tests require Go 1.21 or higher
