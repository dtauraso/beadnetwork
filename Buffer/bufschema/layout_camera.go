package bufschema

type bufLayoutCamera struct {
	PX       float32 `buf:"f32"`
	PY       float32 `buf:"f32"`
	PZ       float32 `buf:"f32"`
	R        float32 `buf:"f32"`
	PosTheta float32 `buf:"f32"`
	PosPhi   float32 `buf:"f32"`
	UpTheta  float32 `buf:"f32"`
	UpPhi    float32 `buf:"f32"`
}
