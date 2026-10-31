# Open questions

Every question [summary_ideas.md](summary_ideas.md) leaves open, with the place it comes
from and an example of how an answer is expressed.

The examples are about other software on purpose (the database and the security system
from [summary_writing.md](summary_writing.md)), so they show the shape of an answer
without being mistaken for the answer. Write the real answer under **Answer**, then
replace the *Not answered yet* where the question comes from.

---

## 1. What is it for, beyond rotating links?

From: [Vision → Purpose](summary_ideas.md#purpose), [Future directions](summary_ideas.md#future-directions).

**Shape:** It lets *someone* do *something*, which *the usual way* cannot, because *reason*.
Say what it lets them do, not how it does it.

**Example:** "A migration tool lets a team change a live database's schema without taking
it offline, which hand-written migrations cannot, because they lock the tables they change."

**Answer:** It lets me make visual simulations by directing an AI, without reading the code
it writes. The usual way to trust code is to read it; I check by watching instead, so the
screen has to be the simulation itself, and every behaviour has to be testable by pressing
buttons and watching.

## 2. Who is it for?

From: [Vision → Purpose](summary_ideas.md#purpose).

**Shape:** A person or group named by what they are trying to do, not by job title alone.

**Example:** "Teams running a database that cannot go offline, who change its schema every
week."

**Answer:** A visual thinker for whom reading code and prose is the bottleneck.

## 3. Why does node selection get a replacement?

From: [Roadmap](summary_ideas.md#roadmap).

**Shape:** The current one does *X*, which goes against *a part of the vision*; the
replacement serves that part. Name the part of the vision, not a dislike of the code.

**Example:** "The current secret store keeps keys in a config file, which breaks the rule
that a secret never touches disk; the replacement keeps them only in memory."

**Answer:** *Not answered yet.*

## 4. What does the node selection replacement have to do?

From: [Roadmap](summary_ideas.md#roadmap).

**Shape:** Each thing it must do, written as a property someone can check by running it:
what to do, and what they then see.

**Example:** "Done when: a key is rotated while the service runs, and no file on disk ever
held it."

**Answer:** *Not answered yet.*

## 5. Why does the camera get a replacement?

From: [Roadmap](summary_ideas.md#roadmap).

**Shape:** Same as question 3: what the current one does, which part of the vision it goes
against, which part the replacement serves.

**Example:** "The current schema viewer redraws from its own cached copy, so it can show a
column that is already gone; the replacement reads the live schema only."

**Answer:** *Not answered yet.*

## 6. What does the camera replacement have to do?

From: [Roadmap](summary_ideas.md#roadmap).

**Shape:** Same as question 4: checkable properties.

**Example:** "Done when: a column is dropped, and the viewer stops showing it without a
reload."

**Answer:** *Not answered yet.*

## 7. What should it be able to do that it cannot do now?

From: [Future directions](summary_ideas.md#future-directions).

**Shape:** It cannot yet *X*. It should, because *X* follows from *the bet* (or the
purpose). A capability, as an outcome, not a task list.

**Example:** "It cannot yet roll a schema change back. It should, because the purpose is
changing a live database safely, and a change that cannot be undone is not safe."

**Answer:** *Not answered yet.*

## 8. Which of those matters most?

From: [Future directions](summary_ideas.md#future-directions).

**Shape:** *X* matters most, because without it *the purpose* cannot be shown at all;
the others only make it better.

**Example:** "Rollback matters most: without it no team would trust the tool on a live
database, so speed and a nicer viewer would never be used."

**Answer:** *Not answered yet.*
