// consts.go — nodeactor's own inbox-depth/retry-bound constants, mirroring package
// Wiring's mover_registry.go moverInboxDepth/maxPendingSends (movedispatch-decomposition
// §20), the same duplication precedent edgemover.InboxDepth already set (§17): both sides
// need the same declared depth, and nodeactor cannot import package Wiring's constant
// without an import cycle (Wiring imports nodeactor, not the reverse).
package nodeactor

// inboxDepth is this actor's own per-channel inbox capacity (extIn, and each directed
// neighborIn) — a chosen depth for a burst of drag-time move messages, not a derived
// value. See package Wiring's moverInboxDepth for the full reasoning; both sides carry the
// same 8.
const inboxDepth = 8

// maxPendingSends is the declared, asserted upper bound on this node's own outbound retry
// queue (EnqueueSend/flushPending) — see package Wiring's maxPendingSends for the full
// reasoning. Same value, same formula.
const maxPendingSends = inboxDepth * inboxDepth
