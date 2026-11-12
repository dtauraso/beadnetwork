package Sheet

import (
	"fmt"
	"strconv"
	"unicode/utf8"
)

type Edit struct {
	Active bool
	Act    Act
	Draft  string
	Words  bool
}

type Typed struct {
	V    float64
	Text string
}

func StartDraft(e *Edit, c Cell) {
	*e = Edit{Active: true, Act: c.Act, Words: c.Words}
	if c.Words {
		e.Draft = c.Value
	}
}

func Key(e *Edit, key string) (Act, Typed, bool, error) {
	if !e.Active {
		return Act{}, Typed{}, false, nil
	}
	switch key {
	case "Escape":
		*e = Edit{}
	case "Enter":
		a, draft, words := e.Act, e.Draft, e.Words
		*e = Edit{}
		if words {
			return a, Typed{Text: draft}, true, nil
		}
		if draft == "" {
			return Act{}, Typed{}, false, nil
		}
		v, err := strconv.ParseFloat(draft, 64)
		if err != nil {
			return Act{}, Typed{}, false, fmt.Errorf("%q: not a number, so nothing changed", draft)
		}
		return a, Typed{V: v}, true, nil
	case "Backspace":
		if len(e.Draft) > 0 {
			_, n := utf8.DecodeLastRuneInString(e.Draft)
			e.Draft = e.Draft[:len(e.Draft)-n]
		}
	case "-":
		if e.Words || e.Draft == "" {
			e.Draft += "-"
		}
	case ".":
		e.Draft += "."
	default:
		if e.Words && utf8.RuneCountInString(key) == 1 {
			e.Draft += key
		} else if len(key) == 1 && key[0] >= '0' && key[0] <= '9' {
			e.Draft += key
		}
	}
	return Act{}, Typed{}, false, nil
}
