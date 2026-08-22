package Panel

import (
	"github.com/dtauraso/wirefold/src/Input/Codec"
	"github.com/dtauraso/wirefold/src/Input/Stdin"
)

func init() { Stdin.RegisterUpdateDecoder("panels", decodeUpdate) }

func decodeUpdate(r *Codec.Reader, attr byte) (Stdin.StdinMsg, bool) {
	if attr != attrToggle {
		return Stdin.StdinMsg{}, false
	}
	flagID, err := r.U8()
	if err != nil || int(flagID) >= len(Codec.InPanelFlags) {
		return Stdin.StdinMsg{}, false
	}
	return Stdin.StdinMsg{Type: "edit", Op: "update", Kind: "panels", Attr: "toggle", Flag: Codec.InPanelFlags[flagID]}, true
}
