package owners

type Readout struct {
	roundsToParallel int32

	msgsToParallel int32
}

func (r *Readout) SetRoundsToParallel(rounds, msgs int32) {
	r.roundsToParallel = rounds
	r.msgsToParallel = msgs
}

func (r *Readout) RoundsToParallel() (rounds, msgs int32) {
	return r.roundsToParallel, r.msgsToParallel
}
