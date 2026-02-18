#!/bin/sh

echo "--- All Go files (relative) ---"
./cmd/lsgo/lsgo \
	--use-relative-path \
	./...

echo "\n--- Excluding main.go ---"
./cmd/lsgo/lsgo \
	--use-relative-path \
	--skip-pattern 'main\.go$' \
	./...

echo "\n--- Including only lsgo.go ---"
./cmd/lsgo/lsgo \
	--use-relative-path \
	--keep-pattern 'lsgo\.go$' \
	./...
