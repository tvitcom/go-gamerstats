#!/bin/sh

go run main.go #2>> ./logs/errors.log

echo "Server running. See logs."
