package TopVector

import (
	"context"
	"fmt"
	"os"

	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

type Owner struct {
	SelfIndex  func() polarindex.Index
	Constants  func() polarindex.SceneConstants
	WorldPosAt func(polarindex.Index) Vec3
}

type Runner struct {
	owner  Owner
	holder *Holder
	values *ValueWriter
	id     string
}

func NewRunner(id string, owner Owner, holder *Holder, values *ValueWriter) *Runner {
	if owner.SelfIndex == nil || owner.Constants == nil || owner.WorldPosAt == nil {
		panic("TopVector.NewRunner: node " + id + " was given an Owner missing SelfIndex, Constants or WorldPosAt; this goroutine derives the far end from the node's OWN index composed with the stored delta, so with any of them absent it would draw an arrow that is not this node's top vector")
	}
	return &Runner{owner: owner, holder: holder, values: values, id: id}
}

func (r *Runner) Run(ctx context.Context) {
	if r == nil {
		return
	}
	clk := clock.NewRealClock()

	for {
		if ctx.Err() != nil {
			return
		}

		r.writeValues()

		if err := clk.SleepPulse(ctx); err != nil {
			return
		}
	}
}

func (r *Runner) writeValues() {
	if r.values == nil {
		return
	}
	_, set := r.holder.Delta()

	selfIdx := r.owner.SelfIndex()
	target := r.holder.TargetIndex(selfIdx, r.owner.Constants())

	var f Frame

	if set {
		shaft, head, ok := ArrowMatrices(r.owner.WorldPosAt(selfIdx), r.owner.WorldPosAt(target))
		f.Drawn, f.Shaft, f.Head = ok, shaft, head
	}

	if err := r.values.Write(f); err != nil {
		fmt.Fprintf(os.Stderr, "top vector values write (node %s): %v\n", r.id, err)
	}
}
