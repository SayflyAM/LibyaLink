# LibyaLink Network Tuning Guide for Ubuntu 22.04

> **Powered by Hysteria 2** — Performance‑tuned for Libyan internet infrastructure.

This guide explains how to optimize your Ubuntu 22.04 server for maximum throughput
on unstable Libyan connections — specifically high-jitter 4G networks (Libyaana/Al‑Madar)
and high-RTT LTT DSL/Fiber international links.

---

## Why Tuning Matters

Hysteria 2 (and LibyaLink) uses the QUIC protocol over UDP. By default, Linux limits
the UDP socket buffer sizes to **~208 KB** (`net.core.rmem_max` / `net.core.wmem_max`).
On high-RTT links (common for Libyan servers connecting to Europe/US), this creates a
bottleneck: the OS drops packets before the application can process them.

LibyaLink requests **8 MB** buffers at startup. If the OS cap is lower, you'll see a
warning like:

```
[LibyaLink] Requested 8MB read buffer -> OS granted 208KB.
Warning: Run the tuning script to unlock full speed. See docs/libya_tuning.md
```

## Quick Fix (One Command)

```bash
sudo bash -c 'cat >> /etc/sysctl.d/99-libyalink.conf << EOF
# LibyaLink / Hysteria 2 UDP buffer tuning
# Optimized for high-RTT Libyan international links (Libyaana, Al-Madar, LTT)
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
net.core.rmem_default = 8388608
net.core.wmem_default = 8388608

# Increase the UDP memory limits
net.ipv4.udp_mem = 65536 131072 262144

# Allow more queued connections
net.core.netdev_max_backlog = 10000

# Increase connection tracking for NAT
net.netfilter.nf_conntrack_max = 131072
EOF

sysctl --system'
```

After running this, restart LibyaLink:

```bash
sudo systemctl restart libyalink
```

You should now see:

```
[LibyaLink] UDP read buffer: requested 8MB -> granted 8MB. Optimal!
[LibyaLink] UDP write buffer: requested 8MB -> granted 8MB. Optimal!
```

---

## Detailed Explanation

### What Each Setting Does

| Sysctl Parameter | Value | Purpose |
|---|---|---|
| `net.core.rmem_max` | 16777216 (16 MB) | Max UDP receive buffer per socket |
| `net.core.wmem_max` | 16777216 (16 MB) | Max UDP send buffer per socket |
| `net.core.rmem_default` | 8388608 (8 MB) | Default UDP receive buffer |
| `net.core.wmem_default` | 8388608 (8 MB) | Default UDP send buffer |
| `net.ipv4.udp_mem` | 65536 131072 262144 | System-wide UDP memory pages (min/pressure/max) |
| `net.core.netdev_max_backlog` | 10000 | NIC receive queue length |

### Why 8 MB?

The bandwidth-delay product (BDP) formula:

$$BDP = Bandwidth \times RTT$$

For a typical Libyan server scenario:
- **Bandwidth**: 100 Mbps (LTT Fiber to Europe)
- **RTT**: 80-150ms (Libya → Europe typical)

$$BDP = 100 \text{ Mbps} \times 0.15 \text{s} = 15 \text{ Mbit} = 1.875 \text{ MB}$$

We set **8 MB** to provide headroom for burst handling and jitter absorption —
especially critical on 4G networks where jitter can spike to 200ms+.

---

## Verifying Your Settings

### Check Current Buffer Sizes

```bash
sysctl net.core.rmem_max net.core.wmem_max
```

Expected output:
```
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
```

### Run LibyaLink Doctor

```bash
libyalink doctor -c /etc/libyalink/config.yaml
```

This will check:
- ✅ Config file syntax
- ✅ TLS certificate accessibility
- ✅ Port availability
- ✅ UDP buffer sizes

---

## Network-Specific Notes

### Libyaana / Al-Madar 4G

