# Contribution Guidelines

## Development Environment

### pre-requisites:

1. GoLang > 1.15

### Initial Setup

1. `go mod vendor`
2. `cd patches`
3. `./apply-patches.sh`
   - This step must be executed in a *nix bash/shell
   - On Windows use `git bash` / WSL

### Running Tests

- `go test -v ./...`

