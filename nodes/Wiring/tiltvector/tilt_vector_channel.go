package tiltvector

type TiltVectorMsg struct {
	ThetaIdx int32

	Points int32

	Reset bool

	Machine TiltMachine
}

type TiltMachine int8

const (
	TiltMachineNone TiltMachine = iota

	TiltMachinePerpendicular

	TiltMachineParallel
)

func (m TiltMachine) String() string {
	switch m {
	case TiltMachinePerpendicular:
		return "perpendicular"
	case TiltMachineParallel:
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

const PerpendicularThetaIdx int32 = 6

const HalfTurnThetaIdx = 2 * PerpendicularThetaIdx

const FullTurnThetaIdx = 2 * HalfTurnThetaIdx
