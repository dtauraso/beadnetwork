package rulenode

import "context"

type Link struct {
	node *RuleNode

	started bool

	groupID int32

	groupSize int32
}

func (l *Link) Attach(node *RuleNode) {
	l.node = node

	l.groupID, l.groupSize = node.mesh.RuleGroup(node.id)
}

func (l *Link) Node() *RuleNode { return l.node }

func (l *Link) Start(ctx context.Context) {
	if l.node == nil || l.started {
		return
	}
	l.started = true
	go l.node.Run(ctx)
}

func (l *Link) Wake() <-chan struct{} {
	if l.node == nil {
		return nil
	}
	return l.node.wake
}

func (l *Link) TakeState() (State, bool) {
	if l.node == nil {
		return State{}, false
	}
	select {
	case s := <-l.node.out:
		return s, true
	default:
		return State{}, false
	}
}

func (l *Link) SetGroup(groupID, groupSize int32) {
	l.groupID = groupID
	l.groupSize = groupSize
}

func (l *Link) Group() (groupID, size int32) {
	return l.groupID, l.groupSize
}
