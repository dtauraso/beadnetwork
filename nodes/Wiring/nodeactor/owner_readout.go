package nodeactor

func (r *pairReadout) SetRoundsToParallel(rounds, msgs int32) {
	r.roundsToParallel = rounds
	r.msgsToParallel = msgs
}

func (r *pairReadout) RoundsToParallel() (rounds, msgs int32) {
	return r.roundsToParallel, r.msgsToParallel
}
