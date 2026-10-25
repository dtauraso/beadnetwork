package Sheet

import "strconv"

type Edit struct {
	Active bool
	Act    Act
	Draft  string
}

func StartDraft(e *Edit, a Act) {
	*e = Edit{Active: true, Act: a}
}

func Key(e *Edit, key string) (Act, float64, bool) {
	if !e.Active {
		return Act{}, 0, false
	}
	switch key {
	case "Escape":
		*e = Edit{}
	case "Enter":
		a, draft := e.Act, e.Draft
		*e = Edit{}
		v, err := strconv.ParseFloat(draft, 64)
		if err != nil {
			return Act{}, 0, false
		}
		return a, v, true
	case "Backspace":
		if len(e.Draft) > 0 {
			e.Draft = e.Draft[:len(e.Draft)-1]
		}
	case "-":
		if e.Draft == "" {
			e.Draft = "-"
		}
	case ".":
		e.Draft += "."
	default:
		if len(key) == 1 && key[0] >= '0' && key[0] <= '9' {
			e.Draft += key
		}
	}
	return Act{}, 0, false
}
