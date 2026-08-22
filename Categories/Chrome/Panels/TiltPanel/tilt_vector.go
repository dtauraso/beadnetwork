package TiltPanel

type TiltVectorMsg struct {
	PhiIdx int32

	Points int32

	Reset bool

	Machine TiltMachine
}

type TiltMachine int8

const (
	TiltMachineNone TiltMachine = iota

	TiltMachineParallel
)

func (m TiltMachine) String() string {
	if m == TiltMachineParallel {
		return "parallel"
	}
	return "none"
}

var vectorCapableKinds = map[string]bool{
	"PairNode": true,
}

func KindWantsVectorChannel(kind string) bool {
	return vectorCapableKinds[kind]
}

func SendVectorLatestNonBlocking(ch chan<- TiltVectorMsg, v TiltVectorMsg) {
	if ch == nil {
		return
	}
	select {
	case ch <- v:
		return
	default:
	}

}

func PollRecvVector(ch <-chan TiltVectorMsg) (TiltVectorMsg, bool) {
	if ch == nil {
		return TiltVectorMsg{}, false
	}
	select {
	case v := <-ch:
		return v, true
	default:
		return TiltVectorMsg{}, false
	}
}

const QuarterTurnPhiIdx int32 = 6

const HalfTurnPhiIdx = 2 * QuarterTurnPhiIdx

const FullTurnPhiIdx = 2 * HalfTurnPhiIdx
