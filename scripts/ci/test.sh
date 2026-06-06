#!/bin/bash
set -e
echo "Running tests..."
go test ./... -v -race -coverprofile=coverage.out
echo "Tests passed."
