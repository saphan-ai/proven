# fleet/orders/

An order is what a master asks an executor to build. It is written before any code, and it is written so that a reviewer can check the result without asking the author anything.

**Rules every order follows**

- **The owner's words come first.** If the order exists because the owner asked for something, his words are quoted with their date.
- **The test plan is written together with the implementation plan, before any code.** The acceptance section is that test plan.
- **Acceptance is something a reviewer can run.** A command and the result it must give. "Works well" is not acceptance.
- **Scope is stated both ways.** What to build, and what the executor must not touch.
- **An order may be wrong.** If the executor finds that the order cannot be done as written, it stops and says why instead of guessing.
- **The executor never merges.** It works in its own copy of the repository. Merging is the owner's act.

**Template**

```markdown
# ORDER — <slug> · round <n>

- change: <slug>
- round: <n>
- base: main @ <commit>
- executor: <which Cursor agent>
- written: <date and time, UTC>
- task: <one sentence: what is built and why>

## The owner's words
<the owner's request, verbatim, with its date — or "none">

## What to build
1. <a concrete change, and the file or package it lands in>

## Acceptance
1. <a command a reviewer can run, and the result it must give>

## Not in scope
- <what the executor must not touch>

## When this order cannot proceed
Stop, write in the status report what blocks you and whose decision it is, and wait.

## What to return
A status report in `fleet/status/` with the acceptance answered item by item.
```
