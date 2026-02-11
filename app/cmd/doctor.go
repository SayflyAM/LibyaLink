package cmd

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	checkOK   = "✅"
	checkFail = "❌"
	checkWarn = "⚠️"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose server configuration and environment",
	Long: `Run a comprehensive diagnostic check on the server configuration and system environment.
Validates YAML syntax, TLS/ACME config, file permissions, port availability,
and system tuning parameters. Designed for operators to quickly identify issues.`,
	Run: runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

type checkResult struct {
	Name    string
	Status  string // checkOK, checkFail, checkWarn
	Message string
}

func runDoctor(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║          LibyaLink Doctor — System Diagnostic       ║")
	fmt.Println("║          Powered by Hysteria 2                      ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Println()

	results := make([]checkResult, 0, 10)

	// 1. Check config file readability
	results = append(results, checkConfigReadable()...)

	// 2. Check TLS / ACME conflict
	results = append(results, checkTLSACMEConflict()...)

	// 3. Check TLS cert/key file permissions
	results = append(results, checkTLSFiles()...)

	// 4. Check listen port availability
	results = append(results, checkPortAvailability()...)

	// 5. Check UDP buffer sizes (Linux)
	results = append(results, checkUDPBuffers()...)

	// 6. Check auth configuration
	results = append(results, checkAuthConfig()...)

	// Print results
	fmt.Println("─── Diagnostic Results ───")
	fmt.Println()

	failCount := 0
	warnCount := 0
	for _, r := range results {
		fmt.Printf("  %s  [%s] %s\n", r.Status, r.Name, r.Message)
		if r.Status == checkFail {
			failCount++
		}
		if r.Status == checkWarn {
			warnCount++
		}
	}

	fmt.Println()
	fmt.Println("──────────────────────────")

	if failCount == 0 && warnCount == 0 {
		fmt.Println("  ✅ System Healthy — All checks passed!")
	} else if failCount == 0 {
		fmt.Printf("  %s System OK with %d warning(s)\n", checkWarn, warnCount)
	} else {
		fmt.Printf("  %s %d error(s), %d warning(s) found. Fix the issues above.\n", checkFail, failCount, warnCount)
	}
	fmt.Println()
}

func checkConfigReadable() []checkResult {
	err := viper.ReadInConfig()
	if err != nil {
		return []checkResult{{
			Name:    "Config File",
			Status:  checkFail,
			Message: fmt.Sprintf("Cannot read config file: %v", err),
		}}
	}
	return []checkResult{{
		Name:    "Config File",
		Status:  checkOK,
		Message: fmt.Sprintf("Config loaded from: %s", viper.ConfigFileUsed()),
	}}
}

func checkTLSACMEConflict() []checkResult {
	// If config is not readable, skip
	if viper.ConfigFileUsed() == "" {
		return nil
	}

	hasTLS := viper.IsSet("tls")
	hasACME := viper.IsSet("acme")

	if hasTLS && hasACME {
		return []checkResult{{
			Name:    "TLS/ACME",
			Status:  checkFail,
			Message: "Both 'tls' and 'acme' are set. You must use one or the other, not both.",
		}}
	}
	if !hasTLS && !hasACME {
		return []checkResult{{
			Name:    "TLS/ACME",
			Status:  checkFail,
			Message: "Neither 'tls' nor 'acme' is configured. One is required for the server to start.",
		}}
	}
	if hasTLS {
		return []checkResult{{
			Name:    "TLS/ACME",
			Status:  checkOK,
			Message: "TLS mode: using local certificate files.",
		}}
	}
	return []checkResult{{
		Name:    "TLS/ACME",
		Status:  checkOK,
		Message: "ACME mode: using automatic certificate provisioning.",
	}}
}

func checkTLSFiles() []checkResult {
	if !viper.IsSet("tls") {
		return nil // ACME mode, no files to check
	}

	certPath := viper.GetString("tls.cert")
	keyPath := viper.GetString("tls.key")

	var results []checkResult

	// Check cert file
	if certPath == "" {
		results = append(results, checkResult{
			Name:    "TLS Cert",
			Status:  checkFail,
			Message: "tls.cert path is empty.",
		})
	} else {
		if r := checkFileReadable("TLS Cert", certPath); r.Status != checkOK {
			results = append(results, r)
		} else {
			results = append(results, r)
		}
	}

	// Check key file
	if keyPath == "" {
		results = append(results, checkResult{
			Name:    "TLS Key",
			Status:  checkFail,
			Message: "tls.key path is empty.",
		})
	} else {
		if r := checkFileReadable("TLS Key", keyPath); r.Status != checkOK {
			results = append(results, r)
		} else {
			results = append(results, r)
		}
	}

	// If both are readable, try to parse the pair
	if certPath != "" && keyPath != "" {
		_, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			results = append(results, checkResult{
				Name:    "TLS Pair",
				Status:  checkFail,
				Message: fmt.Sprintf("Certificate/Key pair is invalid: %v", err),
			})
		} else {
			results = append(results, checkResult{
				Name:    "TLS Pair",
				Status:  checkOK,
				Message: "Certificate and key pair loaded successfully.",
			})
		}
	}

	return results
}

