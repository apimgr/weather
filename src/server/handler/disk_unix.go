//go:build !windows
// +build !windows

package handler

import "syscall"

func getDiskUsage(path string) DiskUsage {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return DiskUsage{
			Path:        path,
			UsedBytes:   0,
			FreeBytes:   0,
			TotalBytes:  0,
			UsedPercent: 0,
		}
	}

	totalBytes := int64(stat.Blocks) * int64(stat.Bsize)
	freeBytes := int64(stat.Bavail) * int64(stat.Bsize)
	usedBytes := totalBytes - freeBytes
	usedPercent := 0
	if totalBytes > 0 {
		usedPercent = int(float64(usedBytes) / float64(totalBytes) * 100)
	}

	return DiskUsage{
		Path:        path,
		UsedBytes:   usedBytes,
		FreeBytes:   freeBytes,
		TotalBytes:  totalBytes,
		UsedPercent: usedPercent,
	}
}
