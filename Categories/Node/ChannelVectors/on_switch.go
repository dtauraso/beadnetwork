package ChannelVectors

type OnSwitch struct {
	in map[string]chan bool
}

func (s *OnSwitch) ClaimChannelVectorsIn(id string, ch chan bool) {
	if s.in == nil {
		s.in = map[string]chan bool{}
	}
	s.in[id] = ch
}

func (s *OnSwitch) BroadcastChannelVectorsOn(on bool) {
	for _, ch := range s.in {
		select {
		case <-ch:
		default:
		}
		select {
		case ch <- on:
		default:
		}
	}
}
