package TiltPanel

type TiltVectorMsg struct {
	PhiIdx int32

	ThetaIdx int32

	RIdx int32

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
	"NodePhi":      true,
	"NodePhiTheta": true,
}

func KindWantsVectorChannel(kind string) bool {
	return vectorCapableKinds[kind]
}

var tiltPanelKinds = map[string]bool{
	"NodePhi": true,
}

func KindDrivenByTiltPanel(kind string) bool {
	return tiltPanelKinds[kind]
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
