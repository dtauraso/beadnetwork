# InputNode

## View

| Field | Value |
|-------|-------|
| kindId | 3 |
| kind | input |
| bg | #e0e0e0 |
| border | #666 |
| text | #1a1a1a |
| accent | #3fb950 |
| minWidth | 90 |
| shape | rect |
| fill | #e0e0e0 |
| stroke | #666 |
| width | 80 |
| height | 60 |

## Non-channel fields

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| Init | []int | `data.init` | sequence of values to emit |
| Repeat | bool | `data.repeat` | if true, cycle through Init indefinitely |

## Ports

| Name | Direction | EdgeKind | Optional | Notes |
|------|-----------|----------|----------|-------|
| Primary | out | chain |  | required broadcast output: forwards Init values, and its edge length sets this node's own emission cadence |
| ToExcitatory | out | chain | yes | fans the emitted value out to a Pulse node (sample-and-hold); active when wired |
| ToFeedbackSource | out | chain | yes | fans the emitted value out to the node that computes the step read back on FeedbackIn (change-step feedback); active when wired |
| FeedbackIn | in | chain | yes | receives step (1=advance, 0=hold index) from Time; enables feedback-ring mode when wired |

## Firing rule

Emission is **end-first**: the last element of Init is emitted first, working toward the front. Init=[10,20,30] emits 30, then 20, then 10; the default init:[0,1] emits 1 then 0. This is the `popEnd` read/pop of a working/backup double buffer, not a front-to-back index walk.

Plain emit path (FeedbackIn not wired): end-pop the working array each iteration, Fire and send each popped value on Primary. Exit when all values sent (or refill and continue if Repeat).

Feedback-ring path (FeedbackIn wired): iterate indefinitely.
1. Fire.
2. Send the current end value of the working array on Primary.
3. Block on FeedbackIn for a step value `s` from Time.
4. If `s == 1`, pop the end (advance the bead); refill working from backup when it empties. Otherwise hold and resend the same value.
5. Loop (exit on ctx cancel or wire close).

The first send is the ring seed; there is no t=0 deadlock because the send precedes the first FeedbackIn read.

## Runtime status

- Loader-registered: yes
- TSX render: present

## Default data

```json
{ "init": [0, 1] }
```
