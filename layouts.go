package main

import "sync"

// Layout maps printable runes to their HID keystroke for a given keyboard layout
type Layout struct {
	Code    string
	Display string
	Keys    map[rune]KeyStroke
}

var (
	currentLayout *Layout
	layoutMu      sync.RWMutex
)

// LayoutOrder defines display order in the UI
var LayoutOrder = []string{"us", "uk", "fr", "de", "es", "it"}

// Layouts is the registry of all supported keyboard layouts
var Layouts = map[string]*Layout{
	"us": {Code: "us", Display: "US QWERTY", Keys: buildUS()},
	"uk": {Code: "uk", Display: "UK QWERTY", Keys: buildUK()},
	"fr": {Code: "fr", Display: "FR AZERTY", Keys: buildFR()},
	"de": {Code: "de", Display: "DE QWERTZ", Keys: buildDE()},
	"es": {Code: "es", Display: "ES QWERTY", Keys: buildES()},
	"it": {Code: "it", Display: "IT QWERTY", Keys: buildIT()},
}

func init() { currentLayout = Layouts["us"] }

func GetLayout() *Layout {
	layoutMu.RLock()
	defer layoutMu.RUnlock()
	return currentLayout
}

func SetLayout(code string) bool {
	l, ok := Layouts[code]
	if !ok {
		return false
	}
	layoutMu.Lock()
	currentLayout = l
	layoutMu.Unlock()
	return true
}

func copyKeys(src map[rune]KeyStroke) map[rune]KeyStroke {
	m := make(map[rune]KeyStroke, len(src))
	for k, v := range src {
		m[k] = v
	}
	return m
}

// ── US QWERTY (base) ─────────────────────────────────────────────────────────

func buildUS() map[rune]KeyStroke {
	return copyKeys(charToKey)
}

// ── UK QWERTY ─────────────────────────────────────────────────────────────────
// " @ # ~ \ | differ from US

func buildUK() map[rune]KeyStroke {
	m := copyKeys(charToKey)
	m['"']  = KeyStroke{ModLShift, 0x1f} // SHIFT+2
	m['@']  = KeyStroke{ModLShift, 0x34} // SHIFT+'
	m['#']  = KeyStroke{ModNone, 0x32}   // ISO extra key (right of ')
	m['~']  = KeyStroke{ModLShift, 0x32} // SHIFT+#
	m['\\'] = KeyStroke{ModNone, 0x64}   // ISO key left of Z
	m['|']  = KeyStroke{ModLShift, 0x64}
	return m
}

// ── FR AZERTY ─────────────────────────────────────────────────────────────────
// Letters: a↔q, z↔w, m→; — Numbers: all need SHIFT — Symbols: remapped

func buildFR() map[rune]KeyStroke {
	m := copyKeys(charToKey)

	// Letter positions
	m['a'] = KeyStroke{ModNone, 0x14}   // A is at Q position
	m['A'] = KeyStroke{ModLShift, 0x14}
	m['z'] = KeyStroke{ModNone, 0x1a}   // Z is at W position
	m['Z'] = KeyStroke{ModLShift, 0x1a}
	m['q'] = KeyStroke{ModNone, 0x04}   // Q is at A position
	m['Q'] = KeyStroke{ModLShift, 0x04}
	m['w'] = KeyStroke{ModNone, 0x1d}   // W is at Z position
	m['W'] = KeyStroke{ModLShift, 0x1d}
	m['m'] = KeyStroke{ModNone, 0x33}   // M is at ; position
	m['M'] = KeyStroke{ModLShift, 0x33}

	// Digits need SHIFT on AZERTY
	m['1'] = KeyStroke{ModLShift, 0x1e}
	m['2'] = KeyStroke{ModLShift, 0x1f}
	m['3'] = KeyStroke{ModLShift, 0x20}
	m['4'] = KeyStroke{ModLShift, 0x21}
	m['5'] = KeyStroke{ModLShift, 0x22}
	m['6'] = KeyStroke{ModLShift, 0x23}
	m['7'] = KeyStroke{ModLShift, 0x24}
	m['8'] = KeyStroke{ModLShift, 0x25}
	m['9'] = KeyStroke{ModLShift, 0x26}
	m['0'] = KeyStroke{ModLShift, 0x27}

	// Unshifted number row produces these symbols
	m['&']  = KeyStroke{ModNone, 0x1e}
	m['"']  = KeyStroke{ModNone, 0x20}
	m['\''] = KeyStroke{ModNone, 0x21}
	m['(']  = KeyStroke{ModNone, 0x22}
	m['-']  = KeyStroke{ModNone, 0x23}
	m['_']  = KeyStroke{ModNone, 0x25}
	m[')']  = KeyStroke{ModNone, 0x2d}
	m['=']  = KeyStroke{ModNone, 0x2e}
	m['+']  = KeyStroke{ModLShift, 0x2e}

	// Bottom-row symbol remapping
	// On AZERTY: pos of M(US) → ,   pos of ,(US) → ;   pos of .(US) → :   pos of /(US) → !
	m[',']  = KeyStroke{ModNone, 0x10}   // , is at M position
	m['?']  = KeyStroke{ModLShift, 0x10} // ? is SHIFT+,
	m[';']  = KeyStroke{ModNone, 0x36}   // ; is at , position
	m['.']  = KeyStroke{ModLShift, 0x36} // . is SHIFT+;
	m[':']  = KeyStroke{ModNone, 0x37}   // : is at . position
	m['/']  = KeyStroke{ModLShift, 0x37} // / is SHIFT+:
	m['!']  = KeyStroke{ModNone, 0x38}   // ! is at / position (unshifted)

	return m
}

