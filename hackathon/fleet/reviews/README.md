# fleet/reviews/

A review is written by an agent that did not write the change. It starts from a fresh context, with no memory of how the code was written, so it cannot lean on the author's reasoning. It does not only read the code: it breaks the code on purpose and checks that the tests fail.

**Rules every review follows**

- **Base against tip.** The tests run on the commit before the change and on the change itself. Every failure is classified: it was already there, or the change brought it.
- **Break it on purpose.** The reviewer removes a check the change added and confirms that a test turns red. A check that no test notices is not protected.
- **Add one case the executor did not test**, and see whether the code handles it.
- **The verdict is one of three:** `accept`, `accept-with-conditions` or `reject`.
- **Every condition has a weight:** `BLOCKING` (the change cannot land until it is fixed), `OWED` (it lands, and the fix is owed in a named later change) or `NOTE`.
- **Every review states its cost, to the cent, and counts its conditions** by weight in its header.
- **What was not run is listed.** A review with anything in "Not run" is never a plain `accept`.
- **A review is advice.** The reviewer can say no. Only the owner can say yes.

**Template**

```markdown
# REVIEW — <slug> · round <n>

- verdict: <accept | accept-with-conditions | reject>
- reviewer: <which agent and which model>
- cost: <$0.00, to the cent, for this review>
- conditions: <n> BLOCKING · <n> OWED · <n> NOTE
- reviewed: <base commit> → <tip commit>
- updated: <date and time, UTC>

## Summary
<at most five lines: what holds, what turned red, what is owed>

## Base against tip
<the tests run on both commits, with counts; every failure classified>

## Broken on purpose
<the check removed, the test that turned red, the result after restoring it>

## A case the executor did not test
<the case, and whether the code handles it>

## Conditions
| # | weight | where | what closes it |
|---|---|---|---|
| 1 | <BLOCKING / OWED / NOTE> | <file:line> | <the fix> |

## Not run
<everything not run, and why — or "none">
```
