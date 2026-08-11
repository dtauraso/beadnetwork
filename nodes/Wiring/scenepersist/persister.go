// persister.go — the ONE debounce-then-write actor shape shared by every scene-level file
// writer (overlays.json, sphere.json, speed.json, lattice.json — and mirrored by
// nodes/Wiring/camerapersist's ViewpointPersister for camera.json, which stays its own type
// since it also owns the FSM-viewpoint→PolarCamera conversion, not just a write call).
//
// Each of the four callers in nodes/Wiring's own persister files (scene_overlays_persist.go,
// scene_sphere_persist.go, scene_speed_persist.go, scene_lattice_persist.go) used to declare
// an almost-identical unexported type: one path-string field, and one method that checked
// nil/unarmed, called a package write func (e.g. WriteSceneOverlays, defined alongside this
// file), and logged any error under a fixed site tag. There is no debounce despite the
// historical "schedule" naming — each write is SYNCHRONOUS, inline on the caller's own (the
// view-owner) goroutine; the prior debounce/coalescing window was removed (see this
// package's other files' own headers) and never came back. Parameterized here by payload
// type T, with the write function bound as a func value — the codebase's own
// bound-func-value pattern (move_dispatch_construct.go's
// `ng.msg.sendMove = md.mr.enqueueFor(ng)`).
package scenepersist

import "github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"

// Persister writes a value of type T to its own file as it changes. Path == "" (unarmed —
// bare test construction, or before EnableEditPersist runs) makes Schedule a no-op. Write is
// bound once, at construction, to the specific WriteScene* func for this file (e.g.
// WriteSceneOverlays/WriteSceneSphere/WriteSceneSpeed/WriteSceneLattice); Tag is the site tag
// jsonpersist.LogPersistErr reports an error under, matching what each caller's file used to
// hardcode.
type Persister[T any] struct {
	Path  string
	Write func(path string, v T) error
	Tag   string
}

// Schedule writes v to p.Path synchronously via p.Write. Fire-and-forget: errors are logged
// (jsonpersist.LogPersistErr), not returned, so a write failure never blocks the caller.
// A nil *Persister or an unarmed (Path == "") one is a no-op — the same guard every one of
// the four original types carried individually.
func (p *Persister[T]) Schedule(v T) {
	if p == nil || p.Path == "" {
		return
	}
	if err := p.Write(p.Path, v); err != nil {
		jsonpersist.LogPersistErr(p.Tag, p.Path, err)
	}
}
