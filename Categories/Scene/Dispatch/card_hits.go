package Dispatch

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/CardPanel"
)

func toggleCardPanel(md *MoveDispatch) {
	md.UI.Card.Open = !md.UI.Card.Open
	if !md.UI.Card.Open {
		md.UI.Card.Edit = CardPanel.Edit{}
	}
	md.UI.EmitViewFrame(nil)
}

func applyCardHit(md *MoveDispatch, f CardPanel.Field) {
	CardPanel.StartDraft(&md.UI.Card.Edit, f)
	md.UI.EmitViewFrame(nil)
}

func applyCardKey(ctx context.Context, md *MoveDispatch, key string) {
	msg, ok := CardPanel.Key(&md.UI.Card.Edit, key)
	sc := md.UI.Constants
	if ok && msg.Field.Vector == CardPanel.VecS && !CardPanel.SFits(msg.Value, sc.MaxIndexPhi, sc.MaxIndexTheta) {
		md.UI.RefuseStructuralEdit(fmt.Sprintf("s = %d: 12s = %d does not divide the scene turn (%d φ, %d θ index steps), so a tick would not be a whole number of steps",
			msg.Value, CardPanel.TurnTicks*msg.Value, sc.MaxIndexPhi, sc.MaxIndexTheta))
		ok = false
	}
	if ok {
		if msg.Field.Vector.SceneWide() {
			value, persist := &md.UI.Card.S, md.UI.PersistCardS
			if msg.Field.Vector == CardPanel.VecM {
				value, persist = &md.UI.Card.M, md.UI.PersistCardM
			}
			*value = msg.Value
			if persist != nil {
				persist(int32(msg.Value))
			}
			md.CardInboxes.Broadcast(ctx, msg)
		} else {
			card := md.UI.Card.Cards[msg.Field.Node]
			card.Set(msg.Field, msg.Value)
			md.UI.Card.Cards[msg.Field.Node] = card
			md.CardInboxes.Send(ctx, strconv.Itoa(msg.Field.Node), msg)
		}
	}
	md.UI.EmitViewFrame(nil)
}
