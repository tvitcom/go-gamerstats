#!/bin/sh

if [ $(ip route | grep 192.168.10.100 > /dev/null 2>&1) ]; then
	export GIN_MODE=release
	echo "Server in production mode"
else
	echo "Server have DEBUG mode"
fi

go run main.go #2>> ./logs/errors.log

echo "Server running. See logs."
