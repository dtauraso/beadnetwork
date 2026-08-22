package input

import (
	"context"

	clock "github.com/dtauraso/wirefold/Categories/Clock"
)

func (n *Node) updateFeedbackRing(ctx context.Context, working, backup *[]int, init []int, emitBeads func(), clk clock.Clock) {

	awaiting := false
	n.runStepLoop(ctx, clk, func() bool {
		if !awaiting {
			if !n.feedbackRingSend(working, backup, init, emitBeads, clk) {
				return false
			}
			awaiting = true
		}

		s, ok := n.FeedbackIn.PollRecv()
		if !ok {
			return true
		}

		awaiting = false
		n.feedbackRingReact(s, working, backup, init, emitBeads, clk)
		return true
	})
}

func (n *Node) feedbackRingSend(working, backup *[]int, init []int, emitBeads func(), clk clock.Clock) bool {

	if len(*working) == 0 {
		*working = *backup
		*backup = append([]int(nil), init...)
		emitBeads()
	}

	v := (*working)[len(*working)-1]
	if n.Fire != nil {
		n.Fire()
	}
	return n.broadcastPlace(v, clk.Tick())
}

func (n *Node) feedbackRingReact(s int, working, backup *[]int, init []int, emitBeads func(), clk clock.Clock) {
	if s != 1 {

		return
	}

	*working = (*working)[:len(*working)-1]
	if len(*working) == 0 {

		if n.EmitRefillSlide != nil {
			n.EmitRefillSlide(clk, n.SpeedCh, *backup)
		}
		*working = *backup
		*backup = append([]int(nil), init...)
	}
	emitBeads()
}