func checkFileReadable(name, path string) checkResult {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return checkResult{
			Name:    name,
			Status:  checkFail,
			Message: fmt.Sprintf("File not found: %s", path),
		}
	}
	if os.IsPermission(err) {
		return checkResult{
			Name:    name,
			Status:  checkFail,
			Message: fmt.Sprintf("Permission denied on %s", path),
		}
	}
	if err != nil {
		return checkResult{
			Name:    name,
			Status:  checkFail,
			Message: fmt.Sprintf("Error accessing %s: %v", path, err),
		}
	}

	// Try actually reading
	f, err := os.Open(path)
	if err != nil {
		return checkResult{
			Name:    name,
			Status:  checkFail,
			Message: fmt.Sprintf("Cannot open %s: %v", path, err),
		}
	}
	f.Close()

	// Check file isn't empty
	if info.Size() == 0 {
		return checkResult{
			Name:    name,
			Status:  checkFail,
			Message: fmt.Sprintf("File is empty: %s", path),
		}
	}

	return checkResult{
		Name:    name,
		Status:  checkOK,
		Message: fmt.Sprintf("Readable (%d bytes): %s", info.Size(), path),
	}
}

func checkPortAvailability() []checkResult {
	listenAddr := viper.GetString("listen")
	if listenAddr == "" {
		listenAddr = defaultListenAddr
	}

	var results []checkResult

	// Check UDP port (primary for QUIC)
	udpAddr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		results = append(results, checkResult{
			Name:    "UDP Port",
			Status:  checkFail,
			Message: fmt.Sprintf("Invalid listen address '%s': %v", listenAddr, err),
		})
		return results
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "address already in use") ||
			strings.Contains(errStr, "Only one usage of each socket address") {
			results = append(results, checkResult{
				Name:    "UDP Port",
				Status:  checkFail,
				Message: fmt.Sprintf("Port %s is already in use! Another process (Apache/Nginx/Hysteria?) is binding it.", listenAddr),
			})
		} else if strings.Contains(errStr, "permission denied") ||
			strings.Contains(errStr, "bind: permission denied") {
			results = append(results, checkResult{
				Name:    "UDP Port",
				Status:  checkFail,
				Message: fmt.Sprintf("Permission denied binding to %s. Use a port > 1024 or run with elevated privileges.", listenAddr),
			})
		} else {
			results = append(results, checkResult{
				Name:    "UDP Port",
				Status:  checkFail,
				Message: fmt.Sprintf("Cannot bind UDP %s: %v", listenAddr, err),
			})
		}
	} else {
		conn.Close()
		results = append(results, checkResult{
			Name:    "UDP Port",
			Status:  checkOK,
			Message: fmt.Sprintf("UDP %s is available.", listenAddr),
		})
	}

	return results
}

