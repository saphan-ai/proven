# ADR-0027: The code is written to be read by the judges: every package, every exported name and every non-obvious step carries a comment

**Status:** accepted · **Date:** 2026-10-23
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-23
**Follows:** [ADR-0013](ADR-0013-no-hardcoded-settings.md) (settings are named, so comments can name them)
**Trigger:** the owner, 23 October: the executors should write richer comments, because the judges will read this code; the
masters put it in their train orders, and the code written so far gets them too

## Context

The judges read Proven's source as well as its README. On 23 October the code written in the build window had few
comments: a reader could see what a function does, but rarely why it is there, which agent of the review it serves, or
which decision it carries out.

## Decision

1. **Every package has a package comment**: what it is for, and where it sits in the review (which of the six agents,
   or which adapter, uses it).
2. **Every exported name has a doc comment** that begins with the name, in the form `go doc` shows.
3. **Every step a reader would stop at gets a comment that says why**, never a comment that restates the code: a
   guard, an order that matters, a value that comes from a setting, a refusal.
4. **Where code carries out a decision, the comment names it** (for example `ADR-0025: the test runs with no network`).
5. **Every test says, in its comment, which property it proves.**
6. **Comments are in English and plain.** They carry no AI attribution, no internal host names, paths or credential
   names, and no commented-out code.
7. **Every train order of Proven carries this as a standing rule for its executors**, and each review checks it.
8. **The code already written gets the same comments** in one comment-only wagon per train: no behaviour changes, so
   the build, `go vet` and the tests give the same result before and after, and the diff changes comment lines only.

## Consequences

- A judge can follow a review through the code from the comments alone.
- The comment-only wagons cost a few hours of executor time during the build week and change no behaviour.
- A comment can go stale when the code under it changes; the review of each change reads both.
