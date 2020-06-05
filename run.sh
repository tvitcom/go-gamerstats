#!/bin/sh
# For debug mode only
go run main.go #2>> ./logs/errors.log

echo "Server running. See logs."
