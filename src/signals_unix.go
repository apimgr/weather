//go:build !windows
// +build !windows

package main

import (
	"syscall"
)

// SIGRTMIN+3 = signal 37 (Docker STOPSIGNAL per AI.md PART 27)
// On Linux: SIGRTMIN = 34, so SIGRTMIN+3 = 37
const sigRTMIN3 = syscall.Signal(37)

var platformSignals = []syscall.Signal{
	// Reopen log files
	syscall.SIGUSR1,
	// Toggle debug mode
	syscall.SIGUSR2,
	// Docker STOPSIGNAL (SIGRTMIN+3 = 37) per AI.md PART 27 line 6462
	sigRTMIN3,
}
