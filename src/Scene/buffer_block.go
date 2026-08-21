package Scene

var _ = bufLayoutScene{}

type bufLayoutScene struct {
	NodeCount int32 `buf:"i32"`
	EdgeCount int32 `buf:"i32"`
}
