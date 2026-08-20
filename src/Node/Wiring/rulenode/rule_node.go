package rulenode

import (
	"context"
	"fmt"
	"github.com/dtauraso/wirefold/src/Chrome/Panels/PolarRulesPanel"

	"github.com/dtauraso/wirefold/src/Node/Wiring/nodeactor/owners"
	"github.com/dtauraso/wirefold/src/spatial"
)

type EditKind uint8

const (
	EditActiveToggle EditKind = iota
	EditPhiToggle
	EditMaxTheta

	EditSelfActiveToggle
	EditSelfPhiToggle
	EditSelfMaxTheta

	EditRToggle
	EditSelfRToggle
)

type Edit struct {
	Kind     EditKind
	MaxTheta *float64
}

type State struct {
	Rule      *PolarRulesPanel.DragRule
	Active    bool
	GroupID   int32
	GroupSize int32

	SelfRule   *PolarRulesPanel.DragRule
	SelfActive bool

	EdgeActive map[string]bool

	KindActive bool

	PeerCenters map[string]spatial.Vec3
}

type RuleNode struct {
	id          string
	persistRoot string

	mesh owners.RuleMesh

	rule   *PolarRulesPanel.DragRule
	active bool

	selfRule   *PolarRulesPanel.DragRule
	selfActive bool

	edits  chan Edit
	ruleIn chan PolarRulesPanel.Msg

	edgeActive       map[string]bool
	toggleSelfToPeer map[string]chan struct{}
	toggleIn         chan EdgeToggle

	kindActive     bool
	toggleSelfKind chan struct{}
	kindIn         chan struct{}

	geomToRuleCenterIn chan spatial.Vec3

	out  chan State
	wake chan struct{}
}

func New(id string) *RuleNode {
	return &RuleNode{
		id:                 id,
		mesh:               owners.NewRuleMesh(),
		active:             true,
		selfActive:         true,
		edits:              make(chan Edit, 8),
		ruleIn:             make(chan PolarRulesPanel.Msg, 8),
		geomToRuleCenterIn: make(chan spatial.Vec3, 1),
		edgeActive:         map[string]bool{},
		toggleSelfToPeer:   map[string]chan struct{}{},
		toggleIn:           make(chan EdgeToggle, 8),
		kindActive:         true,
		toggleSelfKind:     make(chan struct{}, 4),
		kindIn:             make(chan struct{}, 4),
		out:                make(chan State, 1),
		wake:               make(chan struct{}, 1),
	}
}

func (r *RuleNode) SetPersistRoot(root string) { r.persistRoot = root }

func (r *RuleNode) SeedRule(rule *PolarRulesPanel.DragRule, active bool) {
	r.rule = rule
	r.active = active
	r.mesh.SetSelfRuleKey(PolarRulesPanel.KeyOf(rule))
}

func (r *RuleNode) SeedSelfRule(rule *PolarRulesPanel.DragRule, active bool) {
	r.selfRule = rule
	r.selfActive = active
}

func (r *RuleNode) SelfRule() *PolarRulesPanel.DragRule { return r.selfRule }

func (r *RuleNode) SelfActive() bool { return r.selfActive }

func (r *RuleNode) RuleBackChannel(peerID string) chan PolarRulesPanel.Msg {
	return r.mesh.RuleBackChannel(peerID)
}

func (r *RuleNode) LinkRuleDown(peerID string, down chan PolarRulesPanel.Msg) {
	r.mesh.LinkRuleDown(peerID, down)
}

func (r *RuleNode) BroadcastSelf() {
	r.mesh.SetSelfRuleKey(PolarRulesPanel.KeyOf(r.rule))
	r.mesh.BroadcastRule(r.id)
}

func (r *RuleNode) Edits() chan<- Edit { return r.edits }

func (r *RuleNode) CenterIn() chan<- spatial.Vec3 { return r.geomToRuleCenterIn }

func (r *RuleNode) Out() <-chan State { return r.out }

func (r *RuleNode) Wake() <-chan struct{} { return r.wake }

func (r *RuleNode) Run(ctx context.Context) {
	for peerID, back := range r.mesh.BackChannels() {
		go r.forward(ctx, peerID, back)
	}
	for target, toggle := range r.toggleSelfToPeer {
		go r.forwardToggle(ctx, target, toggle)
	}
	go r.forwardKindToggle(ctx)

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
		case <-r.kindIn:
			r.applyKindToggle()
			r.publish()
		case c := <-r.geomToRuleCenterIn:
			r.mesh.SetSelfCenter(c)
			r.mesh.BroadcastCenter(r.id)
		}
	}
}

func (r *RuleNode) forward(ctx context.Context, peerID string, back chan PolarRulesPanel.Msg) {
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

func (r *RuleNode) publish() {
	groupID, groupSize := r.mesh.RuleGroup(r.id)
	peerCenters := make(map[string]spatial.Vec3, len(r.mesh.PeerCenters()))
	for id, c := range r.mesh.PeerCenters() {
		peerCenters[id] = c
	}
	edgeActive := make(map[string]bool, len(r.edgeActive))
	for target, active := range r.edgeActive {
		edgeActive[target] = active
	}
	state := State{
		Rule: r.rule, Active: r.active,
		GroupID: groupID, GroupSize: groupSize,
		EdgeActive: edgeActive,
		KindActive: r.kindActive,
		SelfRule:   r.selfRule, SelfActive: r.selfActive,
		PeerCenters: peerCenters,
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
