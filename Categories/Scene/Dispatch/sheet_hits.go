package Dispatch

import (
	"context"
	"fmt"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/AngleDropdown"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/StepStore"
	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	"github.com/dtauraso/beadnetwork/Categories/Scene/View"
)

func applySheetHit(ctx context.Context, md *MoveDispatch, speedSinks SliderPanel.Sinks, h Sheet.Hit) {
	sh := &md.UI.Sheet
	if h.Part == Sheet.PartBody || h.Part == Sheet.PartUp || h.Part == Sheet.PartDown || h.Part == Sheet.PartRow {
		md.UI.Builder.PressMenu(h.Cell.Act)
		if h.Cell.Act.Op == Builder.OpMenu {
			md.UI.EmitViewFrame(nil)
			return
		}
	}
	if h.Part == Sheet.PartBody || h.Part == Sheet.PartUp || h.Part == Sheet.PartDown {
		md.UI.Builder.NotePick(h.Cell.ArrowAct())
	}
	switch h.Part {
	case Sheet.PartPill:
		sh.Toggle(h.Label)
	case Sheet.PartGrab:
		sh.StartResize(h.Label, h.DY)
	case Sheet.PartBody:
		if h.Cell.Act.Op == Sheet.OpDockFold {
			sh.ToggleFold(h.Cell.Act.Node)
		} else if h.Cell.Act.Op == SliderPanel.OpSpeed {
			setClockSpeed(md, speedSinks, SliderPanel.Settings[h.Cell.Act.I].Speed)
		} else if h.Cell.Act.Op == Sheet.OpDockTab {
			sh.SelectTab(h.Cell.Act.I)
			sh.StartTabDrag(h.Cell.Act.Node)
		} else if h.Cell.Act.Op == Builder.OpTreeSplit {
			md.UI.Builder.StartSplit(h.Cell.Act.I)
		} else if h.Cell.Act.Op == Builder.OpRename {
			md.UI.Builder.Rename(h.Cell.Act, &sh.Edit)
		} else if h.Cell.Typed {
			Sheet.StartDraft(&sh.Edit, h.Cell)
		} else {
			runSheetAct(ctx, md, h.Cell.Act, 0, nil)
		}
	case Sheet.PartUp, Sheet.PartDown:
		if !h.Cell.Dim {
			runSheetAct(ctx, md, h.Cell.ArrowAct(), map[Sheet.Part]int{Sheet.PartUp: 1, Sheet.PartDown: -1}[h.Part], nil)
		}
	}
	md.UI.EmitViewFrame(nil)
}

func applySheetKey(ctx context.Context, md *MoveDispatch, key string) bool {
	act, t, ok, err := Sheet.Key(&md.UI.Sheet.Edit, key)
	if err != nil {
		md.UI.Builder.Refused = err.Error()
	}
	if ok && act.Op == Builder.OpStepName {
		md.UI.Builder.Refused = md.UI.Builder.NameStep(act.I, t.Text)
	} else if ok && act.Op == Builder.OpTreeName {
		md.UI.Builder.Refused = md.UI.Builder.NamePath(StepStore.ParsePath(act.Node), t.Text)
	} else if ok {
		runSheetAct(ctx, md, act, 0, &t.V)
	}
	md.UI.EmitViewFrame(nil)
	return err != nil || (ok && md.UI.Builder.Refused != "")
}

func (md *MoveDispatch) stopPlayFor(ctx context.Context, a Sheet.Act) {
	if a.Op == Builder.OpAllPlay || a.Op == Builder.OpAllScale || a.Op == Builder.OpToastClose || a.Op == Builder.OpFolderPlay {
		return
	}
	md.stopPlay(ctx)
}

func runSheetAct(ctx context.Context, md *MoveDispatch, a Sheet.Act, sign int, typed *float64) {
	if a.Op == Builder.OpLinkGap {
		return
	}
	if a.Op == Builder.OpLinkPickOpen {
		a = md.UI.Builder.PickOpen(a)
	}
	md.UI.Builder.Refused = md.UI.Builder.PickLoop(a)
	if md.UI.Builder.Refused != "" {
		return
	}
	md.stopPlayFor(ctx, a)
	if a.Op != Builder.OpToastClose {
		md.UI.Builder.Pressed = a
	}
	if Builder.IsStepEdit(a) {
		if !Builder.IsStartEdit(a) {
			md.showFor(ctx, a.I)
		}
		md.sendAfterWalk(ctx, a.To, Builder.StepEditOf(a, sign, typed))
		return
	}
	if Builder.IsStepHead(a) {
		md.runStepHead(ctx, a, sign, typed)
		return
	}
	if Builder.IsStepDirOp(a) {
		refused, err := md.UI.Builder.StepDirOp(a)
		md.UI.Builder.Refused = refused
		if err != nil {
			View.LogPersistErr("sheet_step_dir", md.UI.Builder.Root, err)
		}
		md.ShowStep(ctx)
		return
	}
	if runSheetStructure(ctx, md, a, sign, typed) || runSheetTree(ctx, md, a) || runTreeStepOp(ctx, md, a) || runFramePop(md, a, typed) {
		return
	}
	switch a.Op {
	case AngleDropdown.OpLattice:
		setLatticeStep(md, sign, typed)
	case Builder.OpFolderPlay:
		md.runFolderPlay(ctx)
	case Builder.OpCardFold:
		md.UI.Builder.ToggleCard(a)
	case Builder.OpStepFold:
		md.UI.Builder.ToggleStepFold(a.I)
	case Builder.OpStepShow:
		was := md.UI.Builder.StepShown()
		md.UI.Builder.ShowStep(a, sign)
		if md.UI.Builder.StepShown() != was {
			md.ShowStep(ctx)
		}
	case Builder.OpToastClose:
		md.UI.Builder.Refused = ""
	case Builder.OpNetFold:
		if err := md.UI.Builder.SetNetworkOpen(true); err != nil {
			md.UI.Builder.Refused = fmt.Sprintf("could not save that the Network view is open: %v", err)
		}
		md.UI.Builder.NewOpen = false
		md.UI.Builder.FramePop, md.UI.Builder.FrameDraft = false, nil
	case Builder.OpStartFold:
		md.UI.Builder.StartOpen = !md.UI.Builder.StartOpen
	case Pills.OpOverlay, Pills.OpOverlayBoth:
		applyOverlayAct(md, a)
	}
}

func (md *MoveDispatch) Settle(ctx context.Context) { md.stepFrom(ctx, "", Links.KBoth, 0) }

func (md *MoveDispatch) stepFrom(ctx context.Context, from, k string, sign int) {
	m := &md.UI.Builder
	m.Step.Round++
	h := clock.HumanTick(m.TickMs)
	for _, id := range m.Nodes {
		md.sendLink(ctx, id, Links.StepAll{Round: m.Step.Round, H: h, K: k, Sign: sign, From: from, LockAll: m.StepOf(from).Locks(), FixedR: m.ManyIn()})
	}
}
