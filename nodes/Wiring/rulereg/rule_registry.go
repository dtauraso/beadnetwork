package rulereg

import (
	"context"
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/rulenode"
)

type RuleRegistry struct {
	byID map[string]*rulenode.RuleNode
}

func (rr *RuleRegistry) Claim(id string) *rulenode.RuleNode {
	if rr.byID == nil {
		rr.byID = map[string]*rulenode.RuleNode{}
	}
	if rn, ok := rr.byID[id]; ok {
		return rn
	}
	rn := rulenode.New(id)
	rr.byID[id] = rn
	return rn
}

func (rr *RuleRegistry) Get(id string) (*rulenode.RuleNode, bool) {
	rn, ok := rr.byID[id]
	return rn, ok
}

func (rr *RuleRegistry) MustGet(id string) *rulenode.RuleNode {
	rn, ok := rr.byID[id]
	if !ok {
		panic(fmt.Sprintf(
			"rulereg.MustGet: node %q has no rule goroutine, so its polar rule could never be edited, "+
				"broadcast or grouped — every loaded node claims one at build time",
			id))
	}
	return rn
}

func (rr *RuleRegistry) All() map[string]*rulenode.RuleNode { return rr.byID }

func (rr *RuleRegistry) Start(ctx context.Context) {
	for _, rn := range rr.byID {
		go rn.Run(ctx)
	}
}
