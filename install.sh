#!/bin/bash
# canara install script — run as root on the Pi Zero 2W
set -e

BINARY="./canara-arm64"
INSTALL_BIN="/usr/local/bin/canara"
SCRIPTS_DIR="/var/lib/canara/scripts"
SERVICE_SRC="./canara.service"
SERVICE_DST="/etc/systemd/system/canara.service"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
info()  { echo -e "${GREEN}[+]${NC} $*"; }
warn()  { echo -e "${YELLOW}[!]${NC} $*"; }
error() { echo -e "${RED}[✗]${NC} $*"; exit 1; }

[ "$(id -u)" -eq 0 ] || error "Run as root: sudo bash install.sh"
[ -f "$BINARY" ]     || error "Binary not found: $BINARY  (build with: GOOS=linux GOARCH=arm64 go build -o canara-arm64 .)"

info "Installing dependencies (hostapd, dnsmasq)..."
apt-get install -y hostapd dnsmasq > /dev/null

# Prevent distro from auto-managing these — canara starts them directly
systemctl stop    hostapd dnsmasq 2>/dev/null || true
systemctl disable hostapd dnsmasq 2>/dev/null || true
systemctl unmask  hostapd         2>/dev/null || true

# Stop any previous instance before replacing the binary
systemctl stop canara 2>/dev/null || true

info "Installing binary to $INSTALL_BIN..."
install -m 755 "$BINARY" "$INSTALL_BIN"

info "Creating scripts directory $SCRIPTS_DIR..."
mkdir -p "$SCRIPTS_DIR"

info "Installing systemd service..."
cp "$SERVICE_SRC" "$SERVICE_DST"
systemctl daemon-reload
systemctl enable canara
systemctl start  canara

echo ""
info "canara is running!"
echo ""
echo "   WiFi SSID : canara"
echo "   Password  : canara123"
echo "   Web UI    : http://192.168.4.1:8080"
echo ""
echo "   Useful commands:"
echo "   systemctl status canara      — check status"
echo "   journalctl -fu canara        — live logs"
echo "   systemctl restart canara     — restart"
echo "   systemctl stop canara        — stop"
echo ""
warn "To change SSID/password edit $SERVICE_DST then: systemctl daemon-reload && systemctl restart canara"
