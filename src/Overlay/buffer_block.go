package Overlay

var _ = bufLayoutOverlay{}

type bufLayoutOverlay struct {

	DragNodeRow int32 `buf:"i32"`

	EditRefused uint32 `buf:"u32"`

	SceneEditable uint8 `buf:"u8"`

	SceneKinds uint32 `buf:"u32"`

	Speed float32 `buf:"f32"`
}
