# Remove the subscriptions; redo what they coordinated in the CRUD system

## The target

No subscription, no external-store sync, and no derived cache in the webview. A reader reads
its column when it draws. What the subscriptions coordinated — a DOM panel finding out that
Go changed something — is redone in the CRUD system: the panel sends an edit, Go changes the
thing, Go streams the new thing, TS renders it.

## Why they are coordinators

A subscription is a writer telling a list of readers that something happened. It is the
shape the network model excludes, and it is here because TS derives: the panels build their
own rows from columns, so they need a copy to compare against and a signal to know when to
rebuild. Both the copy and the signal exist only because the deriving does.

## The inventory

**Subscribe mechanisms — 3 live, 4 already dead:**

| mechanism | used by |
|---|---|
| `subscribeFrame` (`src/webview/frame-tick.ts`) | 9 flag/row modules |
| `subscribeViewBlocks` (`three/scene/view-blocks.ts`) | `Tabs/tab-state.ts` |
| `subscribeViewFrame` (`snapshot-buffer.ts`) | `view-blocks.ts` |
| `subscribeEdgeStreamFrame`, `subscribeBeadStreamFrame`, `subscribeNodeStreamFrame`, `subscribeInteriorStreamFrame` | NOBODY — already dead, delete outright |

**12 hooks over 10 files, 8 consumers.** Each hook is `useSyncExternalStore(subscribe, readX)`
and every `readX` already exists as a plain function that reads columns.

**5 derived caches**, which exist ONLY to give `useSyncExternalStore` a stable reference:
`cachedRuleRows`, `cachedTiltVectorRows`, `cachedVals` (overlay), `cachedVals` (panel),
`cached` (tabs). Three of them bundle many flags into one object — `overlays.json` again, in
memory, after that file was split on disk.

**2 hooks have no callers at all**: `useDragNodeRow`, `useDraggedNodeName`.

## What breaks, and why that is the point

Removing the subscription removes the only thing that re-renders a DOM panel when Go changes
something the user did not just click. Until the redo lands, a panel shows what it read when
it last rendered. That is the visible cost of taking the coordinator out, and it is what the
CRUD redo restores — not by re-adding a signal, but by removing the reason a signal was
needed.

The canvas is unaffected: `useFrame` already runs every frame and reads columns directly.
`node-instances-update` is the model — it calls `overlayFlag()` inline and sets `count = 0`
rather than being told to re-render.

## The redo

A panel becomes a goroutine that owns its rows, their values and its screen rect, and streams
them; TS positions and paints. `docs/planning/panels-are-goroutines.md` carries that plan,
including the one decision to settle first: once Go owns the layout, TS knows only WHERE a
click landed, so panel input becomes raw-input plus hit-testing in Go rather than an
addressed edit per control.

## Order

1. Delete the 4 dead subscribe functions and the 2 uncalled hooks. Nothing observable changes.
2. Delete the 10 remaining hooks; every consumer calls the `readX` it already wraps.
3. Delete `frame-tick.ts`, `subscribeViewBlocks`, `subscribeViewFrame` once nothing calls them.
4. Delete the 5 caches. Without `useSyncExternalStore` nothing needs a stable reference, and
   each `readX` recomputes from the columns — no second copy.
5. Tighten `check-no-webview-state.sh`: its allowlist of files permitted to use
   `useSyncExternalStore` becomes empty, so the guard fails if any of this returns.

## Verification

`useSyncExternalStore` and `subscribe` appear nowhere in the webview, and the guard's
allowlist is empty — checked by the guard itself, with the allowlist emptied in the same
commit so it cannot pass vacuously.

Then drive the editor: the scene still draws (it never used these), and each panel shows
correct values on the render after an edit. A panel that goes stale until the next render is
EXPECTED at this point, not a regression to chase.

## Risks

- **This does not fix the hang.** The subscriber fan-out was a proposed fix for a proposed
  cause and it did not hold. Do not read a working editor after this as evidence either way.
- **Step 4 changes cost, not just shape.** The caches avoided rebuilding row arrays per read;
  without them `readNodeRuleRows` walks every node and edge on each call. It is called during
  render rather than per frame, but that is worth measuring rather than assuming, given the
  outstanding hang.
- **Panels go stale between renders** until the goroutine redo lands. If that is not
  acceptable in the interim, the order should be inverted: move the panels first and let the
  subscriptions fall out.
