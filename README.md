# LibyaLink

[![License][1]][2]

[1]: https://img.shields.io/badge/license-MIT-blue
[2]: LICENSE.md

> **Powered by [Hysteria 2](https://github.com/apernet/hysteria)** — Optimized for Libyan networks.

<h2 style="text-align: center;">LibyaLink is a stability-first proxy fork, tuned for Libyan internet infrastructure.</h2>

---

<div class="feature-grid">
  <div>
    <h3>🇱🇾 Built for Libya</h3>
    <p>Aggressive UDP buffer tuning, bandwidth presets for Libyaana/Al-Madar 4G and LTT DSL/Fiber, and sysctl guides tailored to high-jitter Libyan networks.</p>
  </div>

  <div>
    <h3>⚡ Blazing fast</h3>
    <p>Powered by Hysteria 2's customized QUIC protocol, designed to deliver unparalleled performance over unreliable and lossy networks.</p>
  </div>

  <div>
    <h3>🩺 Ops-First Reliability</h3>
    <p><code>libyalink doctor</code> validates your config, TLS certs, port availability, and system tuning — so the server screams exactly why it failed.</p>
  </div>

  <div>
    <h3>📱 One-Command Client Configs</h3>
    <p><code>libyalink gen-client</code> generates ready-to-paste NekoBox/sing-box JSON and native Hysteria 2 configs. Zero client-side errors.</p>
  </div>

  <div>
    <h3>🛠️ Jack of all trades</h3>
    <p>SOCKS5, HTTP Proxy, TCP/UDP Forwarding, Linux TProxy, TUN — all inherited from Hysteria 2 with full compatibility.</p>
  </div>

  <div>
    <h3>✊ Censorship resistant</h3>
    <p>The protocol masquerades as standard HTTP/3 traffic, making it very difficult for censors to detect and block.</p>
  </div>
</div>

---

## Quick Start

### Server Setup (Ubuntu 22.04)

```bash
# Run the automated setup script
sudo bash scripts/setup_libyalink.sh

# Or manually:
sudo systemctl enable --now libyalink
```

### Generate Client Config

```bash
# For 4G users (Libyaana/Al-Madar)
libyalink gen-client --server YOUR_IP --auth "password" --preset 4g

# For Fiber users (LTT)
libyalink gen-client --server YOUR_IP --auth "password" --preset fiber
```

### Run Diagnostics

```bash
libyalink doctor -c /etc/libyalink/config.yaml
```

---

## Features

| Feature | Description |
|---|---|
| `libyalink doctor` | Full system diagnostic — config, TLS, ports, UDP buffers, auth |
| `libyalink gen-client` | NekoBox/sing-box + native client config generator |
| UDP Buffer Tuning | Auto-requests 8MB buffers, logs granted vs requested |
| Bandwidth Presets | `4g` (1/10 Mbps) and `fiber` (20/100 Mbps) |
| Systemd Service | Hardened unit file with pre-start validation |
| Tuning Guide | [docs/libya_tuning.md](docs/libya_tuning.md) — sysctl for high-RTT links |

---

## Documentation

- [Network Tuning Guide](docs/libya_tuning.md) — Sysctl optimization for Libyan connections
- [Systemd Service](docs/libyalink.service) — Production service unit
- [Setup Script](scripts/setup_libyalink.sh) — One-command Ubuntu 22.04 deployment
- [Protocol Specification](PROTOCOL.md) — Hysteria 2 protocol details
- [Changelog](CHANGELOG.md)

---

## Credits

LibyaLink is a downstream fork of [Hysteria 2](https://github.com/apernet/hysteria) by [Aperture Internet Laboratory](https://github.com/apernet). All upstream licenses are preserved.

**Powered by Hysteria 2** | MIT License
