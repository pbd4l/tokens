# tokens

[![CI](https://github.com/pbd4l/tokens/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/pbd4l/tokens/actions/workflows/ci.yml)

CLI for token management.

## Installation & usage

1. Ensure you have a working Go installation - https://golang.org/. 

<!-- TODO build and publish a binary in CI -->

2. Install the `tokens` binary

```
go get -u github.com/pbd4l/tokens
```

3. Discover usage via `--help`

```
❯ tokens --help
CLI for token management

Usage:
  tokens [command]

Available Commands:
  generate    Generate random tokens
  help        Help about any command
  import      Import tokens to a postgres database

Flags:
  -h, --help   help for tokens

Use "tokens [command] --help" for more information about a command.
```
