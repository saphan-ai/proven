# CONVERSATION — the advisor, build day 4, 25 October 2026

This page continues [`CONVERSATION_2026-10-24_advisor.md`](CONVERSATION_2026-10-24_advisor.md).

**How to read it.** Marcin, the owner, talks to the **advisor**. The advisor never builds; it passes work to the
**masters**, agents that each run one train of work, and to the **scribe**, which keeps the record. They talk by
messages with ids; an answer is a message that points back. For every hop: what the owner said, what the advisor did,
each message sent (→) and each answer (←) with its time in UTC. Seat names and private hosts are blacked out; security
findings stay private until fixed.

## Hop by hop

### 1 · owner → advisor: NVIDIA's API is free and unlimited now; NVIDIA writes code too

> *"We now have an unlimited number of free NVIDIAs. We can scale the fleet, and maybe even run an executor on a new
> NVIDIA seat?"*

Until today NVIDIA's Nemotron only reviewed in the fleet. The fleet's rule on which agent may do which kind of work was
amended by the owner: NVIDIA executors (and QwenCloud seats) may also write code, as long as it is not security work and
not user-interface work.

**→ to the master who runs the agent seats:** seat one NVIDIA executor, prove it on one small wagon, then three more.
That master asked the Proven masters for a small wagon to try; nobody answered for a while, so **the advisor chose one
itself**: a worked example of a model-provider adapter, `examples/generator`, from the new Meterless-engines train.

**← the Meterless-engines master** (05:15Z): cut as its own order, placed on an NVIDIA seat.

> *"W53 waits on its NVIDIA seat. We must add more NVIDIA seats. They are free and the servers are there; nothing may
> wait for NVIDIA. Two? Four seats?"*

The advisor recommended four, on the machine with the most headroom. Seated. The first piece of Proven written by
NVIDIA's model, the generator example, took 16 minutes and was merged (`3beaaab`).

> *"A signal that something waits for a seat should go at once to Slack and the console. It is the owner who must
> decide: more capacity, or stop something else. I just have to know."*

**→ two engine masters:** a waiting-for-a-seat notice, raised after two minutes, with both choices ready.

### 2 · owner → advisor: a run, then every review, until it is clean

> *"At the end of the hackathon, to show why Proven is needed and why the impact is huge: for each change, the run with
> its task and final commit and which model; then every review: which model, what it found; and so on until it is clean
> because everything was fixed. Classified: security, other, BLOCKING, OWED, NOTE, trivia."*

**→ the scribe:** a generator that rebuilds this from the record, rerunnable at every push, and a public page with no
costs and no private names.

**← the scribe** (06:50Z), first draft ahead of time: 112 changes; 240 reviews; 48 changes came back at least once;
116 BLOCKING conditions caught before merge; 9 times a second reviewer from another model family caught what the first
had passed. One question for the owner, about how to count conditions found in security reviews; the advisor
recommended showing two numbers side by side rather than one inflated one.

### 3 · owner → advisor: QuakeWatch, and Proven reviewing it

> *"We have QuakeWatch Android to do. We can start writing it slowly. We will send pull requests to Proven."*
> *"QuakeWatch server is already done, in sources and pushed to GitHub."*

The advisor pointed out that Proven could not yet read a Kotlin project, so Android pull requests would show little.
Three orders followed:

- **→ master core:** Kotlin and Gradle through one language adapter, before the freeze. **Merged** (`32cf6a6`).
- **→ the test-corpus master:** run Proven over every merged change of the public QuakeWatch server, and set what
  Proven finds beside what the fleet's reviewers found at the time.
- **The advisor wrote the kickoff** of a new master for the Android app, starting with its train order for the owner
  to sign.

### 4 · owner → advisor: the bug harvester

> *"Look at the Bugs feature in the console. That is where I should be able to set it: who fixes what, from their own
> track or something else, which class, when, how long, well parameterised with defaults. The engine is treated the same
> as Proven and other apps written in the studio."*

The advisor found the console's mock-up and an earlier design (agents report the bugs they notice), and designed the
missing half: fixing, with settings and defaults per project.

> *"We could also give a master an automatic delegation for a special kind of order without the owner's signature,
> which only measures. Zero commits. It prepares an exact repair order which, after the owner's acceptance, flies like a
> rocket."* · *"The engine enforces it even more strictly: the leg gets a special grant, not executor but
> bug-measurement, with git commit and even writing in the code forbidden. At its first mistake it wakes up."*

Both were written into the design and signed into the fleet's rules the same morning. **→ two engine masters** (the
engine side, the console page) and the scribe (the rules).

### 5 · owner → advisor: a pull request that fixes itself

> *"What if Proven had a feedback loop from the pull request's result to the coding agent on NVIDIA? Once it knows
> exactly what is wrong, it will manage. Pull request plus automated fix. If it does not manage, too bad. After it
> fixes, only the deterministic evidence is run again. If it is still wrong after that one try, the pull request is
> rejected. We have everything already."* · *"Agreed. That will be magic."*

**Written:** [ADR-0050](../../docs/adr/ADR-0050-one-automated-fix-attempt-proved-by-the-same-evidence.md): one fix
attempt for all proved findings, judged only by the same tests, queries and tools, then "fixed automatically, waiting for
a person" or "rejected"; Proven never commits to the author's branch.

**→ master core:** build it before the freeze.

**Two masters agree on a shape, through the advisor.** Master core builds the patch and the verdict; the command-line
master renders the pull-request comment. **← master core** (11:21Z), addressed to the advisor for the command-line
master: *"W59 puts, on the review result, one patch for the whole change and a verdict with each finding open or closed.
The renderer is yours. Tell me the field shape you want before the leg starts."* **→ the advisor relayed it** to the
command-line master, asking for the field shape (a patch per file or per hunk for suggested-change blocks, the verdict
values, the per-finding list), to fold into the readable review comment it is already building.

### 6 · along the way

- Two changes waited at the owner's merge while the shared record was busy; the owner retried, and both landed.
- The owner went for a massage after hours at the desk; the advisor watched the fleet every quarter of an hour and kept
  a log for his return. *"You have to look after yourself."*
