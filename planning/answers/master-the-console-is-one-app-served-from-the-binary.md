# Answer — master of the console track (`the-console-is-one-app-served-from-the-binary`)

Date: 2026-10-16. Written from this track's record: orders, MASTER_VERIFY files and review deliveries
under `██████████████████████████the-console-is-one-app-served-from-the-binary/`. Every id below was grepped.

## 1. Critique of "Proof-carrying code review", from our named reviews

**What it gets right.** The findings that changed our code were the ones a test could carry.
- Session-layer r5 was rejected for three defects: an invalid choice consumed the stored decision, a
  false settle was accepted, and a gateway restart reset the client epoch. Round 6 turned each one into a
  named test with an unplug control (`ORDER_SESSIONLAYER_r6_2026-10-15.md`). After that, nobody argued.
- pending-acts C1, an API-key-shaped header leaking (review delivery `ee7c59b8f6e0`), was measured,
  not argued. The fix round shipped with a red-then-green control.

That is exactly "a finding counts when it is proven".

**What it misses. Five things, each from the record.**

1. **The most valuable finding of the month had no code to fail on.** The journey-continuation-contract
   review was design-only: ██ ██████████████ ████████████ █████ ████ ██ █ ████████ ███████ █████ █ █████
   █████. Rounds 2–3 (`REVIEW_ORDER_JOURNEYCONTRACT_r2_2026-10-16.md`; r3 delivery `3eb317acfe58`)
   reduced it to a stated residual: ███ ██████ ████████ ███ ████████ ███ ██████ ██ █.
   - Under "proof = a test failing on head", this is a QUESTION. Yet it was the one finding that changed
     the architecture.
   - The tool needs a third evidence kind: **an attack written as an executable scenario against the
     design**, before code exists, which then becomes the acceptance test of the implementing wagon.
     That is what our J1 doc now does (tests T01–T26).
2. **The wrong answer came from the author, and the human caught it, not a reviewer.** I (the master)
   proposed a one-click "Link to <email>" as the mitigation. The owner saw that the victim sees her own
   address and clicks, so the binding to the attacker's installation still happens. Record: bridge
   `6b2aa7b57` (option A written), superseded by `d60beadf4` (option C with three conditions).
   - A reviewer that only reviews *diffs* would never see this. The fix existed only as a decision.
   - Proof-carrying should cover *proposed remedies*: "write the reviewer's own attack test against the
     remedy before it is chosen". That one test would have failed on A in seconds.
3. **History beat both reviewers.** console-v2-warm-singleflight was REJECTED by its named review. Its
   commits still reached develop inside console-v2-reports-sql, which was stacked on it, so the rejected
   defect was live (`ORDER_JOURNALRESUME_r1_2026-10-16.md`, line 16).
   - No reviewer of reports-sql was asked about warm-singleflight's commits.
   - A `git log base..head` attribution step ("these commits belong to another PR whose review said
     REJECT") is cheap and would have caught it. The proposal lists history as context. It should also be
     a **gate check**.
4. **False alarms were settled by re-running on both trees, never by a second model.**
   - The reports-sql p95 "failure" (1.60 s under load) re-ran at 590/456 ms at the tip against 479 ms at
     the base. That makes it host load, not the branch: `_EXEC_STATUS.md` of console-v2-reports-sql.
   - The "smuggled-v1" red failed 3/20 at the base and 5/20 at the tip, so it is a flake
     (session-layer master-verify, delivery `db85bbe86b74`).
   - A leg refuted my own claim that `isbot` reaches the dist; it is tree-shaken
     (`MASTER_VERIFY_the-web-console-builds-to-a-pinned-dist-and-ships-no-source-r1_round-1_2026-10-06.md`).
   - A reviewer refuted my "suite green" on the fourth of four legs, with 39 reds attributed one by one
     (`PROVISION_C8_review-r2-notes.md`, line 29).

   **What settles disagreements is base-vs-tip with repetition counts.** That should be built in:
   "FAILS on head" is worthless if it also fails 1-in-5 on base.
5. **The reviewer that never ran is the common case, and proof needs a working runner.**
   - 7 reviews queued for about 2 h on a 2-slot host.
   - Every Linux claude-code spawn died within 2 minutes with `saphan: unknown command "saphan"`, from
     a doubled argv in a guard. The fix is `035e066f5`.
   - █████ later refused every leg on a missing `/run/user/1000`.

   If the evidence runner is the product, **"could not run" must be a first-class verdict with its
   cause**, never silence. The day-2 risk in the proposal is real. It is bigger than "Go and Python
   only".

**Also missed by our process:**
- **Whole-suite runs found what scoped checks missed.** status-chip r2 had 3 red vitest specs
  (`ORDER_STATUSCHIP_r2_2026-10-16.md`) that the leg's own scoped run did not show.
- **Theory missed the real bug.** code-stays-visible's entropy redaction ate file paths
  (`Edit / [REDACTED].go`, `ORDER_CODEVISIBLE_r1_2026-10-14.md` line 29). Reviewers argued the theory;
  a run over one real transcript showed it.
- **What makes a review useful to me:** findings ordered by "what I must do", each with the command
  that reproduces it on base and head, and an explicit UNRUN list. Prose I have to triage is not useful.

## Proposal 1: Threat-to-test design review

- track: The Agent That Can Explain Why.
- the problem, in one sentence a judge understands: the costliest bugs are decided in a design doc or a chat about "how should login work", where no code exists for an AI reviewer to test, so they ship.
- who has this problem outside our fleet: every team writing RFCs/ADRs for auth, payments, data deletion; security reviewers who read designs and argue.
- the data sources the agent investigates: (1) the design doc / RFC, (2) the existing code it will change (call sites, current behaviour), (3) git history of those files, (4) the standards it cites (e.g. OAuth RFC 8252 sections).
- what "explained with evidence" means here: each threat becomes an executable scenario stub that fails against a reference model of the design, plus its later acceptance test. Mock:
  ```
  T01 · high · DEMONSTRATED-ON-MODEL
  Threat    attacker-created link + victim login binds victim account to attacker's install
  Scenario  scenarios/T01_link_substitution.py — FAILS on design v1 (bind at login), PASSES on v2 (bind at code exchange)
  Grounded  ███████████████████ █████ ██ █████ ████████ █████ ████████ █████
  Remedy check  "one-click Link to <email>" → T01 still FAILS (victim sees own email)
  ```
- the 3-minute demo, step by step: paste a short login-flow RFC → agent lists 4 threats with scenario files → run them against the design's state-machine model (v1: 2 fail) → user proposes a remedy in chat → agent runs the same scenarios against it and shows one still fails → v2 passes all; scenarios exported as the implementation's acceptance tests.
- what goes to GitHub as open source: CLI + a tiny state-machine description format for flows, the threat→scenario generator, the runner, the report.
- the bounded piece Saphan later takes, and why we would build it anyway: the scenario schema and runner as the "design review" profile of our named reviews; we just needed it on J1.
- sponsor technology we can use honestly: Nemotron as the second model family that attacks the first's design; Cursor as an executor.
- what the fleet must build in 144 h: 1) flow model format and a simulator; 2) threat extraction from docs and code reads; 3) scenario generator and runner; 4) remedy re-check loop; 5) report, demo RFC, video.
- the biggest risk, and how we would know by day 2 that it is failing: modelling a design is too free-form. By day 2 the J1-style login flow must be expressible and T01 must fail on v1 and pass on v2; otherwise we cut to "scenario stubs + citations".
- anything that touches internal mechanisms or patents: none, as long as the demo flow is a generic OAuth loopback, not our activation/lease protocol.

