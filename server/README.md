# TESTING

## Unit-Test Case

```sh
# Run all tests
go test ./...

# Run tests for specific package
go test ./handlers
go test ./store

# Run with verbose output
go test -v ./handlers ./store

# Run specific test
go test -run TestHandleBooking ./handlers

# Run with coverage
go test -cover ./handlers ./store
```
