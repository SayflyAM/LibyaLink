//go:build windows
// +build windows

package cmd

import (
	"syscall"
	"unsafe"
)

const (
	sysSOL_SOCKET  = 0xFFFF
	sysSO_RCVBUF   = 0x1002
	sysSO_SNDBUF   = 0x1001
)

func getSockOptRcvBuf(fd uintptr) int {
	return getsockoptInt(syscall.Handle(fd), sysSOL_SOCKET, sysSO_RCVBUF)
}

func getSockOptSndBuf(fd uintptr) int {
	return getsockoptInt(syscall.Handle(fd), sysSOL_SOCKET, sysSO_SNDBUF)
}

func getsockoptInt(fd syscall.Handle, level, opt int) int {
	var val int32
	vallen := int32(unsafe.Sizeof(val))
	err := syscall.Getsockopt(fd, int32(level), int32(opt), (*byte)(unsafe.Pointer(&val)), &vallen)
	if err != nil {
		return 0
	}
	return int(val)
}
