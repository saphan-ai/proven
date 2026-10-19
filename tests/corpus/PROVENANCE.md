# Provenance — the 31 review disagreements behind this corpus

Prepared 2026-10-19. Public and redacted: the source records stay private, so each pair is named by a number, not by its
location. Model names are the ones recorded on each review run.

## Where the 31 come from

The fleet's own record, July–October 2026. We took every pair of reviews of the same change by two **different model**
**families**, started at most 30 minutes apart, where both verdicts could be read: 41 pairs, of which 31 were counted as
disagreeing on 2026-10-16. Re-read by hand on 2026-10-19:
- **3 of the 31 are not disagreements** (P05, P09, P41 below). The first count read a template line in one model's output
  as its verdict.
- So: **40 pairs with two readable verdicts, 28 of them disagree (70 %)** — not 31 of 41 (76 %), as published earlier.
- The 28 come from **19 incidents** (pieces of work) and give **21 case candidates** (PV-001 … PV-021). Later review
  rounds of the same piece of work count once per distinct finding, never once per round.

Verdicts: A = accept · C = accept with conditions · R = reject. Truth = what a later measurement or round established;
it is never the majority, the model, or whether the change merged.

| pair | month | original area | reviewer models and verdicts | incident | case | why they diverged | established | status |
|---|---|---|---|---|---|---|---|---|
| P01 | 2026-09 | Go | gpt-5.6-sol C · glm-5.2 A | I01 | PV-020 | policy or threshold | scope decision | case candidate |
| P02 | 2026-09 | Python tooling over HTML | gpt-5.6-sol R · glm-5.2 A | I02 | PV-001 | outside the order pointing | established defect | case candidate |
| P03 | 2026-09 | Go | gpt-5.6-sol R · claude-opus-5 C | I03 | PV-003 | policy or threshold | established defect | case candidate |
| P04 | 2026-09 | Go | gpt-5.6-sol R · glm-5.2 A | I04 | PV-005 | outside the order pointing | established defect | case candidate |
| P05 | 2026-09 | Go | gpt-5.6-sol R · glm-5.2 R | I02 | — | no verdict disagreement (pair list label artefact) | established defect | not a disagreement: both reviewers rejected (the 2026-10-16 count misread one verdict) |
| P06 | 2026-09 | Go | gpt-5.6-sol R · glm-5.2 A | I03 | PV-004 | contradiction on fact | established defect | case candidate |
| P07 | 2026-09 | Go | gpt-5.6-sol R · glm-5.2 A | I02 | PV-002 | test checks wrong mechanism | established defect | case candidate |
| P08 | 2026-09 | Go | gpt-5.6-sol R · glm-5.2 A | I02 | PV-002 | test checks wrong mechanism | established defect | case candidate |
| P09 | 2026-09 | Go | gpt-5.6-sol R · glm-5.2 R | I03 | — |  verdict mislabel in pairs 41 (both rejected)   disjoint grounds decided by host capability | established defect | not a disagreement: both reviewers rejected (the 2026-10-16 count misread one verdict) |
| P10 | 2026-09 | Go | claude-opus-5 C · gpt-5.6-sol R | I05 | PV-006 | policy or threshold | established defect | case candidate |
| P11 | 2026-09 | Go | claude-opus-5 C · gpt-5.6-sol R | I06 | PV-007 | test checks wrong mechanism | established defect | case candidate |
| P12 | 2026-09 | Go | gpt-5.6-sol R · claude-opus-5 C | I06 | — | policy or threshold | established defect | repeat round of I06; folded into PV-007 and PV-008 |
| P14 | 2026-09 | Go | claude-opus-5 C · gpt-5.6-sol R | I06 | PV-008 | policy or threshold | established defect | case candidate |
| P15 | 2026-09 | Go | claude-opus-5 C · glm-5.2 A | I06 | — | policy or threshold | established defect | repeat round of I06; minor; folded into PV-007 |
| P16 | 2026-09 | Go | claude-opus-5 C · glm-5.2 A | I07 | PV-009 | outside the order pointing | established defect | case candidate |
| P18 | 2026-09 | Go | glm-5.2 A · claude-opus-5 R | I07 | PV-010 | test checks wrong mechanism | established defect | case candidate |
| P19 | 2026-09 | Go | claude-opus-5 R · glm-5.2 A | I07 | — | environment artefact | established defect | environment: both reviewers were given the previous commit; documented, no fixture |
| P20 | 2026-09 | Go | claude-opus-5 C · gpt-5.6-sol R | I08 | PV-021 | policy or threshold | context-dependent | case candidate |
| P21 | 2026-09 | Go | claude-opus-5 C · gpt-5.6-sol R | I07 | PV-011 | policy or threshold | established defect | case candidate |
| P24 | 2026-09 | Go | claude-opus-5 R · glm-5.2 A | I09 | PV-012 | test checks wrong mechanism | established defect | case candidate |
| P25 | 2026-09 | Go | claude-opus-5 C · glm-5.2 A | I07 | PV-011 | contradiction on fact | established defect | case candidate |
| P26 | 2026-09 | Go | glm-5.2 C · gpt-5.6-sol R | I10 | PV-013 | test checks wrong mechanism | established defect | case candidate |
| P27 | 2026-09 | Go | glm-5.2 C · gpt-5.6-sol R | I11 | PV-014 | disjoint assigned subjects | established defect | case candidate |
| P28 | 2026-09 | Go | gpt-5.6-sol R · claude-opus-5 C | I12 | PV-015 | test checks wrong mechanism | established defect | case candidate |
| P29 | 2026-09 | documentation (MDX) | claude-opus-5 C · gpt-5.6-sol R | I13 | PV-016 | disjoint assigned subjects | established defect | case candidate |
| P32 | 2026-09 | Go | gpt-5.6-sol R · glm-5.2 A | I14 | — | environment artefact | context-dependent | environment: one reviewer could not open a local socket; context-dependent; documented, no fixture |
| P35 | 2026-09 | Go | gpt-5.6-sol R · claude-opus-5 C | I15 | PV-017 | environment artefact | established defect | case candidate |
| P37 | 2026-09 | Go | claude-opus-5 C · glm-5.2 R | I16 | PV-018 | test checks wrong mechanism | established defect | case candidate |
| P38 | 2026-09 | Go | gpt-5.6-sol R · glm-5.2 C | I17 | PV-019 | policy or threshold | established defect | case candidate |
| P39 | 2026-09 | Go | gpt-5.6-sol R · claude-opus-5 A | I18 | — | environment artefact | context-dependent | environment: one reviewer could not open a local socket; not reconstructable |
| P41 | 2026-09 | Go | glm-5.2 A · gpt-5.6-terra C | I19 | — | environment artefact | established defect | not a disagreement: one review ended with an access error and wrote no verdict |

## What stays private

The review texts, the code under review, the orchestration records and the names of the work streams. Every case in
`cases/` is a synthetic reconstruction: it keeps the reason two reviewers diverged and replaces everything else.
