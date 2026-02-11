package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	genClientServer   string
	genClientPort     int
	genClientAuth     string
	genClientInsecure bool
	genClientSNI      string
	genClientObfs     string
	genClientPreset   string
	genClientOutput   string
)

var genClientCmd = &cobra.Command{
	Use:   "gen-client",
	Short: "Generate client configuration for NekoBox/sing-box",
	Long: `Generate a pre-formatted JSON configuration snippet specifically for
NekoBox (sing-box format). Handles the specific syntax differences including
server_name mapping and TLS insecure flags. Designed to eliminate client
configuration errors for Libyan operators.

Examples:
  libyalink gen-client --server 1.2.3.4 --auth "mypassword"
  libyalink gen-client --server 1.2.3.4 --port 8443 --auth "mypassword" --insecure
  libyalink gen-client --server 1.2.3.4 --auth "mypassword" --preset fiber
  libyalink gen-client --server 1.2.3.4 --auth "mypassword" -o client.json`,
	Run: runGenClient,
}

func init() {
	initGenClientFlags()
	rootCmd.AddCommand(genClientCmd)
}

func initGenClientFlags() {
	genClientCmd.Flags().StringVar(&genClientServer, "server", "", "server IP address or hostname (required)")
	genClientCmd.Flags().IntVar(&genClientPort, "port", 443, "server port")
	genClientCmd.Flags().StringVar(&genClientAuth, "auth", "", "authentication password (required)")
	genClientCmd.Flags().BoolVar(&genClientInsecure, "insecure", true, "skip TLS certificate verification (default: true for self-signed)")
	genClientCmd.Flags().StringVar(&genClientSNI, "sni", "", "TLS SNI (server name indication)")
	genClientCmd.Flags().StringVar(&genClientObfs, "obfs", "", "obfuscation password (salamander)")
	genClientCmd.Flags().StringVar(&genClientPreset, "preset", "4g", "bandwidth preset: '4g' (1-10 Mbps) or 'fiber' (50-100 Mbps)")
	genClientCmd.Flags().StringVar(&genClientOutput, "output", "", "output file path (default: stdout)")

	genClientCmd.MarkFlagRequired("server")
	genClientCmd.MarkFlagRequired("auth")
}

// bandwidthPreset holds up/down bandwidth values
type bandwidthPreset struct {
	Up   string `json:"up"`
	Down string `json:"down"`
}

var bandwidthPresets = map[string]bandwidthPreset{
	"4g": {
		Up:   "1 mbps",
		Down: "10 mbps",
	},
	"fiber": {
		Up:   "20 mbps",
		Down: "100 mbps",
	},
}

// singBoxOutbound represents a sing-box Hysteria2 outbound configuration
type singBoxOutbound struct {
	Type       string          `json:"type"`
	Tag        string          `json:"tag"`
	Server     string          `json:"server"`
	ServerPort int             `json:"server_port"`
	Password   string          `json:"password"`
	TLS        singBoxTLS      `json:"tls"`
	Obfs       *singBoxObfs    `json:"obfs,omitempty"`
	UpMbps     int             `json:"up_mbps,omitempty"`
	DownMbps   int             `json:"down_mbps,omitempty"`
}

type singBoxTLS struct {
	Enabled    bool   `json:"enabled"`
	Insecure   bool   `json:"insecure"`
	ServerName string `json:"server_name,omitempty"`
}

type singBoxObfs struct {
	Type     string `json:"type"`
	Password string `json:"password"`
}

// singBoxConfig is the full sing-box configuration structure
type singBoxConfig struct {
	Log       singBoxLog        `json:"log"`
	DNS       singBoxDNS        `json:"dns"`
	Inbounds  []singBoxInbound  `json:"inbounds"`
	Outbounds []interface{}     `json:"outbounds"`
	Route     singBoxRoute      `json:"route"`
}

type singBoxLog struct {
	Level string `json:"level"`
}

type singBoxDNS struct {
	Servers []singBoxDNSServer `json:"servers"`
}

type singBoxDNSServer struct {
	Tag     string `json:"tag"`
	Address string `json:"address"`
}

type singBoxInbound struct {
	Type   string `json:"type"`
	Tag    string `json:"tag"`
	Listen string `json:"listen"`
	Port   int    `json:"listen_port"`
}

type singBoxRoute struct {
	AutoDetectInterface bool             `json:"auto_detect_interface"`
	FinalTag            string           `json:"final"`
	Rules               []singBoxRouteRule `json:"rules,omitempty"`
}

type singBoxRouteRule struct {
	Protocol string `json:"protocol,omitempty"`
	Outbound string `json:"outbound"`
}

// hysteria2ClientConfig generates a native Hysteria 2 YAML-style client config
type hysteria2ClientConfig struct {
	Server    string                 `json:"server"`
	Auth      string                 `json:"auth"`
	TLS       hysteria2ClientTLS     `json:"tls"`
	Bandwidth *hysteria2ClientBW     `json:"bandwidth,omitempty"`
	Obfs      *hysteria2ClientObfs   `json:"obfs,omitempty"`
	Socks5    *hysteria2ClientSocks5 `json:"socks5,omitempty"`
	HTTP      *hysteria2ClientHTTP   `json:"http,omitempty"`
}

type hysteria2ClientTLS struct {
	SNI      string `json:"sni,omitempty"`
	Insecure bool   `json:"insecure"`
}

type hysteria2ClientBW struct {
	Up   string `json:"up"`
	Down string `json:"down"`
}

type hysteria2ClientObfs struct {
	Type       string `json:"type"`
	Salamander struct {
		Password string `json:"password"`
	} `json:"salamander"`
}

type hysteria2ClientSocks5 struct {
	Listen string `json:"listen"`
}

type hysteria2ClientHTTP struct {
	Listen string `json:"listen"`
}

