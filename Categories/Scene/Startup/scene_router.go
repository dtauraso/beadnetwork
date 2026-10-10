package Startup

import (
	"context"
	"fmt"
	"os"

	SceneB "github.com/dtauraso/beadnetwork/Categories/Scene"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Drag"
)

type OpenedScene struct {
	Inbox chan GestureInboxMsg
	Wait  func()
}

type SceneRouter struct {
	open chan int
}

const sceneOpenDepth = 4

func NewSceneRouter() *SceneRouter {
	return &SceneRouter{open: make(chan int, sceneOpenDepth)}
}

func (r *SceneRouter) Open(ctx context.Context, idx int) {
	select {
	case r.open <- idx:
	case <-ctx.Done():
	}
}

func (r *SceneRouter) Run(ctx context.Context, cancel context.CancelFunc, first int, open func(idx int) (OpenedScene, error)) error {
	opened := map[int]OpenedScene{}
	var viewport *GestureInboxMsg
	selected := first

	show := func(idx int) error {
		s, ok := opened[idx]
		if !ok {
			var err error
			if s, err = open(idx); err != nil {
				return err
			}
			opened[idx] = s
			if viewport != nil {
				SendGestureMsgBlocking(ctx, s.Inbox, *viewport)
			}
		}
		selected = idx
		return nil
	}
	if err := show(first); err != nil {
		return err
	}

	recs := make(chan GestureInboxMsg, gestureInboxDepth)
	stdinWG := startStdinReader(ctx, cancel, recs)

	for {
		select {
		case <-ctx.Done():
			for _, s := range opened {
				s.Wait()
			}
			stdinWG.Wait()
			return nil
		case idx := <-r.open:
			if err := show(idx); err != nil {
				fmt.Fprintf(os.Stderr, "scene tab: could not open scene %d: %v — staying on the current scene\n", idx, err)
			}
		case gm := <-recs:
			if isViewportEdit(gm) {
				viewport = &gm
				for _, s := range opened {
					SendGestureMsgBlocking(ctx, s.Inbox, gm)
				}
				continue
			}
			SendGestureMsgBlocking(ctx, opened[selected].Inbox, gm)
		}
	}
}

func isViewportEdit(gm GestureInboxMsg) bool {
	return gm.Kind == GestureMsgEdit &&
		int(gm.Entity) < len(Drag.UpdateKinds) && Drag.UpdateKinds[gm.Entity] == "scene" &&
		SceneB.IsViewport(gm.Attr)
}
