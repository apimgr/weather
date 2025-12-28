//go:build windows
// +build windows

package handler

import (
	"syscall"
	"unsafe"
)

func getDiskUsage(path string) DiskUsage {
	// Windows implementation using GetDiskFreeSpaceEx
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpaceEx := kernel32.NewProc("GetDiskFreeSpaceExW")

	var freeBytesAvailable, totalBytes, totalFreeBytes int64

	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return DiskUsage{
			Path:        path,
			UsedBytes:   0,
			FreeBytes:   0,
			TotalBytes:  0,
			UsedPercent: 0,
		}
	}

	ret, _, _ := getDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)

	if ret == 0 {
		// Call failed
		return DiskUsage{
			Path:        path,
			UsedBytes:   0,
			FreeBytes:   0,
			TotalBytes:  0,
			UsedPercent: 0,
		}
	}

	usedBytes := totalBytes - totalFreeBytes
	usedPercent := 0
	if totalBytes > 0 {
		usedPercent = int(float64(usedBytes) / float64(totalBytes) * 100)
	}

	return DiskUsage{
		Path:        path,
		UsedBytes:   usedBytes,
		FreeBytes:   freeBytesAvailable,
		TotalBytes:  totalBytes,
		UsedPercent: usedPercent,
	}
}
