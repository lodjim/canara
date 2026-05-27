package main

import (
	"os"
	"time"
)

// HID modifier bits
const (
	ModNone   byte = 0x00
	ModLCtrl  byte = 0x01
	ModLShift byte = 0x02
	ModLAlt   byte = 0x04
	ModLGUI   byte = 0x08
	ModRCtrl  byte = 0x10
	ModRShift byte = 0x20
	ModRAlt   byte = 0x40
	ModRGUI   byte = 0x80
)

// HID keycodes for special keys
const (
	KeyEnter      byte = 0x28
	KeyEscape     byte = 0x29
	KeyBackspace  byte = 0x2a
	KeyTab        byte = 0x2b
	KeySpace      byte = 0x2c
	KeyDelete     byte = 0x4c
	KeyHome       byte = 0x4a
	KeyEnd        byte = 0x4d
	KeyPageUp     byte = 0x4b
	KeyPageDown   byte = 0x4e
	KeyRight      byte = 0x4f
	KeyLeft       byte = 0x50
	KeyDown       byte = 0x51
	KeyUp         byte = 0x52
	KeyCapsLock   byte = 0x39
	KeyInsert     byte = 0x49
	KeyPrintScr   byte = 0x46
	KeyScrollLock byte = 0x47
	KeyPause      byte = 0x48
	KeyNumLock    byte = 0x53
	KeyF1         byte = 0x3a
	KeyF2         byte = 0x3b
	KeyF3         byte = 0x3c
	KeyF4         byte = 0x3d
	KeyF5         byte = 0x3e
	KeyF6         byte = 0x3f
	KeyF7         byte = 0x40
	KeyF8         byte = 0x41
	KeyF9         byte = 0x42
	KeyF10        byte = 0x43
	KeyF11        byte = 0x44
	KeyF12        byte = 0x45
)

// KeyStroke holds a modifier + keycode pair
type KeyStroke struct {
	Modifier byte
	Keycode  byte
}

// charToKey maps printable characters to their HID keystrokes (US layout)
var charToKey = map[rune]KeyStroke{
	// Lowercase
	'a': {ModNone, 0x04}, 'b': {ModNone, 0x05}, 'c': {ModNone, 0x06},
	'd': {ModNone, 0x07}, 'e': {ModNone, 0x08}, 'f': {ModNone, 0x09},
	'g': {ModNone, 0x0a}, 'h': {ModNone, 0x0b}, 'i': {ModNone, 0x0c},
	'j': {ModNone, 0x0d}, 'k': {ModNone, 0x0e}, 'l': {ModNone, 0x0f},
	'm': {ModNone, 0x10}, 'n': {ModNone, 0x11}, 'o': {ModNone, 0x12},
	'p': {ModNone, 0x13}, 'q': {ModNone, 0x14}, 'r': {ModNone, 0x15},
	's': {ModNone, 0x16}, 't': {ModNone, 0x17}, 'u': {ModNone, 0x18},
	'v': {ModNone, 0x19}, 'w': {ModNone, 0x1a}, 'x': {ModNone, 0x1b},
	'y': {ModNone, 0x1c}, 'z': {ModNone, 0x1d},
	// Uppercase
	'A': {ModLShift, 0x04}, 'B': {ModLShift, 0x05}, 'C': {ModLShift, 0x06},
	'D': {ModLShift, 0x07}, 'E': {ModLShift, 0x08}, 'F': {ModLShift, 0x09},
	'G': {ModLShift, 0x0a}, 'H': {ModLShift, 0x0b}, 'I': {ModLShift, 0x0c},
	'J': {ModLShift, 0x0d}, 'K': {ModLShift, 0x0e}, 'L': {ModLShift, 0x0f},
	'M': {ModLShift, 0x10}, 'N': {ModLShift, 0x11}, 'O': {ModLShift, 0x12},
	'P': {ModLShift, 0x13}, 'Q': {ModLShift, 0x14}, 'R': {ModLShift, 0x15},
	'S': {ModLShift, 0x16}, 'T': {ModLShift, 0x17}, 'U': {ModLShift, 0x18},
	'V': {ModLShift, 0x19}, 'W': {ModLShift, 0x1a}, 'X': {ModLShift, 0x1b},
	'Y': {ModLShift, 0x1c}, 'Z': {ModLShift, 0x1d},
	// Digits
	'1': {ModNone, 0x1e}, '2': {ModNone, 0x1f}, '3': {ModNone, 0x20},
	'4': {ModNone, 0x21}, '5': {ModNone, 0x22}, '6': {ModNone, 0x23},
	'7': {ModNone, 0x24}, '8': {ModNone, 0x25}, '9': {ModNone, 0x26},
	'0': {ModNone, 0x27},
	// Symbols — no shift
	'-': {ModNone, 0x2d}, '=': {ModNone, 0x2e},
	'[': {ModNone, 0x2f}, ']': {ModNone, 0x30}, '\\': {ModNone, 0x31},
	';': {ModNone, 0x33}, '\'': {ModNone, 0x34}, '`': {ModNone, 0x35},
	',': {ModNone, 0x36}, '.': {ModNone, 0x37}, '/': {ModNone, 0x38},
	// Symbols — shift
	'!': {ModLShift, 0x1e}, '@': {ModLShift, 0x1f}, '#': {ModLShift, 0x20},
	'$': {ModLShift, 0x21}, '%': {ModLShift, 0x22}, '^': {ModLShift, 0x23},
	'&': {ModLShift, 0x24}, '*': {ModLShift, 0x25}, '(': {ModLShift, 0x26},
	')': {ModLShift, 0x27}, '_': {ModLShift, 0x2d}, '+': {ModLShift, 0x2e},
	'{': {ModLShift, 0x2f}, '}': {ModLShift, 0x30}, '|': {ModLShift, 0x31},
	':': {ModLShift, 0x33}, '"': {ModLShift, 0x34}, '~': {ModLShift, 0x35},
	'<': {ModLShift, 0x36}, '>': {ModLShift, 0x37}, '?': {ModLShift, 0x38},
	// Whitespace
	' ': {ModNone, KeySpace}, '\t': {ModNone, KeyTab}, '\n': {ModNone, KeyEnter},
}

// HIDKeyboard wraps the /dev/hidg0 device
type HIDKeyboard struct {
	dev      *os.File
	keyDelay time.Duration
}

func NewHIDKeyboard(devPath string) (*HIDKeyboard, error) {
	f, err := os.OpenFile(devPath, os.O_WRONLY, 0)
	if err != nil {
		return nil, err
	}
	return &HIDKeyboard{dev: f, keyDelay: 10 * time.Millisecond}, nil
}

func (k *HIDKeyboard) Close() {
	k.dev.Close()
}

var releaseReport = []byte{0, 0, 0, 0, 0, 0, 0, 0}

func (k *HIDKeyboard) SendKey(modifier, keycode byte) error {
	press := []byte{modifier, 0x00, keycode, 0x00, 0x00, 0x00, 0x00, 0x00}
	if _, err := k.dev.Write(press); err != nil {
		return err
	}
	time.Sleep(k.keyDelay)
	if _, err := k.dev.Write(releaseReport); err != nil {
		return err
	}
	time.Sleep(k.keyDelay)
	return nil
}

func (k *HIDKeyboard) TypeString(s string) error {
	keys := GetLayout().Keys
	for _, ch := range s {
		ks, ok := keys[ch]
		if !ok {
			continue
		}
		if err := k.SendKey(ks.Modifier, ks.Keycode); err != nil {
			return err
		}
	}
	return nil
}
