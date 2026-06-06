#!/bin/bash
set -e
echo "Running golangci-lint..."
golangci-lint run ./...
echo "Lint passed."
