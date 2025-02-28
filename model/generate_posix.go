//go:build linux || darwin

package model

// when you are working on linux or macos
//go:generate docker run --rm --volume ".:/workspace" --workdir /workspace bufbuild/buf generate
