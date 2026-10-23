# Tests: real review disagreements, turned into cases you can check

proven has to prove what it says. So before we write proven, we write down what it will be tested on.

This directory holds a corpus of code-review cases taken from the work of our own fleet of AI coding agents. Every case
starts from a real moment: two reviewers from different model families looked at the same change, at the same time,
and did not agree. We went back to each one, found what each reviewer claimed, what their evidence actually showed, and
what was established later. From that we write a small, synthetic case that keeps the reason they disagreed.

**This corpus is prior art.** It is built before the hackathon's build window opens on 22 October, so that proven can
be shown working on it, and it is declared as pre-existing evaluation material: dated, not judged as hackathon work,
never re-dated. Proven's own code starts on 22 October and nothing here imports it or prescribes how it works.

## 1. Why this corpus exists

A reviewer that has to prove its findings needs situations where the truth can be checked: a real defect, an allegation
that turns out false, a test that checks the wrong thing, two readings that are both right under different assumptions,
and evidence that simply was not available. This corpus gives proven those situations, with the expected outcome
written down **before** proven exists. The scenarios and their acceptance conditions are part of the plan, not something
fitted to the product afterwards.

## 2. Where it comes from

From the fleet's own development record, which runs from **6 July 2026** (first entry) and **16 July 2026** (first
recorded agent run) to today. In it we took every pair of reviews of the same change by two different model families,
started at most 30 minutes apart, where both verdicts could be read.

- That gave **41 pairs**. On 16 October we counted 31 of them as disagreeing, and published that figure.
- Re-reading every pair by hand on 19 October showed the count was wrong: in three pairs one verdict had been misread
  (two were both rejects; one review never wrote a verdict). The correct figure is **28 disagreements among 40 pairs with
  two readable verdicts (70 %)**.
- The disagreements date from **1 to 28 September 2026**.

What is public: the provenance of all 31 candidate pairs in [`corpus/PROVENANCE.md`](corpus/PROVENANCE.md) (models,
verdicts, why the reviewers diverged, what was established), and every case written from them. What stays private:
the review texts, the code under review, the internal records of how the fleet works, and the names of the work items.

## 3. The work it takes to make it usable

A disagreement in a log is not a test case. Each one needed:
- finding both reviews, the commit they reviewed, and the later record of what happened;
- separating what each reviewer **claimed** from what its evidence actually **established**;
- finding how it was settled — a measurement, a later round, or a human decision — and never using the majority, the
  model, or the eventual merge as the truth;
- grouping repeated rounds of the same work, so one problem is counted once;
- writing a synthetic reconstruction that keeps the reason for the disagreement and drops everything private;
- and building it in Go or TypeScript, checking it on the base and on the change, having a reviewer of another model
  family check it, and removing anything internal before publication.

An expected outcome is a conclusion that needs evidence, not an answer we picked.

## 4. What is available now (2026-10-22)

| | count | source |
|---|---|---|
| candidate pairs accounted for | 31 of 31 | `corpus/PROVENANCE.md` |
| real disagreements among them | 28 | `corpus/PROVENANCE.md` |
| distinct incidents (pieces of work) | 19 | `corpus/manifest.json` |
| case candidates | 21 | `corpus/manifest.json` |
| case contracts written | 21 of 21, language-neutral | `corpus/cases/` |
| cases with code | 5 written on 21 October, before the window, on branches, not merged; 16 ordered on 22 October, in the window | `corpus/manifest.json` |
| cases independently checked | 0 | |
| pairs documented without a case | 5 (three environment problems, two repeat rounds) | `corpus/PROVENANCE.md` |
| pairs that are not disagreements | 3 | `corpus/PROVENANCE.md` |

The counts are produced from `corpus/manifest.json`:

```
jq '{cases: (.cases|length),
     contracts: [.cases[]|select(.status.contract|startswith("written"))]|length,
     with_code: [.cases[]|select(.status.fixture!="not built")]|length}' test/corpus/manifest.json
```

## 5. How to use it

- `corpus/manifest.json` lists every case: its language, the mechanism of the disagreement, its source pairs and four
  statuses (contract, case code, checks, independent check).
- `corpus/cases/PV-NNN/CASE.md` is the case: the story, the contract, the provenance, what was kept and what replaced,
  both reviewers' claims, the reproduction plan and the limits.
- `corpus/cases/PV-NNN/expected/` is the answer key. **Withhold it** from a reviewer you are evaluating: give it only the
  case's `base/`, `change/` and the change's description.
- Each case gets two versions of the code under review, a readable diff and a check with exact commands, written from
  its contract by a test author who did not write the contract. A command that has not been run says so.

The corpus does not prescribe how proven, or any other reviewer, consumes it.

## 6. What it contributes

- A public set of review scenarios from real work, each with evidence behind its expected outcome.
- Acceptance expectations written before the product exists.
- A way for anyone to rerun a case, check our result, and tell us where we are wrong.

The work so far, measured: 41 review pairs, 82 review records and 7.4 MB of review text re-read, 31 pairs analysed one
by one, and a published count corrected. It was done in interactive sessions, for which our platform records no cost,
so we give no money figure.

## 7. What it does not establish

- These are selected, synthetic reconstructions from one fleet. They are not a sample of all code review.
- It is not a blind benchmark: proven's development will see these cases.
- It is not evidence that proven works.
- Publishing the cases does not publish how our fleet works inside.

## 8. Chronology and the build-window declaration

| what | when |
|---|---|
| the original disagreements | 1–28 September 2026 |
| the fleet's record they come from | 6 July 2026 – today |
| this corpus's documents (provenance, list, contracts) | from 19 October 2026, before the build window |
| case code and checks | five cases on 21 October (prior art); the other 16 in the build window, from 22 October 2026 |
| proven's implementation and its evaluation on this corpus | in the build window, 22–27 October 2026 |

The case contracts are prior art (from 19 October). Five cases were coded on 21 October, before the window, and are
declared as prior art; every other case's code is written in the build window, from 22 October.
Everything in this directory written before 22 October 2026 is **prior art**:
pre-existing evaluation material, declared as such in the submission and not judged as hackathon work. It is
independent of proven's code and design. Each file's date is its commit date; nothing in this directory will be
re-dated.

Licence: Apache-2.0, as the rest of this repository. All case material is synthetic and written for this corpus.
