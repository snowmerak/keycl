//go:build linux || darwin

package store

// when you are working on linux or macos
//go:generate docker run --rm -v .:/src -w /src sqlc/sqlc generate
