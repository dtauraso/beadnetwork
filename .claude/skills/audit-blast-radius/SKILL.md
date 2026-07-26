---
name: audit-blast-radius
description: Run the read-only blast-radius architecture audit (shared mutable state, god-modules, lockstep clusters) as an Explore subagent and return a ranked findings table.
---

Launch a **read-only** blast-radius architecture audit of this repo (Go network + TS
webview) via the `Explore` subagent. Do NOT edit any files.

Goal: find "blast radius" hotspots — code that forces an AI to load large swaths of the
system to change one part safely. This feeds the token-reduction program (target ~5× less
AI usage per feature: shrink blast radius, fit AI's priors, lock each win with a guard).

Give the subagent this task (grep-first; scope out `node_modules,out,.git,handoff-archive,memory`):

1. **Shared mutable state** in the Go layer — `sync.Mutex`, `sync.RWMutex`, `atomic.`,
   package-level `var` written after init, shared maps written by multiple goroutines.
   Doctrine is "ownership replaces locking," so each is a candidate violation. For each:
   file:line, what it guards, and how many distinct call sites touch it.
2. **Wide coupling** — packages imported by many others; files importing many packages;
   god-modules (single files >400 lines doing multiple jobs). `wc -l` the top ~15 largest
   non-generated non-test source files with a one-line "what it does" each.
3. **Change-path centrality** — files edited in lockstep per CLAUDE.md's primitive-landing
   rule (schema/codec/layout). Identify the lockstep-edit clusters.

Rank each finding High/Med/Low blast radius. Concrete file:line. No fixes — findings only,
as raw material for a categorized table.

When it returns, present the findings as a table by category and rank. If asked to make it
durable, write to a branch-local doc with the current `branch:` frontmatter, or file work
items as `task/audit-*` branch descriptions.
