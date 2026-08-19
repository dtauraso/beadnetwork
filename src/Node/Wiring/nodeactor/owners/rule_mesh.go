package owners

import (
	"fmt"
	"github.com/dtauraso/wirefold/src/Node/spatial"
	"github.com/dtauraso/wirefold/src/PolarRulesPanel"
	"strconv"
)

type RuleMesh struct {
	backFromPeer map[string]chan PolarRulesPanel.Msg

	downToPeer map[string]chan PolarRulesPanel.Msg

	peerKey map[string]PolarRulesPanel.Key

	selfKey PolarRulesPanel.Key

	peerCenter map[string]spatial.Vec3

	selfCenter spatial.Vec3
}

func NewRuleMesh() RuleMesh {
	return RuleMesh{
		backFromPeer: map[string]chan PolarRulesPanel.Msg{},
		downToPeer:   map[string]chan PolarRulesPanel.Msg{},
		peerKey:      map[string]PolarRulesPanel.Key{},
		peerCenter:   map[string]spatial.Vec3{},
	}
}

func (r *RuleMesh) SetSelfCenter(c spatial.Vec3) { r.selfCenter = c }

func (r *RuleMesh) BroadcastCenter(selfID string) {
	msg := PolarRulesPanel.Msg{FromID: selfID, Key: r.selfKey, Center: r.selfCenter, HasCenter: true}
	for _, down := range r.downToPeer {
		select {
		case <-down:
		default:
		}
		select {
		case down <- msg:
		default:
		}
	}
}

func (r *RuleMesh) PeerCenters() map[string]spatial.Vec3 { return r.peerCenter }

func (r *RuleMesh) RuleBackChannel(peerID string) chan PolarRulesPanel.Msg {
	ch, ok := r.backFromPeer[peerID]
	if !ok {
		ch = make(chan PolarRulesPanel.Msg, 1) // chan-name-ok: ch IS backFromPeer[peerID]
		r.backFromPeer[peerID] = ch
	}
	return ch
}

func (r *RuleMesh) LinkRuleDown(peerID string, down chan PolarRulesPanel.Msg) {
	r.downToPeer[peerID] = down
}

func (r *RuleMesh) SetSelfRuleKey(key PolarRulesPanel.Key) { r.selfKey = key }

func (r *RuleMesh) BroadcastRule(selfID string) {
	msg := PolarRulesPanel.Msg{FromID: selfID, Key: r.selfKey}
	for _, down := range r.downToPeer {
		select {
		case <-down:
		default:
		}
		select {
		case down <- msg:
		default:
		}
	}
}

func (r *RuleMesh) BackChannels() map[string]chan PolarRulesPanel.Msg {
	return r.backFromPeer
}

func (r *RuleMesh) ApplyPeerRule(msg PolarRulesPanel.Msg) bool {
	changed := false
	if msg.HasCenter {
		if prev, seen := r.peerCenter[msg.FromID]; !seen || prev != msg.Center {
			r.peerCenter[msg.FromID] = msg.Center
			changed = true
		}
	}
	prev, seen := r.peerKey[msg.FromID]
	if seen && prev == msg.Key {
		return changed
	}
	r.peerKey[msg.FromID] = msg.Key
	return true
}

func (r *RuleMesh) DrainRules() bool {
	changed := false
	for peerID, back := range r.backFromPeer {
		select {
		case msg := <-back:
			if prev, seen := r.peerKey[peerID]; !seen || prev != msg.Key {
				r.peerKey[peerID] = msg.Key
				changed = true
			}
			if msg.HasCenter {
				if prev, seen := r.peerCenter[peerID]; !seen || prev != msg.Center {
					r.peerCenter[peerID] = msg.Center
					changed = true
				}
			}
		default:
		}
	}
	return changed
}

func (r *RuleMesh) RuleGroup(selfID string) (groupID, size int32) {
	groupID = nodeIDNumber(selfID)
	size = 1
	for peerID, key := range r.peerKey {
		if key != r.selfKey {
			continue
		}
		size++
		if n := nodeIDNumber(peerID); n < groupID {
			groupID = n
		}
	}
	return groupID, size
}

func nodeIDNumber(id string) int32 {
	n, err := strconv.Atoi(id)
	if err != nil {
		panic(fmt.Sprintf(
			"owners.RuleMesh: node id %q is not a number, so the rule group has no lowest member to agree on — "+
				"ids are numbers that happen to be directory names (loadTree parses every one with strconv.Atoi "+
				"and fails the load otherwise), so a non-numeric id here means something built a node outside the loader",
			id))
	}
	return int32(n)
}
