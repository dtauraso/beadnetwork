# Writing the summary

How to write [summary_ideas.md](summary_ideas.md): what to keep, how to express a vision,
and what a summary cannot do.

## Keep what matters to the design on its own

- Keep things that are independently important to the design of the software.
  - If I am explaining that I designed a database, the schema matters. It is technical, but
    the names and the columns are important to how the application works.
  - In a security system, how the secrets are stored matters.
- Things that arose from the implementation matter less.
  - They could have gone one way or another; I just picked, and it does not affect the
    overall design.
  - When functionality was factored to achieve something, the class design is probably not
    important, and neither is the exact way the functions were written.

## Example: polar vector math

- Say there is a layer (a file or a class) that handles rotations in 3D, using Euler angles
  under the hood.
- I talk about what is around it: we have a 3D system and rotations are supported.
  Quaternions are on the roadmap to replace Euler angles.
- If it is on the roadmap, say how it fits into the vision of the software: Euler angles are
  not fine here because of reason X, and reason X should link back to the project.
- If it is novel, it gets paragraphs.

## How to express a vision

A vision says what the software is for and why it is built the way it is. It is the thing
every other section points back to. Write it in this order:

1. **The purpose.** One or two sentences: what the software is for, and who it is for.
   Say what it lets someone do, not how it does it.
2. **The bet.** What the design believes that the usual approach does not. A vision with no
   bet just describes what everyone already builds.
3. **What follows from the bet.** The few commitments the design keeps because of it. Each
   one should be something a reader can check in the running code.
4. **What it rules out.** The approaches it will not take, and the reason, tied back to the
   bet. ("Euler angles are not fine here because of reason X" is this step.)
5. **Where it is going.** The direction after the roadmap, as an outcome, not a task list.

Checks that it is a vision and not a feature list:

- Every roadmap item can say which part of the vision it serves. An item that cannot is
  either outside the vision or shows the vision is missing a part.
- It stays true if the implementation changes. If swapping a class or a file would make it
  false, it is describing the implementation, not the vision.
- It is short enough to say aloud. Detail belongs in the sections that point back to it.
- Questions that are still open are written down as open, not filled in with a guess.

## It will not be an accurate picture

A summary leaves things out, so it is not the full picture. The full picture is the code
and the running editor. Each property in the summary names the code to read and the
buttons to press, so a reader can go from the summary to the real thing.