// ── DE QWERTZ ─────────────────────────────────────────────────────────────────
// Y↔Z swapped; shifted number row differs; - and _ remapped

func buildDE() map[rune]KeyStroke {
	m := copyKeys(charToKey)

	// Y and Z are swapped vs US
	m['y'] = KeyStroke{ModNone, 0x1d}
	m['Y'] = KeyStroke{ModLShift, 0x1d}
	m['z'] = KeyStroke{ModNone, 0x1c}
	m['Z'] = KeyStroke{ModLShift, 0x1c}

	// Shifted number row symbols differ
	m['"']  = KeyStroke{ModLShift, 0x1f} // SHIFT+2
	m['§']  = KeyStroke{ModLShift, 0x20} // SHIFT+3
	m['&']  = KeyStroke{ModLShift, 0x23} // SHIFT+6
	m['/']  = KeyStroke{ModLShift, 0x24} // SHIFT+7
	m['(']  = KeyStroke{ModLShift, 0x25} // SHIFT+8
	m[')']  = KeyStroke{ModLShift, 0x26} // SHIFT+9
	m['=']  = KeyStroke{ModLShift, 0x27} // SHIFT+0
	m['?']  = KeyStroke{ModLShift, 0x2d} // SHIFT+ß position

	// Bottom row
	m['-']  = KeyStroke{ModNone, 0x38}   // - is at / position on US
	m['_']  = KeyStroke{ModLShift, 0x38}
	m[':']  = KeyStroke{ModLShift, 0x37} // : is SHIFT+.
	m[';']  = KeyStroke{ModLShift, 0x36} // ; is SHIFT+,

	return m
}

// ── ES QWERTY (Spanish) ───────────────────────────────────────────────────────
// Letters same as US; ñ at ;; some symbol shifts differ

func buildES() map[rune]KeyStroke {
	m := copyKeys(charToKey)
	// ñ replaces the ; key
	m['"']  = KeyStroke{ModLShift, 0x1f} // SHIFT+2
	m['@']  = KeyStroke{ModLShift, 0x34}
	m['#']  = KeyStroke{ModNone, 0x32}
	m['\\'] = KeyStroke{ModNone, 0x64}
	m['|']  = KeyStroke{ModLShift, 0x64}
	return m
}

// ── IT QWERTY (Italian) ───────────────────────────────────────────────────────
// Letters same as US; minor symbol differences

func buildIT() map[rune]KeyStroke {
	m := copyKeys(charToKey)
	m['"']  = KeyStroke{ModLShift, 0x1f} // SHIFT+2
	m['@']  = KeyStroke{ModLShift, 0x34}
	m['#']  = KeyStroke{ModLShift, 0x31}
	m['-']  = KeyStroke{ModNone, 0x38}
	m['_']  = KeyStroke{ModLShift, 0x38}
	m['\''] = KeyStroke{ModNone, 0x2d}
	m['?']  = KeyStroke{ModLShift, 0x2d}
	return m
}
