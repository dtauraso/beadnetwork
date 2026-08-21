package trace

const (
	KindRecv = "recv"
	KindFire = "fire"
	KindSend = "send"

	KindArrive = "arrive"

	KindBreadcrumb = "breadcrumb"
)

var TraceEventKinds = []string{KindRecv, KindFire, KindSend, KindArrive, KindBreadcrumb}