- **Characteristics**: High jitter (50-200ms), variable bandwidth, frequent packet reordering
- **Recommendation**: Use the `4g` preset for client configs:
  ```bash
  libyalink gen-client --server YOUR_IP --auth "pass" --preset 4g
  ```
- **Server-side**: Keep the 8 MB buffer settings. The buffer absorbs jitter spikes.

### LTT DSL / Fiber

- **Characteristics**: Stable but high RTT to international destinations (80-150ms)
- **Recommendation**: Use the `fiber` preset:
  ```bash
  libyalink gen-client --server YOUR_IP --auth "pass" --preset fiber
  ```
- **Server-side**: The tuning above covers this scenario well.

---

## Firewall Configuration (UFW)

LibyaLink/Hysteria 2 primarily uses UDP. Common mistake: only opening TCP.

```bash
# Open UDP port for Hysteria (default 443)
sudo ufw allow 443/udp comment "LibyaLink/Hysteria 2"

# If using masquerade, also open TCP
sudo ufw allow 443/tcp comment "LibyaLink masquerade HTTPS"

# Reload
sudo ufw reload
```

---

## Optional: BBR Congestion Control

While LibyaLink uses QUIC (not TCP BBR), enabling BBR on the system helps any
masquerade TCP traffic:

```bash
sudo bash -c 'cat >> /etc/sysctl.d/99-libyalink.conf << EOF

# BBR congestion control for TCP masquerade traffic
net.core.default_qdisc = fq
net.ipv4.tcp_congestion_control = bbr
EOF

sysctl --system'
```

Verify:
```bash
sysctl net.ipv4.tcp_congestion_control
# Expected: net.ipv4.tcp_congestion_control = bbr
```

---

## Troubleshooting

### "Permission denied" on port 443

If running LibyaLink as a non-root user:

```bash
# Option 1: Use setcap (recommended)
sudo setcap cap_net_bind_service=+ep /usr/local/bin/libyalink

# Option 2: Use a port above 1024 and redirect with iptables
sudo iptables -t nat -A PREROUTING -p udp --dport 443 -j REDIRECT --to-port 8443
```

### "Address already in use"

Another service is using port 443:

```bash
# Find what's using port 443
sudo ss -ulnp | grep 443
sudo lsof -i :443

# Common culprits: Apache, Nginx, another Hysteria instance
sudo systemctl stop nginx  # if applicable
```

### Buffers Still Low After sysctl

Some VPS providers override sysctl in their init scripts. Check:

```bash
# Ensure the file is loaded
sudo sysctl -p /etc/sysctl.d/99-libyalink.conf

# Check if another file overrides it
ls -la /etc/sysctl.d/
grep -r rmem_max /etc/sysctl.d/
```

---

## Full Tuning Script

Save as `tune-libyalink.sh` and run with `sudo`:

```bash
#!/bin/bash
set -e

echo "╔══════════════════════════════════════════════╗"
echo "║  LibyaLink Network Tuning Script             ║"
echo "║  For Ubuntu 22.04 — Powered by Hysteria 2    ║"
echo "╚══════════════════════════════════════════════╝"

# Apply sysctl settings
cat > /etc/sysctl.d/99-libyalink.conf << 'EOF'
# LibyaLink / Hysteria 2 UDP & network tuning
# Optimized for Libyan internet infrastructure
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
net.core.rmem_default = 8388608
net.core.wmem_default = 8388608
net.ipv4.udp_mem = 65536 131072 262144
net.core.netdev_max_backlog = 10000
net.core.default_qdisc = fq
net.ipv4.tcp_congestion_control = bbr
EOF

sysctl --system

echo ""
echo "✅ Network tuning applied!"
echo ""
echo "Current buffer settings:"
sysctl net.core.rmem_max net.core.wmem_max
echo ""
echo "Restart LibyaLink to apply: sudo systemctl restart libyalink"
```

---

*LibyaLink — Delivering reliable connectivity for Libya. Powered by Hysteria 2.*
