package Dispatch

import (
	"context"
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/CardPanel"
)

func applyCardHit(md *MoveDispatch, f CardPanel.Field) {
	CardPanel.StartDraft(&md.UI.Card.Edit, f)
	md.UI.EmitViewFrame(nil)
}

func applyCardKey(ctx context.Context, md *MoveDispatch, key string) {
	c, ok := CardPanel.Key(&md.UI.Card.Edit, key)
	if ok {
		msg := CardPanel.EditMsg{Field: c.Field, Value: c.Value}
		if c.Field.Vector == CardPanel.VecS {
			md.UI.Card.S = c.Value
			if md.UI.PersistCardS != nil {
				md.UI.PersistCardS(int32(c.Value))
			}
			md.CardInboxes.Broadcast(ctx, msg)
		} else {
			card := md.UI.Card.Cards[c.Field.Node]
			card.Set(c.Field, c.Value)
			md.UI.Card.Cards[c.Field.Node] = card
			md.CardInboxes.Send(ctx, strconv.Itoa(c.Field.Node), msg)
		}
	}
	md.UI.EmitViewFrame(nil)
}
