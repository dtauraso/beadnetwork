---
branch: task/step-tree
---

# 4 · How it looks in the Links tab

The Links tab shows the same tree the disk holds, one fold per level:

```
▾ sequence 8                         breadcrumb: sequence 8 › step 1 › 2→3
  ▾ step 1  "lift 3"        [+] [⧉] [×]
      head   φ ↑  θ ·  r ·
    ▸ node 2
    ▾ node 3
        ● 2→3   start θ 0   Δ 12   nΔ 1   1 0
  ▸ step 2                  [+] [⧉] [×]
  + add step
```

- `[+]` adds a step after this one, `[⧉]` duplicates it and `[×]` deletes it. These are the
  operations on page 3.
- A step's fold state is remembered per step, as `ToggleStepFold` does now.
- The breadcrumb names where the card being edited lives, so it is always clear which copy of
  2→3 you are changing: the network's, step 1's or step 2's.
- Nothing here is new drawing work. These are the existing pill and sheet pieces with one more
  fold level. The only new pieces are the three buttons and the breadcrumb.
