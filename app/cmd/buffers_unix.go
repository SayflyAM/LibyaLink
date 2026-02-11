//go:build linux || darwin || freebsd || openbsd || netbsd
// +build linux darwin freebsd openbsd netbsd

package cmd

import "golang.org/x/sys/unix"

func getSockOptRcvBuf(fd uintptr) int {
	val, err := unix.GetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_RCVBUF)
	if err != nil {
		return 0
	}
	return val
}

func getSockOptSndBuf(fd uintptr) int {
	val, err := unix.GetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_SNDBUF)
	if err != nil {
		return 0
	}
	return val
}
