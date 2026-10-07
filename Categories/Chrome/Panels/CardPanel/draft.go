package CardPanel

import "strconv"

type Commit struct {
	Field Field
	Value int
}

func StartDraft(e *Edit, f Field) {
	*e = Edit{Active: true, Field: f}
}

func Key(e *Edit, key string) (Commit, bool) {
	if !e.Active {
		return Commit{}, false
	}
	switch key {
	case "Escape":
		*e = Edit{}
	case "Enter":
		f, draft := e.Field, e.Draft
		*e = Edit{}
		v, err := strconv.Atoi(draft)
		if err != nil || !Valid(f, v) {
			return Commit{}, false
		}
		return Commit{Field: f, Value: v}, true
	case "Backspace":
		if len(e.Draft) > 0 {
			e.Draft = e.Draft[:len(e.Draft)-1]
		}
	case "-":
		if e.Draft == "" {
			e.Draft = "-"
		}
	default:
		if len(key) == 1 && key[0] >= '0' && key[0] <= '9' {
			e.Draft += key
		}
	}
	return Commit{}, false
}
