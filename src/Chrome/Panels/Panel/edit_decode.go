package Panel

import "github.com/dtauraso/wirefold/src/Input/Codec"

func init() { Codec.RegisterUpdateDecoder("panels", decodeUpdate) }

func decodeUpdate(r *Codec.Reader, attr byte) (Codec.StdinMsg, bool) {
	if attr != attrToggle {
		return Codec.StdinMsg{}, false
	}
	flagID, err := r.U8()
	if err != nil || int(flagID) >= len(Codec.InPanelFlags) {
		return Codec.StdinMsg{}, false
	}
	return Codec.StdinMsg{Type: "edit", Op: "update", Kind: "panels", Attr: "toggle", Flag: Codec.InPanelFlags[flagID]}, true
}
