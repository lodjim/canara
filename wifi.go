package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

func setupWifi(ssid, password string) error {
	hostapdConf := fmt.Sprintf(`interface=wlan0
driver=nl80211
ssid=%s
hw_mode=g
channel=6
wmm_enabled=0
macaddr_acl=0
auth_algs=1
ignore_broadcast_ssid=0
wpa=2
wpa_passphrase=%s
wpa_key_mgmt=WPA-PSK
wpa_pairwise=TKIP
rsn_pairwise=CCMP
`, ssid, password)

	dnsmasqConf := `interface=wlan0
bind-interfaces
dhcp-range=192.168.4.2,192.168.4.20,255.255.255.0,24h
`

	// Stop anything that might be managing wlan0
	exec.Command("pkill", "-f", "hostapd").Run()
	exec.Command("pkill", "-f", "dnsmasq").Run()
	// If using NetworkManager, tell it to leave wlan0 alone
	exec.Command("nmcli", "dev", "set", "wlan0", "managed", "no").Run()
	time.Sleep(300 * time.Millisecond)

	if err := os.WriteFile("/tmp/duckypad-hostapd.conf", []byte(hostapdConf), 0644); err != nil {
		return fmt.Errorf("write hostapd.conf: %w", err)
	}
	if err := os.WriteFile("/tmp/duckypad-dnsmasq.conf", []byte(dnsmasqConf), 0644); err != nil {
		return fmt.Errorf("write dnsmasq.conf: %w", err)
	}

	// Flush and configure interface
	exec.Command("ip", "addr", "flush", "dev", "wlan0").Run()
	if out, err := exec.Command("ip", "link", "set", "wlan0", "up").CombinedOutput(); err != nil {
		return fmt.Errorf("ip link set wlan0 up: %v — %s", err, out)
	}
	if out, err := exec.Command("ip", "addr", "add", "192.168.4.1/24", "dev", "wlan0").CombinedOutput(); err != nil {
		return fmt.Errorf("ip addr add: %v — %s", err, out)
	}

	// Start hostapd in background
	go func() {
		cmd := exec.Command("hostapd", "/tmp/duckypad-hostapd.conf")
		if out, err := cmd.CombinedOutput(); err != nil {
			fmt.Printf("[!] hostapd stopped: %v: %s\n", err, out)
		}
	}()
	time.Sleep(time.Second) // give hostapd time to come up

	// Start dnsmasq in background
	go func() {
		cmd := exec.Command("dnsmasq", "--no-daemon", "--conf-file=/tmp/duckypad-dnsmasq.conf")
		if out, err := cmd.CombinedOutput(); err != nil {
			fmt.Printf("[!] dnsmasq stopped: %v: %s\n", err, out)
		}
	}()

	return nil
}