func checkUDPBuffers() []checkResult {
	if runtime.GOOS != "linux" {
		return []checkResult{{
			Name:    "UDP Buffers",
			Status:  checkWarn,
			Message: fmt.Sprintf("Buffer check only runs on Linux (current OS: %s). See docs/libya_tuning.md", runtime.GOOS),
		}}
	}

	var results []checkResult

	// Check rmem_max
	rmem, err := os.ReadFile("/proc/sys/net/core/rmem_max")
	if err == nil {
		val := strings.TrimSpace(string(rmem))
		results = append(results, checkBufferValue("UDP rmem_max", val, 8388608))
	}

	// Check wmem_max
	wmem, err := os.ReadFile("/proc/sys/net/core/wmem_max")
	if err == nil {
		val := strings.TrimSpace(string(wmem))
		results = append(results, checkBufferValue("UDP wmem_max", val, 8388608))
	}

	if len(results) == 0 {
		results = append(results, checkResult{
			Name:    "UDP Buffers",
			Status:  checkWarn,
			Message: "Could not read sysctl buffer values. Run 'sysctl net.core.rmem_max' manually.",
		})
	}

	return results
}

func checkBufferValue(name, valStr string, recommended int) checkResult {
	var val int
	fmt.Sscanf(valStr, "%d", &val)
	if val >= recommended {
		return checkResult{
			Name:    name,
			Status:  checkOK,
			Message: fmt.Sprintf("%d bytes (>= %d recommended). Good!", val, recommended),
		}
	}
	return checkResult{
		Name:    name,
		Status:  checkWarn,
		Message: fmt.Sprintf("%d bytes (< %d recommended). Run the tuning script for full speed. See docs/libya_tuning.md", val, recommended),
	}
}

func checkAuthConfig() []checkResult {
	authType := viper.GetString("auth.type")
	if authType == "" {
		return []checkResult{{
			Name:    "Auth",
			Status:  checkFail,
			Message: "No auth.type configured. Server requires authentication.",
		}}
	}

	switch strings.ToLower(authType) {
	case "password":
		pw := viper.GetString("auth.password")
		if pw == "" {
			return []checkResult{{
				Name:    "Auth",
				Status:  checkFail,
				Message: "auth.type is 'password' but auth.password is empty.",
			}}
		}
		if len(pw) < 8 {
			return []checkResult{{
				Name:    "Auth",
				Status:  checkWarn,
				Message: "auth.password is very short (< 8 chars). Consider using a stronger password.",
			}}
		}
		return []checkResult{{
			Name:    "Auth",
			Status:  checkOK,
			Message: "Password authentication configured.",
		}}
	case "userpass":
		up := viper.GetStringMapString("auth.userpass")
		if len(up) == 0 {
			return []checkResult{{
				Name:    "Auth",
				Status:  checkFail,
				Message: "auth.type is 'userpass' but no user:password entries found.",
			}}
		}
		return []checkResult{{
			Name:    "Auth",
			Status:  checkOK,
			Message: fmt.Sprintf("User/pass authentication configured (%d users).", len(up)),
		}}
	case "http", "https":
		url := viper.GetString("auth.http.url")
		if url == "" {
			return []checkResult{{
				Name:    "Auth",
				Status:  checkFail,
				Message: "auth.type is 'http' but auth.http.url is empty.",
			}}
		}
		return []checkResult{{
			Name:    "Auth",
			Status:  checkOK,
			Message: fmt.Sprintf("HTTP authentication configured: %s", url),
		}}
	default:
		return []checkResult{{
			Name:    "Auth",
			Status:  checkOK,
			Message: fmt.Sprintf("Authentication type: %s", authType),
		}}
	}
}
