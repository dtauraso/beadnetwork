package owners

import streamframe "github.com/dtauraso/wirefold/Categories/Scene/Vectors"

func (c *ChannelVectors) VectorsFrom(self Vec3) []streamframe.ChannelVector {
	if !c.on {
		return nil
	}
	peers := make(map[string]streamframe.Vec3, len(c.peerCenters))
	for id, p := range c.peerCenters {
		peers[id] = streamframe.Vec3(p)
	}
	return streamframe.ChannelVectorsFor(streamframe.Vec3(self), peers)
}

type ChannelVectors struct {
	sceneToNodeOn chan bool

	on bool

	peerCenters map[string]Vec3

	sentCenter    Vec3
	hasSentCenter bool
}

func (c *ChannelVectors) NeedsBroadcast(center Vec3) bool {
	if !c.on {
		return false
	}
	if c.hasSentCenter && c.sentCenter == center {
		return false
	}
	c.sentCenter, c.hasSentCenter = center, true
	return true
}

func (c *ChannelVectors) Forget() { c.hasSentCenter = false }

func (c *ChannelVectors) In() chan bool {
	if c.sceneToNodeOn == nil {
		c.sceneToNodeOn = make(chan bool, 1) // chan-name-ok: sceneToNodeOn names both ends
	}
	return c.sceneToNodeOn
}

func (c *ChannelVectors) TakeOn() (on, turnedOn bool) {
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

func (c *ChannelVectors) On() bool { return c.on }

func (c *ChannelVectors) SetPeerCenters(m map[string]Vec3) { c.peerCenters = m }

func (c *ChannelVectors) PeerCenters() map[string]Vec3 { return c.peerCenters }
