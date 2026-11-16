package Builder

const (
	OpFrame  = "frame"
	OpTickMs = "tick-ms"

	FrameRho    = "rho"
	FrameTphi   = "tphi"
	FrameTtheta = "ttheta"
	FrameCr     = "cr"
	FrameS      = "s"
)

func frameAct(k string) Act { return Act{Op: OpFrame, K: k} }
