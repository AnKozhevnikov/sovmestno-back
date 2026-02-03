//go:build tools
// +build tools

package tools

import (
	_ "github.com/swaggo/swag/cmd/swag"
)

// This file declares dependencies for build/dev tools
// These imports ensure 'go mod tidy' doesn't remove tool dependencies