func runGenClient(cmd *cobra.Command, args []string) {
	// Validate preset
	preset, ok := bandwidthPresets[genClientPreset]
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: unknown preset '%s'. Use '4g' or 'fiber'.\n", genClientPreset)
		os.Exit(1)
	}

	serverAddr := fmt.Sprintf("%s:%d", genClientServer, genClientPort)

	sni := genClientSNI
	if sni == "" && genClientInsecure {
		sni = genClientServer
	}

	// Parse bandwidth to Mbps integers for sing-box format
	upMbps, downMbps := parseBandwidthToMbps(preset)

	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "╔══════════════════════════════════════════════════════════╗")
	fmt.Fprintln(os.Stderr, "║  LibyaLink Client Config Generator                      ║")
	fmt.Fprintln(os.Stderr, "║  Powered by Hysteria 2                                  ║")
	fmt.Fprintln(os.Stderr, "╚══════════════════════════════════════════════════════════╝")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintf(os.Stderr, "  Server:   %s\n", serverAddr)
	fmt.Fprintf(os.Stderr, "  Preset:   %s (%s up / %s down)\n", genClientPreset, preset.Up, preset.Down)
	fmt.Fprintf(os.Stderr, "  Insecure: %v\n", genClientInsecure)
	fmt.Fprintln(os.Stderr, "")

	// --- Generate sing-box / NekoBox format ---
	fmt.Fprintln(os.Stderr, "─── NekoBox / sing-box Configuration ───")
	fmt.Fprintln(os.Stderr, "")

	var obfs *singBoxObfs
	if genClientObfs != "" {
		obfs = &singBoxObfs{
			Type:     "salamander",
			Password: genClientObfs,
		}
	}

	hy2Outbound := singBoxOutbound{
		Type:       "hysteria2",
		Tag:        "libyalink-proxy",
		Server:     genClientServer,
		ServerPort: genClientPort,
		Password:   genClientAuth,
		TLS: singBoxTLS{
			Enabled:    true,
			Insecure:   genClientInsecure,
			ServerName: sni,
		},
		Obfs:     obfs,
		UpMbps:   upMbps,
		DownMbps: downMbps,
	}

	singBoxCfg := singBoxConfig{
		Log: singBoxLog{Level: "info"},
		DNS: singBoxDNS{
			Servers: []singBoxDNSServer{
				{Tag: "google", Address: "tls://8.8.8.8"},
			},
		},
		Inbounds: []singBoxInbound{
			{
				Type:   "tun",
				Tag:    "tun-in",
				Listen: "0.0.0.0",
				Port:   0,
			},
			{
				Type:   "socks",
				Tag:    "socks-in",
				Listen: "127.0.0.1",
				Port:   2080,
			},
			{
				Type:   "http",
				Tag:    "http-in",
				Listen: "127.0.0.1",
				Port:   2081,
			},
		},
		Outbounds: []interface{}{
			hy2Outbound,
			map[string]string{"type": "direct", "tag": "direct"},
		},
		Route: singBoxRoute{
			AutoDetectInterface: true,
			FinalTag:            "libyalink-proxy",
		},
	}

	singBoxJSON, err := json.MarshalIndent(singBoxCfg, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating sing-box config: %v\n", err)
		os.Exit(1)
	}

	// --- Also generate native Hysteria 2 client format ---
	fmt.Fprintln(os.Stderr, "─── Native Hysteria 2 Client Configuration ───")
	fmt.Fprintln(os.Stderr, "")

	nativeConfig := hysteria2ClientConfig{
		Server: serverAddr,
		Auth:   genClientAuth,
		TLS: hysteria2ClientTLS{
			SNI:      sni,
			Insecure: genClientInsecure,
		},
		Bandwidth: &hysteria2ClientBW{
			Up:   preset.Up,
			Down: preset.Down,
		},
		Socks5: &hysteria2ClientSocks5{Listen: "127.0.0.1:1080"},
		HTTP:   &hysteria2ClientHTTP{Listen: "127.0.0.1:8080"},
	}

	if genClientObfs != "" {
		nativeConfig.Obfs = &hysteria2ClientObfs{
			Type: "salamander",
		}
		nativeConfig.Obfs.Salamander.Password = genClientObfs
	}

	nativeJSON, err := json.MarshalIndent(nativeConfig, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating native config: %v\n", err)
		os.Exit(1)
	}

	// Build full output
	output := fmt.Sprintf(`// ============================================================
// LibyaLink Client Configuration — Generated Automatically
// Powered by Hysteria 2
// Preset: %s (%s up / %s down)
// ============================================================

// ─── For NekoBox / sing-box ─────────────────────────────────
// Import this JSON in NekoBox > Manual Configuration > sing-box

%s

// ─── For Native Hysteria 2 Client ───────────────────────────
// Save as config.yaml and run: libyalink client -c config.yaml

%s
`, genClientPreset, preset.Up, preset.Down, string(singBoxJSON), string(nativeJSON))

	// Write to file or stdout
	if genClientOutput != "" {
		err := os.WriteFile(genClientOutput, []byte(output), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to %s: %v\n", genClientOutput, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "  ✅ Configuration written to: %s\n", genClientOutput)
	} else {
		fmt.Print(output)
	}

	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "  📋 Copy the sing-box JSON block into NekoBox's manual config.")
	fmt.Fprintln(os.Stderr, "  📋 Or save the Hysteria 2 block as config.yaml for the native client.")
	fmt.Fprintln(os.Stderr, "")
}

func parseBandwidthToMbps(preset bandwidthPreset) (upMbps, downMbps int) {
	fmt.Sscanf(preset.Up, "%d", &upMbps)
	fmt.Sscanf(preset.Down, "%d", &downMbps)
	return
}
