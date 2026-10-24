package core

import (
	"os"
	"strings"

	"github.com/go-typeset/bidi"
	"github.com/orochaa/go-clack/third_party/picocolors"
)

func rtlEnabled() bool { return os.Getenv("CLACK_RTL") == "1" }

func isArabicBlock(r rune) bool {
	return (r >= 0x0600 && r <= 0x06FF) || (r >= 0x0750 && r <= 0x077F)
}

// visualWithCursor converts logical text to visual order and draws the cursor.
// Pass cursor < 0 to render without a cursor (e.g. the submitted state).
func VisualWithCursor(value string, cursor int) string {
	runes := []rune(value)
	if len(runes) == 0 {
		if cursor >= 0 {
			return "█"
		}
		return ""
	}
	if cursor > len(runes) {
		cursor = len(runes)
	}

	// 1. shape: logical letters -> joined presentation forms
	forms := bidi.JoinForms(runes)
	shaped := make([]rune, len(runes))
	for i, r := range runes {
		shaped[i] = r
		if isArabicBlock(r) {
			if pf := bidi.PresentationForm(r, forms[i]); pf != 0 {
				shaped[i] = pf
			}
		}
	}

	// 2. reorder: visual position -> logical index
	levels := bidi.ResolveLevels(runes, bidi.LeftToRight)
	order := bidi.Reorder(runes, levels)

	// 3. draw
	var sb strings.Builder
	last := len(runes) - 1
	for _, idx := range order {
		r := shaped[idx]

		// cursor "after the last char": which side that is depends on its direction
		endBlock := cursor == len(runes) && idx == last
		if endBlock && levels[idx]&1 == 1 {
			sb.WriteString("█") // RTL char: insertion point is on its visual left
		}

		switch {
		case r == 0x200C || r == 0x200D:
			// ZWNJ/ZWJ: joining is already applied; drawing them gives font boxes
			if idx == cursor {
				sb.WriteString(picocolors.Inverse(" "))
			}
		case idx == cursor:
			sb.WriteString(picocolors.Inverse(string(r)))
		default:
			sb.WriteRune(r)
		}

		if endBlock && levels[idx]&1 == 0 {
			sb.WriteString("█")
		}
	}
	return sb.String()
}

func Visual(s string) string {
	if !rtlEnabled() {
		return s
	}
	return VisualWithCursor(s, -1)
}
