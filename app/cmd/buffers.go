package cmd

import (
	"fmt"
	"net"

	"go.uber.org/zap"
)

const (
	// libyalinkDesiredReadBuffer is the desired read buffer size (8 MB).
	// Tuned for high-jitter Libyan 4G networks (Libyaana/Al-Madar) and
	// high-RTT international links via LTT.
	libyalinkDesiredReadBuffer = 8 * 1024 * 1024

	// libyalinkDesiredWriteBuffer is the desired write buffer size (8 MB).
	libyalinkDesiredWriteBuffer = 8 * 1024 * 1024
)

// tuneUDPBuffer attempts to set the UDP socket read/write buffers to the
// desired sizes for optimal performance on unstable Libyan connections.
// It logs the requested vs. granted sizes transparently so operators can
// see if the OS limited the buffer and take action (e.g., sysctl tuning).
func tuneUDPBuffer(conn *net.UDPConn, log *zap.Logger) {
	if conn == nil || log == nil {
		return
	}

	log.Info("[LibyaLink] Tuning UDP socket buffers for Libyan network conditions...")

	// --- Read Buffer ---
	err := conn.SetReadBuffer(libyalinkDesiredReadBuffer)
	if err != nil {
		log.Warn("[LibyaLink] Failed to set UDP read buffer",
			zap.Int("requested_bytes", libyalinkDesiredReadBuffer),
			zap.Error(err),
		)
	} else {
		// Retrieve actual granted size
		granted := getUDPReadBufferSize(conn)
		if granted < libyalinkDesiredReadBuffer {
			log.Warn(fmt.Sprintf("[LibyaLink] Requested %s read buffer -> OS granted %s. "+
				"Warning: Run the tuning script to unlock full speed. See docs/libya_tuning.md",
				formatBytes(libyalinkDesiredReadBuffer), formatBytes(granted)),
				zap.Int("requested_bytes", libyalinkDesiredReadBuffer),
				zap.Int("granted_bytes", granted),
			)
		} else {
			log.Info(fmt.Sprintf("[LibyaLink] UDP read buffer: requested %s -> granted %s. Optimal!",
				formatBytes(libyalinkDesiredReadBuffer), formatBytes(granted)),
				zap.Int("granted_bytes", granted),
			)
		}
	}

	// --- Write Buffer ---
	err = conn.SetWriteBuffer(libyalinkDesiredWriteBuffer)
	if err != nil {
		log.Warn("[LibyaLink] Failed to set UDP write buffer",
			zap.Int("requested_bytes", libyalinkDesiredWriteBuffer),
			zap.Error(err),
		)
	} else {
		granted := getUDPWriteBufferSize(conn)
		if granted < libyalinkDesiredWriteBuffer {
			log.Warn(fmt.Sprintf("[LibyaLink] Requested %s write buffer -> OS granted %s. "+
				"Warning: Run the tuning script to unlock full speed. See docs/libya_tuning.md",
				formatBytes(libyalinkDesiredWriteBuffer), formatBytes(granted)),
				zap.Int("requested_bytes", libyalinkDesiredWriteBuffer),
				zap.Int("granted_bytes", granted),
			)
		} else {
			log.Info(fmt.Sprintf("[LibyaLink] UDP write buffer: requested %s -> granted %s. Optimal!",
				formatBytes(libyalinkDesiredWriteBuffer), formatBytes(granted)),
				zap.Int("granted_bytes", granted),
			)
		}
	}
}

// getUDPReadBufferSize attempts to read the actual buffer size via SyscallConn.
// Falls back to 0 if the syscall approach isn't available.
func getUDPReadBufferSize(conn *net.UDPConn) int {
	raw, err := conn.SyscallConn()
	if err != nil {
		return 0
	}
	var size int
	raw.Control(func(fd uintptr) {
		size = getSockOptRcvBuf(fd)
	})
	return size
}

// getUDPWriteBufferSize attempts to read the actual buffer size via SyscallConn.
func getUDPWriteBufferSize(conn *net.UDPConn) int {
	raw, err := conn.SyscallConn()
	if err != nil {
		return 0
	}
	var size int
	raw.Control(func(fd uintptr) {
		size = getSockOptSndBuf(fd)
	})
	return size
}

func formatBytes(b int) string {
	const (
		kb = 1024
		mb = 1024 * kb
	)
	switch {
	case b >= mb:
		return fmt.Sprintf("%dMB", b/mb)
	case b >= kb:
		return fmt.Sprintf("%dKB", b/kb)
	default:
		return fmt.Sprintf("%dB", b)
	}
}
