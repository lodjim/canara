package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	hidDev := flag.String("hid", "/dev/hidg0", "HID device path")
	port := flag.String("port", "8080", "Web server port")
	scriptsDir := flag.String("scripts", "./scripts", "Scripts directory")
	noWifi := flag.Bool("no-wifi", false, "Skip WiFi AP setup (for local testing)")
	ssid := flag.String("ssid", "DuckyPad", "WiFi SSID")
	pass := flag.String("pass", "duckypad123", "WiFi password (min 8 chars)")
	flag.Parse()

	if err := os.MkdirAll(*scriptsDir, 0755); err != nil {
		log.Fatalf("cannot create scripts dir: %v", err)
	}

	if !*noWifi {
		fmt.Println("[*] Setting up WiFi AP...")
		if err := setupWifi(*ssid, *pass); err != nil {
			log.Printf("[!] WiFi setup failed: %v", err)
			log.Println("[!] Continuing without WiFi AP. Use --no-wifi to suppress this.")
		} else {
			fmt.Printf("[+] WiFi AP running: SSID=%s  Password=%s\n", *ssid, *pass)
			fmt.Printf("[+] Connect your phone then open: http://192.168.4.1:%s\n", *port)
		}
	} else {
		fmt.Printf("[*] No-WiFi mode. Web UI: http://0.0.0.0:%s\n", *port)
	}

	fmt.Printf("[*] HID device: %s\n", *hidDev)
	fmt.Printf("[*] Scripts dir: %s\n", *scriptsDir)
	fmt.Printf("[*] Starting web server on :%s\n", *port)
	startWebServer(*port, *scriptsDir, *hidDev)
}
