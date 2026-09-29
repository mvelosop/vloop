// Package vloop holds the files embedded into the vloop binary.
package vloop

import "embed"

// Plugin is the plugin skeleton. The all: prefix is required: without it
// plugin/.claude-plugin/ (a dot-directory) is silently dropped.
//
//go:embed all:plugin
var Plugin embed.FS
