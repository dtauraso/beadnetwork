package FitButton

var _ = bufLayoutFitChip{}

type bufLayoutFitChip struct {
	X float32 `buf:"f32"`
	Y float32 `buf:"f32"`
	W float32 `buf:"f32"`
	H float32 `buf:"f32"`

	LabelText []byte `buf:"bytes"`
}
