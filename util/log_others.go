//go:build !windows

package util

import "github.com/fatih/color"

var SymInfo = color.New(color.FgCyan, color.Bold).Sprint("ℹ")
var SymCheck = color.New(color.FgGreen, color.Bold).Sprint("✔")
var SymWarn = color.New(color.FgYellow, color.Bold).Sprint("⚠")
var SymError = color.New(color.FgRed, color.Bold).Sprint("✖️")
