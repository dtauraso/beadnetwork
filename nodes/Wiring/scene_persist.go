package Wiring

// scene_persist.go — shared machinery for the domain persisters in this package
// (overlays, camera viewpoint, scene sphere, node position, local polars, port anchor).
// Each of those files repeated the same thing: an atomic (tmp+rename) write. This file
// factors that machinery out once so the per-domain files hold only their domain-specific
// shape (which fields they own, how to marshal them).
//
// There used to be a shared debounce/coalesce timer here too (debouncedPersister[T]) — each
// domain persister's schedule() armed a 250ms timer and a background goroutine wrote the
// LATEST value once the burst settled. It was removed: writeJSONAtomic does no fsync (writes
// land in the page cache; the kernel already coalesces them), and the sibling writer
// WriteLocalPolars already wrote synchronously on the same drag path with no reported
// problem, so the debounce was solving a problem nobody had measured. Each persister's
// schedule() now writes synchronously, inline on whatever goroutine called it — see the
// commit that removed debouncedPersister for the reasoning in full.
//
// One-file-per-writer: every lock
// that used to live here (sceneFileMu, entityFileMus/entityFileMuMu) existed because TWO
// WRITERS shared ONE file — three writers read-modify-wrote view/scene.json, and two
// writers read-modify-wrote nodes/<id>/meta.json. Splitting each shared file into one file
// per writer (view/camera.json, view/overlays.json, view/sphere.json,
// nodes/<id>/position.json, nodes/<id>/local-polars.json) removes the sharing, so there is
// nothing left to serialize — those locks are DELETED, not narrowed. Each new file's writer
// marshals its own current state fresh and writes the whole file; there is no
// read-modify-write of a document another goroutine might also be writing.
//
// The one exception is nodes/<id>/{inputs,outputs}/<port>.json (port anchor files): those
// always had exactly one writer (writePortAnchor) — entityReadModifyWrite below is kept,
// delocked, purely to preserve the port file's OTHER field (`name`) across a write; it was
// never actually racing a second writer.
//
// Old-format compatibility, and it is now ASYMMETRIC — the two halves diverged:
//
//   - view/scene.json: GONE. Its fallback was removed (07028170, see scene_paths.go's
//     header), so loadSceneViewpoint / loadSceneOverlays / loadSceneSphere each read their
//     own file ONLY and default when it is absent. Nothing reads scene.json; a topology
//     that still has one on disk is carrying a dead file.
//   - nodes/<id>/meta.json: LIVE and required. It still owns static node identity, and
//     position.json / local-polars.json override its legacy position and localPolars/pole
//     fields when present, falling back to what meta.json carried when they are absent
//     (loader_tree.go).
//
// Neither is migrated or rewritten by this package.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// safeTreePathComponent reports whether s is safe to use as a SINGLE path segment
// under the topology tree root — i.e. it is a plain name, not "", ".", "..", not
// absolute, and contains no path separator. Guards node ids / port names (spec- or
// input-controlled) before they are filepath.Join'd into a write path, so a value like
// "../../x" cannot escape the tree root (path traversal).
func safeTreePathComponent(s string) bool {
	return s != "" && s != "." && s != ".." && !filepath.IsAbs(s) &&
		!strings.ContainsRune(s, '/') && !strings.ContainsRune(s, '\\') &&
		s == filepath.Base(s)
}

// atomicWriteTmpSuffix is the temp-file suffix writeJSONAtomic uses before renaming into
// place, so a reader never observes a partially-written file.
const atomicWriteTmpSuffix = ".tmp"

// logPersistErr logs a persister write failure in the uniform shape used by every
// debounced persister in this package. Fire-and-forget: persistence never blocks or panics
// on a failed write — the caller's `flush` just returns, and the NEXT schedule naturally
// carries the current (still-correct, in-memory) state into a future write attempt.
func logPersistErr(label, path string, err error) {
	fmt.Fprintf(os.Stderr, "%s: write %s: %v\n", label, path, err)
}

// readJSONBestEffort reads path and unmarshals it into v; an absent or malformed file
// leaves v untouched (its zero value) rather than erroring. Used for LEGACY-format fallback
// reads (the old shared scene.json / meta.json a pre-split topology still has on disk) and
// anywhere else an absent file is the normal, expected case (e.g. before the first save).
func readJSONBestEffort(path string, v any) {
	readJSONIfExists(path, v)
}

// readJSONIfExists reads path and unmarshals it into v, reporting whether a file was
// actually present and successfully parsed. Unlike readJSONBestEffort's caller-facing
// contract (which only cares about v), this distinguishes "file absent" from "file present
// but its fields happen to be the zero value" — needed wherever a caller must not confuse a
// genuinely-zero-valued persisted record with no record at all (loader_tree.go's
// position.json/local-polars.json overlay check).
func readJSONIfExists(path string, v any) bool {
	raw, err := os.ReadFile(path)
	if err != nil || len(raw) == 0 {
		return false
	}
	return json.Unmarshal(raw, v) == nil
}

// writeJSONAtomic marshals v and writes it to path via a temp-file-then-rename, creating
// parent directories first, so a reader never observes a partially-written file.
func writeJSONAtomic(path string, v any) error {
	out, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	tmp := path + atomicWriteTmpSuffix
	if err := os.WriteFile(tmp, out, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
