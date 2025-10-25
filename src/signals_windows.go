//go:build windows
// +build windows

package main

import (
	"syscall"
)

var platformSignals = []syscall.Signal{}
