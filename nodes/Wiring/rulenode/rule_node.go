package rulenode

import (
	"context"
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/nodefiles"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/owners"
	"github.com/dtauraso/wirefold/nodes/Wiring/rulemsg"
)

type EditKind uint8

const (
	EditActiveToggle EditKind = iota
	EditPhiToggle
	EditMaxTheta
)

type Edit struct {
	Kind     EditKind
	MaxTheta *float64
}

type State struct {
	Rule      *polar.DragRule
	Active    bool
	GroupID   int32
	GroupSize int32

	EdgeActive map[string]bool
}

type RuleNode struct {
	id          string
	persistRoot string

	mesh owners.RuleMesh

	rule   *polar.DragRule
	active bool

	edits  chan Edit
	ruleIn chan rulemsg.Msg

	edgeActive       map[string]bool
	toggleSelfToPeer map[string]chan bool
	toggleIn         chan EdgeToggle

	out  chan State
	wake chan struct{}
}

func New(id string) *RuleNode {
	return &RuleNode{
		id:               id,
		mesh:             owners.NewRuleMesh(),
		active:           true,
		edits:            make(chan Edit, 8),
		ruleIn:           make(chan rulemsg.Msg, 8),
		edgeActive:       map[string]bool{},
		toggleSelfToPeer: map[string]chan bool{},
		toggleIn:         make(chan EdgeToggle, 8),
		out:              make(chan State, 1),
		wake:             make(chan struct{}, 1),
	}
}

func (r *RuleNode) SetPersistRoot(root string) { r.persistRoot = root }

func (r *RuleNode) SeedRule(rule *polar.DragRule, active bool) {
	r.rule = rule
	r.active = active
	r.mesh.SetSelfRuleKey(rulemsg.KeyOf(rule))
}

func (r *RuleNode) RuleBackChannel(peerID string) chan rulemsg.Msg {
	return r.mesh.RuleBackChannel(peerID)
}

func (r *RuleNode) LinkRuleDown(peerID string, down chan rulemsg.Msg) {
	r.mesh.LinkRuleDown(peerID, down)
}

func (r *RuleNode) BroadcastSelf() {
	r.mesh.SetSelfRuleKey(rulemsg.KeyOf(r.rule))
	r.mesh.BroadcastRule(r.id)
}

func (r *RuleNode) Edits() chan<- Edit { return r.edits }

func (r *RuleNode) Out() <-chan State { return r.out }

func (r *RuleNode) Wake() <-chan struct{} { return r.wake }

func (r *RuleNode) Run(ctx context.Context) {
	for peerID, back := range r.mesh.BackChannels() {
		go r.forward(ctx, peerID, back)
	}
	for target, toggle := range r.toggleSelfToPeer {
		go r.forwardToggle(ctx, target, toggle)
	}

	r.publish()

	for {
		select {
		case <-ctx.Done():
			return
		case e := <-r.edits:
			r.applyEdit(e)
			r.publish()
		case msg := <-r.ruleIn:
			if r.mesh.ApplyPeerRule(msg) {
				r.publish()
			}
		case t := <-r.toggleIn:
			r.applyEdgeToggle(t)
			r.publish()
		}
	}
}

func (r *RuleNode) forward(ctx context.Context, peerID string, back chan rulemsg.Msg) {
	if back == nil {
		panic(fmt.Sprintf(
			"rulenode.forward: node %q has a nil back-channel for peer %q, so that peer's rule could never "+
				"arrive and the two would never group — the mesh was linked with a channel that was never made",
			r.id, peerID))
	}
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-back:
			select {
			case r.ruleIn <- msg:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (r *RuleNode) applyEdit(e Edit) {
	switch e.Kind {
	case EditActiveToggle:
		r.active = !r.active
		r.persistActive()
	case EditPhiToggle:
		var next polar.DragRule
		if r.rule != nil {
			next = *r.rule
		}
		if next.Phi != nil {
			next.Phi = nil
		} else {
			zero := 0.0
			next.Phi = &zero
		}
		r.rule = &next
		r.persistRule()
	case EditMaxTheta:
		var next polar.DragRule
		if r.rule != nil {
			next = *r.rule
		}
		next.MaxTheta = e.MaxTheta
		r.rule = &next
		r.persistRule()
	default:
		panic(fmt.Sprintf(
			"rulenode.applyEdit: node %q was sent edit kind %d, which no case handles, so a rule edit would be "+
				"silently dropped after crossing the bridge", r.id, e.Kind))
	}
	r.mesh.SetSelfRuleKey(rulemsg.KeyOf(r.rule))
	r.mesh.BroadcastRule(r.id)
}

func (r *RuleNode) publish() {
	groupID, groupSize := r.mesh.RuleGroup(r.id)
	edgeActive := make(map[string]bool, len(r.edgeActive))
	for target, active := range r.edgeActive {
		edgeActive[target] = active
	}
	state := State{
		Rule: r.rule, Active: r.active,
		GroupID: groupID, GroupSize: groupSize,
		EdgeActive: edgeActive,
	}

	select {
	case <-r.out:
	default:
	}
	select {
	case r.out <- state:
	default:
	}

	select {
	case r.wake <- struct{}{}:
	default:
	}
}

func (r *RuleNode) persistRule() {
	if r.persistRoot == "" {
		return
	}
	if err := nodefiles.WriteDragRule(r.persistRoot, r.id, r.rule); err != nil {
		jsonpersist.LogPersistErr("rulenode", r.id, err)
	}
}

func (r *RuleNode) persistActive() {
	if r.persistRoot == "" {
		return
	}
	if err := nodefiles.WriteDragActive(r.persistRoot, r.id, r.active); err != nil {
		jsonpersist.LogPersistErr("rulenode", r.id, err)
	}
}
