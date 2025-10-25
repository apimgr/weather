//go:build !windows
// +build !windows

package main

import (
	"syscall"
)

var platformSignals = []syscall.Signal{
	syscall.SIGUSR1, // Reopen log files
	syscall.SIGUSR2, // Toggle debug mode
}
