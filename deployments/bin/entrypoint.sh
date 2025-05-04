#!/bin/sh

while ! nc -z db 5432; do sleep 1; done;

# Use reflex to rerun app when .go/html files change.
reflex -r '\.(go|html)$' -s -- sh -c 'echo "Running migrations..." && go run ./internal/db/migrate.go up && echo "Start application..." && go run ./cmd/app/main.go'
