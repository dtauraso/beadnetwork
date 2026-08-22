package Panel

import (
	"github.com/dtauraso/wirefold/Input/Stdin"
)

func init() { Stdin.RegisterUpdateDecoder("panels", decodeUpdate) }

func decodeUpdate(payload []byte, attr byte) (Stdin.StdinMsg, bool) {
	r := NewReader(payload, 0)
	if attr != attrToggle {
		return Stdin.StdinMsg{}, false
	}
	flagID, err := r.U8()
	if err != nil || int(flagID) >= len(FlagNames) {
		return Stdin.StdinMsg{}, false
	}
	return Stdin.StdinMsg{Type: "edit", Op: "update", Kind: "panels", Attr: "toggle", Flag: FlagNames[flagID]}, true
}
