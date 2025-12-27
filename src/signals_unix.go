//go:build !windows
// +build !windows

package main

import (
	"syscall"
)

var platformSignals = []syscall.Signal{
	// Reopen log files
	syscall.SIGUSR1,
	// Toggle debug mode
	syscall.SIGUSR2,
}
