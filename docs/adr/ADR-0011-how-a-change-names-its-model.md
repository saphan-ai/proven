# ADR-0011: A change names the model that wrote it in an `Assisted-by:` trailer by default; where proven reads it, and which keys it accepts, are configurable

**Status:** accepted · **Date:** 2026-10-19
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-19
**Extends:** [ADR-0008](ADR-0008-reviewer-model-chosen-against-the-authors-model.md) (the reviewer's model chosen against the author's)
**Trigger:** an open point of ADR-0008 and ADR-0009, the form in which a change names its model ([`planning/CONVERSATION_2026-10-19.md`](../../hackathon/planning/CONVERSATION_2026-10-19.md), hop 6)

## Context

[ADR-0008](ADR-0008-reviewer-model-chosen-against-the-authors-model.md) says proven reads the author's model from what the change states, and never guesses. It left open where and in what form a change states it. The owner asked for a proposal based on industry practice, made configurable:

> *"Propose something and make it configurable. Look at industry best practices."*

What the industry does today, as read on 19 October 2026:

- **The Linux kernel** asks contributors to record AI assistance in a commit trailer, `Assisted-by: LLM [TOOL1] [TOOL2]`, and forbids AI agents from adding `Signed-off-by:`, which only a person may give ([kernel documentation](https://docs.kernel.org/process/coding-assistants.html)). Tools built around it write the agent and the model as `AGENT_NAME:MODEL_VERSION`, for example `Assisted-by: Claude:claude-3-opus`, one trailer per agent and model.
- **Coding agents** often add `Co-authored-by:` naming the tool, by default.
- A Git trailer is a `Key: value` line at the end of a commit message. `git interpret-trailers` and `git log --format=%(trailers)` read it, so no new parser is needed.

Some teams, ours among them, put no AI attribution in commits or pull requests at all, and record the model elsewhere: our fleet writes it in each change's status report.

## Decision

1. **The default form is a declaration in settings**: the repository's owner names the author models in `proven.yaml`
   (`author_models: <agent>:<model>`, one entry per agent and model that writes changes there). The review is
   independent of each one's family. *(Amended 2026-10-27, the owner: an `Assisted-by:` line in a commit or a pull request
   is AI attribution, which this project does not write and does not ask anyone to write, so the author model is never
   parsed out of such text by default. Until then this point named the kernel's `Assisted-by: <agent>:<model>` trailer
   as the default form.)*
2. **Where proven looks is configurable**, in an order of precedence:
   - the `author_models` declaration (default: none declared);
   - commit trailers (default off; amended 2026-10-27, was on);
   - a line of the same form in the pull-request description (default off; amended 2026-10-27, was on);
   - a record file in the repository, for teams that keep the model out of commits and pull requests (default off; the path pattern is set by the team).
3. **Which trailer keys count is configurable.** By default `Assisted-by:`. A team may add `Co-authored-by:` lines that name a known coding agent, or its own key.
4. **The family of a model comes from a configurable table**, for example `claude-*` → Anthropic, `gpt-*` / `codex-*` → OpenAI, `nemotron-*` → NVIDIA, `qwen*` → Alibaba. A model the table does not know is reported as "family unknown", never guessed.
5. **A change written by several models is reviewed against all of them.** Each review states which of the author models it is independent of.
6. **Nothing stated means nothing assumed.** If no source names a model, the review says "author model not stated" and uses the default reviewer ([ADR-0008](ADR-0008-reviewer-model-chosen-against-the-authors-model.md)).
7. **Proven never writes a trailer into anyone's commits.** It only reads.

## Consequences

- Teams that already follow the kernel's convention, or whose coding agents add a trailer by default, get reviewer independence with no extra step.
- Teams that keep AI attribution out of their history, as our own fleet does, point proven at their record instead.
- A trailer is a statement, not proof. Proven reports what the change claims; a wrong claim gives a wrong independence label, and the review says where the claim came from.
- For proven's own build, the source is our fleet's record in `hackathon/fleet/status/`, where every executor run states its model.
