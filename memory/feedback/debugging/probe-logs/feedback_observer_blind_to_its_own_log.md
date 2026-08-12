---
name: observer-blind-to-its-own-log
description: A diagnostic that reads a log written BY the process it measures is blind to exactly the fastest-dying instances, so it silently merges two fast events into one slow one. Read the supervising process's log instead.
type: feedback
---

**Rule:** when a tool measures a process by reading that process's own log, it cannot see the instances that died before flushing — and those are precisely the fast ones the measurement is about. The observer is blind to one end of the distribution it exists to measure.

**Where this was paid for (2026-07-28, reload-gap):** `tools/reload-gap.sh` paired VS Code's `exthost.log` "exiting" line with the next "started" line. A host that lived one second produced NEITHER line — absent from the file entirely, not merely partial — so the pairing spanned the invisible host and reported **3.5s for what were two ~1.7s reloads**. One slow reload was reported where two fast ones happened.

Fixed by taking exits from `main.log`, written by the SUPERVISING process, which logs an `exited with code` line for every pid, and pairing each `started` with the latest exit before it. A gap over 60s is a window that sat closed, not a respawn, and is dropped.

**Practice:** before trusting a log-derived measurement, ask which records that log is ALLOWED to omit, and prefer the log written by the supervisor over the log written by the subject.

**Second lesson from the same investigation:** the regression itself was real (3.9/4.9/4.4s against a 1.8s baseline) and was host memory/process pressure — 44 days uptime, 2.83 GB of 4 GB swap. A full machine restart returned it to ~1.6s with swap at 0. A VS Code quit+relaunch was measured and does NOT fix it. Nothing in this repo was at fault, and two earlier theories (`.probe` log size, file watchers) had already been disproven by landing fixes that did not move the number.

See [[debug-data-before-theory]].
