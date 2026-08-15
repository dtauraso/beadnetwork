package owners

type RuleCopy[S any] struct {
	fromRuleNode <-chan S

	ruleNodeWake <-chan struct{}

	groupID int32

	groupSize int32
}

func (r *RuleCopy[S]) LinkRuleNode(fromRuleNode <-chan S, wake <-chan struct{}) {
	r.fromRuleNode = fromRuleNode
	r.ruleNodeWake = wake
}

func (r *RuleCopy[S]) Wake() <-chan struct{} { return r.ruleNodeWake }

func (r *RuleCopy[S]) TakeState() (S, bool) {
	var zero S
	if r.fromRuleNode == nil {
		return zero, false
	}
	select {
	case s := <-r.fromRuleNode:
		return s, true
	default:
		return zero, false
	}
}

func (r *RuleCopy[S]) SetGroup(groupID, groupSize int32) {
	r.groupID = groupID
	r.groupSize = groupSize
}

func (r *RuleCopy[S]) Group() (groupID, size int32) {
	return r.groupID, r.groupSize
}
