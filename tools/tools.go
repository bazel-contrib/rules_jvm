//go:build tools

package main

import (
	_ "github.com/aristanetworks/goarista/cmd/importsort"
	_ "golang.org/x/tools/cmd/goimports"
)
