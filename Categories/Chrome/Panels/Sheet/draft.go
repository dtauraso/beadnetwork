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
	At     int
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
	e.At = len(e.Draft)
}

func (e Edit) Caret() uint32 {
	return uint32(utf8.RuneCountInString(e.Draft[:min(max(e.At, 0), len(e.Draft))]))
}

func (e *Edit) insert(s string) {
	e.At = min(max(e.At, 0), len(e.Draft))
	e.Draft = e.Draft[:e.At] + s + e.Draft[e.At:]
	e.At += len(s)
}

func (e *Edit) move(key string) bool {
	e.At = min(max(e.At, 0), len(e.Draft))
	switch key {
	case "ArrowLeft":
		_, n := utf8.DecodeLastRuneInString(e.Draft[:e.At])
		e.At -= n
	case "ArrowRight":
		_, n := utf8.DecodeRuneInString(e.Draft[e.At:])
		e.At += n
	case "Home":
		e.At = 0
	case "End":
		e.At = len(e.Draft)
	case "Backspace":
		_, n := utf8.DecodeLastRuneInString(e.Draft[:e.At])
		e.Draft = e.Draft[:e.At-n] + e.Draft[e.At:]
		e.At -= n
	case "Delete":
		_, n := utf8.DecodeRuneInString(e.Draft[e.At:])
		e.Draft = e.Draft[:e.At] + e.Draft[e.At+n:]
	default:
		return false
	}
	return true
}

func Key(e *Edit, key string) (Act, Typed, bool, error) {
	if !e.Active {
		return Act{}, Typed{}, false, nil
	}
	if e.move(key) {
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
	case "-":
		if e.Words || e.At == 0 {
			e.insert("-")
		}
	case ".":
		e.insert(".")
	default:
		if e.Words && utf8.RuneCountInString(key) == 1 {
			e.insert(key)
		} else if len(key) == 1 && key[0] >= '0' && key[0] <= '9' {
			e.insert(key)
		}
	}
	return Act{}, Typed{}, false, nil
}
