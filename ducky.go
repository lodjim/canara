package main

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// specialKeys maps DuckyScript key names to HID keystrokes
var specialKeys = map[string]KeyStroke{
	"ENTER":       {ModNone, KeyEnter},
	"ESCAPE":      {ModNone, KeyEscape},
	"ESC":         {ModNone, KeyEscape},
	"BACKSPACE":   {ModNone, KeyBackspace},
	"TAB":         {ModNone, KeyTab},
	"SPACE":       {ModNone, KeySpace},
	"DELETE":      {ModNone, KeyDelete},
	"DEL":         {ModNone, KeyDelete},
	"HOME":        {ModNone, KeyHome},
	"END":         {ModNone, KeyEnd},
	"PAGEUP":      {ModNone, KeyPageUp},
	"PAGEDOWN":    {ModNone, KeyPageDown},
	"UPARROW":     {ModNone, KeyUp},
	"UP":          {ModNone, KeyUp},
	"DOWNARROW":   {ModNone, KeyDown},
	"DOWN":        {ModNone, KeyDown},
	"LEFTARROW":   {ModNone, KeyLeft},
	"LEFT":        {ModNone, KeyLeft},
	"RIGHTARROW":  {ModNone, KeyRight},
	"RIGHT":       {ModNone, KeyRight},
	"CAPSLOCK":    {ModNone, KeyCapsLock},
	"INSERT":      {ModNone, KeyInsert},
	"PRINTSCREEN": {ModNone, KeyPrintScr},
	"SCROLLLOCK":  {ModNone, KeyScrollLock},
	"PAUSE":       {ModNone, KeyPause},
	"NUMLOCK":     {ModNone, KeyNumLock},
	"F1":          {ModNone, KeyF1},
	"F2":          {ModNone, KeyF2},
	"F3":          {ModNone, KeyF3},
	"F4":          {ModNone, KeyF4},
	"F5":          {ModNone, KeyF5},
	"F6":          {ModNone, KeyF6},
	"F7":          {ModNone, KeyF7},
	"F8":          {ModNone, KeyF8},
	"F9":          {ModNone, KeyF9},
	"F10":         {ModNone, KeyF10},
	"F11":         {ModNone, KeyF11},
	"F12":         {ModNone, KeyF12},
}

// modifierKeys maps DuckyScript modifier names to HID modifier bits
var modifierKeys = map[string]byte{
	"CTRL":    ModLCtrl,
	"CONTROL": ModLCtrl,
	"SHIFT":   ModLShift,
	"ALT":     ModLAlt,
	"GUI":     ModLGUI,
	"WINDOWS": ModLGUI,
	"WIN":     ModLGUI,
	"COMMAND": ModLGUI,
}

// ExecuteDucky parses and executes a DuckyScript string.
// logFn is called for each line executed so the caller can stream progress.
func ExecuteDucky(script string, kbd *HIDKeyboard, logFn func(string)) error {
	scanner := bufio.NewScanner(strings.NewReader(script))
	lineNum := 0
	defaultDelay := 0 // ms applied after every non-DELAY line

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		// Split command and argument once
		parts := strings.SplitN(line, " ", 2)
		cmd := strings.ToUpper(parts[0])
		arg := ""
		if len(parts) > 1 {
			arg = parts[1]
		}

		logFn(fmt.Sprintf("[%d] %s", lineNum, line))

		switch cmd {
		case "REM":
			// comment — skip

		case "DELAY":
			ms, err := strconv.Atoi(strings.TrimSpace(arg))
			if err != nil {
				return fmt.Errorf("line %d: bad DELAY value %q", lineNum, arg)
			}
			time.Sleep(time.Duration(ms) * time.Millisecond)
			continue // don't apply defaultDelay after an explicit DELAY

		case "DEFAULTDELAY", "DEFAULT_DELAY":
			ms, err := strconv.Atoi(strings.TrimSpace(arg))
			if err != nil {
				return fmt.Errorf("line %d: bad DEFAULT_DELAY value %q", lineNum, arg)
			}
			defaultDelay = ms
			continue

		case "STRING":
			if err := kbd.TypeString(arg); err != nil {
				return fmt.Errorf("line %d: STRING: %w", lineNum, err)
			}

		case "STRINGLN":
			if err := kbd.TypeString(arg + "\n"); err != nil {
				return fmt.Errorf("line %d: STRINGLN: %w", lineNum, err)
			}

		default:
			// Key combo: one or more modifiers + one key
			// e.g. "GUI R", "CTRL ALT DELETE", "ENTER", "F5"
			if err := executeKeyCombo(kbd, line); err != nil {
				return fmt.Errorf("line %d: %w", lineNum, err)
			}
		}

		if defaultDelay > 0 {
			time.Sleep(time.Duration(defaultDelay) * time.Millisecond)
		}
	}

	return scanner.Err()
}

// executeKeyCombo handles lines like "GUI R", "CTRL ALT DELETE", "ENTER", etc.
func executeKeyCombo(kbd *HIDKeyboard, line string) error {
	tokens := strings.Fields(strings.ToUpper(line))

	var modifier byte
	var keycode byte
	keyFound := false

	for _, token := range tokens {
		if mod, ok := modifierKeys[token]; ok {
			modifier |= mod
			continue
		}
		if ks, ok := specialKeys[token]; ok {
			keycode = ks.Keycode
			keyFound = true
			continue
		}
		// Single character (e.g. "R" in "GUI R")
		if len(token) == 1 {
			r := rune(token[0])
			if r >= 'A' && r <= 'Z' {
				r += 32 // to lowercase — modifier handles shift if needed
			}
			if ks, ok := charToKey[r]; ok {
				keycode = ks.Keycode
				keyFound = true
			}
		}
	}

	if !keyFound && modifier == 0 {
		return fmt.Errorf("unknown key combo: %q", line)
	}

	return kbd.SendKey(modifier, keycode)
}
