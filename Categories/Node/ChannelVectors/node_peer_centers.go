package ChannelVectors

func (c *PeerCenters) VectorsFrom(self Vec3) []ChannelVector {
	if !c.on {
		return nil
	}
	peers := make(map[string]Vec3, len(c.peerCenters))
	for id, p := range c.peerCenters {
		peers[id] = Vec3(p)
	}
	return ChannelVectorsFor(Vec3(self), peers)
}

type PeerCenters struct {
	sceneToNodeOn chan bool

	on bool

	peerCenters map[string]Vec3

	sentCenter    Vec3
	hasSentCenter bool
}

func (c *PeerCenters) NeedsBroadcast(center Vec3) bool {
	if !c.on {
		return false
	}
	if c.hasSentCenter && c.sentCenter == center {
		return false
	}
	c.sentCenter, c.hasSentCenter = center, true
	return true
}

func (c *PeerCenters) Forget() { c.hasSentCenter = false }

func (c *PeerCenters) In() chan bool {
	if c.sceneToNodeOn == nil {
		c.sceneToNodeOn = make(chan bool, 1) // chan-name-ok: sceneToNodeOn names both ends
	}
	return c.sceneToNodeOn
}

func (c *PeerCenters) TakeOn() (on, turnedOn bool) {
	if c.sceneToNodeOn == nil {
		return c.on, false
	}
	select {
	case next := <-c.sceneToNodeOn:
		turnedOn = next && !c.on
		c.on = next
	default:
	}
	return c.on, turnedOn
}

func (c *PeerCenters) On() bool { return c.on }

func (c *PeerCenters) SetPeerCenters(m map[string]Vec3) { c.peerCenters = m }

func (c *PeerCenters) PeerCenters() map[string]Vec3 { return c.peerCenters }