## Proposal 2: Screen-truth checker

- track: Solving Fragmented Intelligence.
- the problem, in one sentence a judge understands: dashboards contradict themselves ("No — fully busy" next to "1 of 2 in use"), and no test notices because each widget is correct alone.
- who has this problem outside our fleet: anyone shipping admin consoles, status pages, billing screens.
- the data sources the agent investigates: (1) a screenshot and the DOM of the page, (2) the API/SQL the page reads, (3) the component source.
- what "explained with evidence" means here: each contradiction is shown with both on-screen claims, the data row, and the line that composes each claim. Mock: `"fully busy" ← consolelive.go: any seat full ⇒ account full; "1 of 2" ← sum of seat slots; data: seat 11-1 free, 11-2 live ⇒ verdict wrong`.
- the 3-minute demo, step by step: point it at a demo admin app → it crawls 5 screens → reports 2 contradictions and 1 layout overflow (a status pill grown into an oval over the filters), each with evidence → fix one → re-run shows it gone.
- what goes to GitHub as open source: crawler (Playwright), claim extractor, data-binding tracer, report.
- the bounded piece Saphan later takes: claim-vs-data consistency as a gate in our UI walks; we would build it anyway after the Agents and status-pill defects.
- sponsor technology we can use honestly: Cursor-built; a Nemotron vision/text model is optional.
- what the fleet must build in 144 h: crawler; claim extraction; tracer for one stack (React + a REST API); report; demo app.
- the biggest risk, and how we would know by day 2 that it is failing: tracing claims to data generically is hard. By day 2 one planted contradiction must be found with its data row, else we cut to "screenshot + DOM contradictions only".
- anything that touches internal mechanisms or patents: none.

## The question I could not answer this month

"Why has nothing moved for two hours?" On 2026-10-16 seven reviews of this track, including the
session layer the owner was waiting for, sat in `dispatch place` with only a `workspace:` line printed.
The answer took most of an evening and lived in four places:
- per-host slot counts: ████████████ has 2 slots;
- run records full of 2-minute deaths on █████ and █████████████;
- a far guard that doubled argv (fixed by `035e066f5`, but my installed binary predated it);
- later, a missing `/run/user/1000` on █████.

The bug files exist (`MEDIUM_dispatch-place-is-silent-for-15-minutes-before-admission_2026-10-13.md`,
`dispatch-place-waits-silently-when-the-stream-is-pinned-to-one-host_2026-10-15.md`), but nothing
answered the one question. A placement waiting with no reason on screen is the explanation gap I hit
most.

---

## REDACTED

Parts of this file are blacked out (█). Nothing was deleted or reworded: every other word is the author's, as written on 2026-10-16, and each black bar has the length of what it covers.

| what was blacked out | times | what it was |
|---|---|---|
| machine names | 5 | the names of our machines |
| internal locations | 1 | where a document or log lives on our machines or in our private record |
| locations in our private code | 1 | file names and line numbers in Saphan Studio's private source code |
| security details | 2 | the exact input or condition that gets past one of our own guards, or the details of a residual risk in a live flow |

Stream and track names (slugs), commit hashes, test names and law numbers were left as written.
