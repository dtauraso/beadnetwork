package CardPanel

import "strconv"

func StartDraft(e *Edit, f Field) {
	*e = Edit{Active: true, Field: f}
}

func Key(e *Edit, key string) (EditMsg, bool) {
	if !e.Active {
		return EditMsg{}, false
	}
	switch key {
	case "Escape":
		*e = Edit{}
	case "Enter":
		f, draft := e.Field, e.Draft
		*e = Edit{}
		v, err := strconv.Atoi(draft)
		if err != nil || !Valid(f, v) {
			return EditMsg{}, false
		}
		return EditMsg{Field: f, Value: v}, true
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
	return EditMsg{}, false
}
