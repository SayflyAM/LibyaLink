#!/bin/bash
# LibyaLink Server Quick Setup Script for Ubuntu 22.04
# Powered by Hysteria 2
set -e

INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/libyalink"
SERVICE_USER="libyalink"

echo ""
echo "╔══════════════════════════════════════════════════════════╗"
echo "║  LibyaLink Server — Quick Setup for Ubuntu 22.04        ║"
echo "║  Powered by Hysteria 2                                  ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "❌ This script must be run as root (use sudo)"
    exit 1
fi

# 1. Create system user
echo "[1/6] Creating system user..."
if ! id "$SERVICE_USER" &>/dev/null; then
    useradd -r -s /usr/sbin/nologin -d /etc/libyalink "$SERVICE_USER"
    echo "  ✅ User '$SERVICE_USER' created"
else
    echo "  ✅ User '$SERVICE_USER' already exists"
fi

# 2. Create config directory
echo "[2/6] Setting up config directory..."
mkdir -p "$CONFIG_DIR"
chown "$SERVICE_USER:$SERVICE_USER" "$CONFIG_DIR"
chmod 750 "$CONFIG_DIR"
echo "  ✅ Config directory: $CONFIG_DIR"

# 3. Generate self-signed certificate
echo "[3/6] Generating self-signed TLS certificate..."
if [ ! -f "$CONFIG_DIR/cert.pem" ]; then
    openssl req -x509 -nodes -newkey ec:<(openssl ecparam -name prime256v1) \
        -keyout "$CONFIG_DIR/key.pem" \
        -out "$CONFIG_DIR/cert.pem" \
        -days 3650 \
        -subj "/CN=libyalink"
    chown "$SERVICE_USER:$SERVICE_USER" "$CONFIG_DIR/cert.pem" "$CONFIG_DIR/key.pem"
    chmod 640 "$CONFIG_DIR/key.pem"
    echo "  ✅ Self-signed certificate generated (valid 10 years)"
else
    echo "  ✅ Certificate already exists, skipping"
fi

# 4. Create default config if not exists
echo "[4/6] Creating default config..."
if [ ! -f "$CONFIG_DIR/config.yaml" ]; then
    # Generate random password
    AUTH_PASS=$(openssl rand -base64 24)

    cat > "$CONFIG_DIR/config.yaml" << YAML
# LibyaLink Server Configuration
# Powered by Hysteria 2
# Generated: $(date -u +%Y-%m-%dT%H:%M:%SZ)

listen: :443

tls:
  cert: /etc/libyalink/cert.pem
  key: /etc/libyalink/key.pem

auth:
  type: password
  password: "${AUTH_PASS}"

# Bandwidth limits (optional — remove to disable)
# bandwidth:
#   up: 100 mbps
#   down: 100 mbps

# Masquerade as a normal HTTPS site
masquerade:
  type: string
  content: "404 Not Found"
  statusCode: 404
YAML

    chown "$SERVICE_USER:$SERVICE_USER" "$CONFIG_DIR/config.yaml"
    chmod 640 "$CONFIG_DIR/config.yaml"
    echo "  ✅ Default config created"
    echo ""
    echo "  ╔══════════════════════════════════════════════════════╗"
    echo "  ║  YOUR AUTO-GENERATED PASSWORD:                      ║"
    echo "  ║  $AUTH_PASS  ║"
    echo "  ║  Save this! You'll need it for client configs.      ║"
    echo "  ╚══════════════════════════════════════════════════════╝"
    echo ""
else
    echo "  ✅ Config already exists, skipping"
fi

# 5. Apply network tuning
echo "[5/6] Applying network tuning..."
cat > /etc/sysctl.d/99-libyalink.conf << 'SYSCTL'
# LibyaLink / Hysteria 2 UDP & network tuning
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
net.core.rmem_default = 8388608
net.core.wmem_default = 8388608
net.ipv4.udp_mem = 65536 131072 262144
net.core.netdev_max_backlog = 10000
net.core.default_qdisc = fq
net.ipv4.tcp_congestion_control = bbr
SYSCTL
sysctl --system > /dev/null 2>&1
echo "  ✅ Sysctl tuning applied (8MB UDP buffers, BBR enabled)"

# 6. Install systemd service
echo "[6/6] Installing systemd service..."
if [ -f "$(dirname "$0")/../docs/libyalink.service" ]; then
    cp "$(dirname "$0")/../docs/libyalink.service" /etc/systemd/system/libyalink.service
elif [ -f "/tmp/libyalink.service" ]; then
    cp /tmp/libyalink.service /etc/systemd/system/libyalink.service
else
    cat > /etc/systemd/system/libyalink.service << 'UNIT'
[Unit]
Description=LibyaLink Server (Powered by Hysteria 2)
After=network.target network-online.target
Wants=network-online.target

[Service]
Type=simple
User=libyalink
Group=libyalink
ExecStart=/usr/local/bin/libyalink server -c /etc/libyalink/config.yaml
Restart=always
RestartSec=5
LimitNOFILE=65535
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
PrivateTmp=true
ReadWritePaths=/etc/libyalink
AmbientCapabilities=CAP_NET_BIND_SERVICE
CapabilityBoundingSet=CAP_NET_BIND_SERVICE
Environment="HYSTERIA_LOG_LEVEL=info"
StandardOutput=journal
StandardError=journal
SyslogIdentifier=libyalink

[Install]
WantedBy=multi-user.target
UNIT
fi

systemctl daemon-reload
echo "  ✅ Systemd service installed"

# Open firewall if UFW is active
if command -v ufw &>/dev/null && ufw status | grep -q "active"; then
    ufw allow 443/udp comment "LibyaLink Hysteria 2" > /dev/null 2>&1
    echo "  ✅ UFW: UDP 443 opened"
fi

echo ""
echo "══════════════════════════════════════════════════════════"
echo ""
echo "  Setup complete! Next steps:"
echo ""
echo "  1. Copy the libyalink binary to $INSTALL_DIR/"
echo "  2. Start the server:"
echo "       sudo systemctl enable --now libyalink"
echo ""
echo "  3. Check status:"
echo "       sudo systemctl status libyalink"
echo "       sudo journalctl -u libyalink -f"
echo ""
echo "  4. Run diagnostics:"
echo "       libyalink doctor -c /etc/libyalink/config.yaml"
echo ""
echo "  5. Generate client config:"
echo "       SERVER_IP=\$(curl -4 -s ifconfig.me)"
echo "       libyalink gen-client --server \$SERVER_IP --auth \"YOUR_PASSWORD\" --insecure"
echo ""
echo "══════════════════════════════════════════════════════════"
echo ""
