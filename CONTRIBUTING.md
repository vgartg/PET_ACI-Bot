# Contributing to ACI-Bot

Thanks for your interest in improving ACI-Bot! This document describes how to contribute changes.

## Getting started

1. Fork the repository and clone your fork.
2. Make sure you have Go 1.22+ installed.
3. Install dependencies: `go mod download`.
4. Run the test suite to confirm a clean baseline: `go test ./...`.

## Development workflow

1. Create a topic branch off `main`:
   ```
   git checkout -b feature/short-description
   ```
2. Make your changes in small, focused commits.
3. Run the linters and tests locally before pushing:
   ```
   go vet ./...
   gofmt -l .
   go test ./...
   ```
4. Push your branch and open a pull request against `main`.

## Commit messages

- Use the imperative mood ("Add command", not "Added command").
- Keep the subject line under 72 characters.
- Reference related issues in the body when applicable.

## Code style

- Format all Go code with `gofmt`.
- Keep exported identifiers documented with short Go doc comments.
- Prefer table-driven tests for new functionality.

## Reporting issues

When opening an issue, please include:

- A clear description of the problem or proposal.
- Steps to reproduce (for bugs).
- The Go version and operating system you are using.

## License

By contributing, you agree that your contributions will be licensed under the
[MIT License](LICENSE) that covers this project.
