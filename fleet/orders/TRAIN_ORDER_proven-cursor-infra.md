# TRAIN ORDER — proven-cursor-infra

> **Public copy.** The original is in our private record. This copy says the same thing, with internal names, paths and private-code locations blacked out (█, one block per character). The list is at the end. The rule, in the owner's words (2026-10-17): *"We save in both places all hackathon documents."*
>
> **State of this document:** a draft. The advisor wrote it; the master may change it; the owner has not signed it yet. When the original changes, this copy changes with it.
>
> - written by: the advisor (claude-code, model `claude-opus-5-5`), 2026-10-17
> - cost: **not measured.** The advisor works in an interactive session, not a dispatched run, and Saphan Studio records a cost only for runs. Not run: no cost figure exists for this document.

- master: a claude-code master, to be appointed by the owner (none appointed when this was written, 2026-10-17)
- status: PAUSED — drafted by the advisor; the master may change it, then hands it to the owner to sign
- track: proven-cursor-infra
- repo: proven (NOT in `saphan workspace repo list` on 2026-10-17 — W1 registers it), with ████████████ and the bridge registry
- base: main at the cut
- author: ADVISOR of the hackathon track "The Agent That Can Explain Why"
- amended: by the advisor, 2026-10-17, after proven's design changed (Go product, Zetaris, Meterless, NVIDIA). Added:
  A13–A16, the additions to W3 and W4, W8, W9, two rules in §3 and their DoD lines. The owner: *"Add points 1 to 6."* (§0)

## WITHDRAWN (advisor, 2026-10-17)
On 2026-10-17 I wrote that ssh to the node "answers `Host key verification failed`" and made a host-key check the owner's
first act. That was my error, not the node's: I dialled as my own user. The owner: *"To ██████ you ssh as ████. This is our
standard."* As `████@██████` it answers (A8, re-measured). I also wrote, from `████████████`, that the node's guests
ask for more memory than it has. That is false too: the node has 123 GiB (A12). What still stands: the master re-measures
the node before setting the size.

## §0 THE OWNER, VERBATIM

2026-10-17, to the advisor (written in English):

> *"Let's prepare the infrastructure:
> 1. Check the status of the Cursor backend train first. If it allows it, we admit four Cursor seats now: one local on
>    macm5, one far native on macmini, and two on a new Linux machine, runner-cursor: one native, one in a Docker container.
> 2. Create runner-cursor on the Proxmox cluster, node █████. Model it on ████████████, scaled down for these two seats.
> 3. Development will be in Go and Python. Install both on runner-cursor. We have no Docker image with both, so take our
>    Go image and add Python as a separate layer. Save the Dockerfile in the proven repository; we may use it later."*

> *"Write a train order for the master to make these changes. The master may change it, and then gives it to me to sign.
> Admitting seats and creating the machine are my acts: the master prepares them for me as .sh scripts."*

> *"The master also documents the work in the public proven repository, and also ONLY IN ENGLISH. Instruct the master to
> write no internal information, and where that is not possible, to black it out: █, one block per character, and a
> REDACTED list at the end (proven/AGENTS.md). macm5, macmini and runner-cursor stay visible, because they show where the
> Cursor executors work. Black out every other machine name, cluster node names, internal paths and the machines'
> location."*

2026-10-17, later, to the advisor (written in Polish; given in the advisor's translation):

> *"D5: no, we will write it in Go with an adapter. Later we will use it in Saphan; that is the point of this game."*

> *"See whether we should add anything to the infrastructure we are preparing."*

> *"Add points 1 to 6."*

The six points are A13–A16 with the W3 and W4 additions, W8, W9, and the two rules in §3.

**Why the train grew.** The hackathon requires every entry to integrate Zetaris and Meterless. proven's product is now Go.
Its evidence is queried through Zetaris. Two Meterless engines (H-MEM, World Model) are built in Go and checked by
Meterless's own TypeScript conformance suites. An NVIDIA Nemotron model refutes findings. `SAMPLE_MODE=true` runs the same
queries on a local engine over small samples. The public description is proven `README.md`, "What proven will do".

## §1 WHAT THIS TRAIN IS

The proven build window opens on 2026-10-22 00:00 UTC, and its executors are Cursor agents only. Today the fleet has two
Cursor seats, both on macm5, and no Linux host for Cursor. When this train is done: runner-cursor exists, carries Go and
Python, and holds the Cursor seats the engine can run today; every Cursor seat the engine can run has taken one turn; the
seats the engine cannot run yet are named, each with the engine change it waits for; the Go + Python Dockerfile is in
proven; and proven's public record says where the Cursor executors work, with everything internal blacked out. Added 2026-10-17:
both lanes also carry Node and a C toolchain; the legs reach the build's data sources, or the gap is named; and a daily
rehearsal runs proven the way a judge will, before every push.

**What the advisor measured on 2026-10-17. These are CLAIMS FOR YOU TO TEST, not ground (law #44). Re-measure each one
before you build on it. Where you find the advisor wrong, write it in the wagon's `master-was-wrong:` field.**

| # | claim (advisor, MEASURED 2026-10-17) | instrument |
|---|---|---|
| A1 | Landed on engine `develop` @ █████████: W1, W2, W3, W5, W7, W9, W16. Not landed: W4 (hooks), W4b, W6 (recorded and sealed), W8 (seat health and usage), W10, W11, W12, W14. | `git -C ████████████████████████████ merge-base --is-ancestor fix/<slug> develop` for each branch; `git log develop --grep=<slug>` |
| A2 | macm5 already has two Cursor executor seats: `████████████████` (slot 8) and `████████████████` (slot 9), billing `subscription`, account `unknown`. The `agent` CLI is in `████████████` on macm5. | `saphan machine list`; `ls ████████████` |
| A3 | macmini has no Cursor install and no Cursor seat. | `ssh macmini 'ls -d ██████████'` → no match |
| A4 | **A Cursor API key is refused over ssh:** `cursor-seat-credential-transport` — "this engine has no carrier that puts a key in the leg's environment alone". On macOS the HOME login fails (`security exit code 154`, no default keychain). ⇒ **a far macOS Cursor seat (macmini) has no working way to authenticate today.** | `git -C ████████████████████████████ grep -n 'no carrier that puts a key' develop -- ███████████████` → `███████████████████`; `████████████████████████████` |
| A5 | **A Cursor seat in a container on a far host is not built.** The container boundary is composed on the local transport only (`████████████████████████`); seated container runs refuse `container-seated-grant-unbuilt` (`██████████████████████`); Cursor is `██████████: false` (`██████████████████████████`); `██████████████` has no `█████████████████`. | `git grep` on `develop` as named; `git ls-tree --name-only develop ██████████████` |
| A6 | A far Linux native Cursor seat is buildable: HOME login, no key. **Whether `agent login` works under a seat HOME on Linux is UNMEASURED.** | `████████████████████████` §7.9; `███████████████████████████████` |
| A7 | Every Cursor leg today runs with no hook verdict (W4), reaches the vendor's endpoint directly with no recording proxy (W6), and reads health and usage as unknown (W8). These are recorded, not refused. A class whose write set narrows below the whole workdir is refused `cursor-class-narrowing-unenforced`. **Whether the executor class narrows, so that every Cursor executor leg is refused today, is UNMEASURED.** | `█████████████████████████████████████`; `████████████████████████`, `████` |
| A8 | The node is reached as **████** (our standard, the owner 2026-10-17): `████@██████` answers `██████`. As another user the same dial gives `Host key verification failed`. | `ssh -o BatchMode=yes ████@██████ hostname` |
| A12 | ██████: 123 GiB memory, 86 used, 36 available; seven VMs (███████) asking for 102,400 MiB in total; `rpool` 616 G free. The next free VMID is ███. | `ssh ████@██████ 'free -g; zfs list -o name,used,avail rpool; qm list'` |
| A9 | ████████████: 8 vCPU, 15 GiB, `/` 387 GB (43 % used), Ubuntu 24.04.4, Python 3.12.3, Go 1.26.5, no docker or podman. `████████████` says the disk is 200G. That row is stale. | `ssh ████████████ 'nproc; free -g; df -h /; python3 --version; go version'` |
| A10 | Our Go image `dev-go` (engine `████████████████████████`: `███████████` + Go 1.26 copied from the official image by digest) is **not published**. `hub.docker.com/v2/repositories/saphan/dev-go/` → 404. Debian trixie, the image's distro, ships Python 3.13; Ubuntu 24.04 ships 3.12. | `curl -s -o /dev/null -w '%{http_code}' …` → 404; engine `████████████████████████████` |
| A11 | `proven` and `████████████` are not workspace repos. | `saphan workspace repo list` |
| A13 | **Meterless's conformance suites need Node.** The H-MEM suite runs as `HMEM_IMPL=<absolute path to index.ts> npx tsx conformance/runner.ts`: it loads the implementation as a TypeScript module, so a Go engine is checked through a thin `index.ts` adapter. **The Node version the suites need is UNMEASURED. Whether a Go engine passes through such an adapter is UNMEASURED.** | github.com/Meterless/Meterless, `engines/hmem` README, read 2026-10-17 |
| A14 | **The local engine for `SAMPLE_MODE` needs cgo.** The plan is DuckDB through its Go driver, which builds with cgo and a C compiler. **Whether `dev-go` carries a C compiler, and its `CGO_ENABLED` default, are UNMEASURED.** | `docker run --rm <dev-go> sh -c 'gcc --version; go env CGO_ENABLED'` |
| A15 | **The build reaches these hosts:** `api.osv.dev` and the OSV data dumps, `www.cisa.gov` (KEV), the FIRST EPSS download host, `api.github.com`, `integrate.api.nvidia.com`, the Zetaris instance (host unknown until the instance arrives), `registry.npmjs.org`, `proxy.golang.org`. **Whether the landed `the-cursor-egress-baseline` admits them for a Cursor leg is UNMEASURED.** | the list: proven `README.md` and the hackathon's technical guide; the admission: W8 |
| A16 | **A live-key test cannot run on a far leg today.** A4 refuses a Cursor key over ssh: "no carrier that puts a key in the leg's environment alone". The same gap is expected for the keys proven's live tests need (NVIDIA, Zetaris, a model provider). **Whether the engine can deliver a non-vendor secret into a far leg's environment is UNMEASURED.** | A4's instrument; W1 answers it with one zero-cost compose |

## §2 WAGONS

Every wagon's DoD carries (i) the commands and their outputs, (ii) a `master-was-wrong:` field against the §1 rows it
builds on, (iii) for each owner-act script, the rehearsal output pasted beside the script (law #33).

### W1 — proven is a workspace repo, and the advisor's readings are re-measured
- **lands:** proven registered as a workspace repo, so a Cursor leg can be placed into it on 2026-10-22. If registering is
  an owner act, a `.sh` for the owner. A1–A11 re-measured, each marked holds / wrong / changed, with the command.
- **dod:** `saphan workspace repo list` names proven. One zero-cost compose for a Cursor executor on `████████████████`
  (`saphan dispatch command`, no run) answers A7. A refusal is a valid answer. Paste it.
- **review:** master-verify-only

### W2 — runner-cursor exists, modelled on ████████████
- **lands:** an owner-act script `create-runner-cursor.sh` that creates one VM on the node that holds ████████████: the
  ████████████ recipe (Ubuntu 24.04 cloud image, x86-64-v2-AES per ████████, q35, virtio-scsi-single, qemu-guest-agent,
  `onboot=1`, cloud-init user and key), the next free VMID and the next free IP of the workbench block, both derived and
  printed. Wiring like ████████████: DNS rewrite, Kuma monitor, the node's nightly PBS job confirmed to cover it, the Mac
  ssh alias, and a tailnet member like ████████████, with the node authentication as an owner step. A brief and a report in
  ████████████ (`███████████████████████████`, `███████████████████████████████████`), `████████████` updated, and A9's
  stale disk figure corrected.
- **size:** the advisor recommends **4 vCPU / 8 GiB / 120 GB**, half of ████████████ in CPU and memory, for two seats and
  image builds. A12 says 8 GiB fits in the 36 GiB available, and 120 GB fits in the 616 G free. Re-measure at the cut and
  print the arithmetic (law #46) before you set the size.
- **access:** the node is reached as `████@██████` (A8). Never accept a changed host key on anyone's behalf.
- **dod:** `ssh runner-cursor 'hostname; nproc; free -g; df -h /'` matches the size; DNS and Kuma green; the PBS job's
  scope includes the new VMID, read from the node's job config; `tailscale status` lists runner-cursor.
- **review:** master-verify-only

### W3 — runner-cursor carries Go, Python, Docker and the pinned Cursor CLI
- **lands:** Go 1.26.5, as on ████████████. **Python: ONE minor version in both lanes, native and container.** The advisor
  recommends 3.13, the image's Debian minor (A10). A reviewer who sees 3.12 in one lane and 3.13 in the other cannot tell
  a code fault from a lane fault. You choose how to install it, and record why. Docker Engine, because the owner asked for
  Docker. Do not put the leg user in the `docker` group without the owner's word: that group gives full control of the host, so it is a
  security decision. The Cursor `agent` CLI via `███████████████████████████████` and `████████████████████████`
  (W9 of saphan-agent-cursor), one seat HOME per seat. Any step that needs sudo goes into an owner-act script.
- **added 2026-10-17 (A13, A14):** Node.js, one LTS major in both lanes (native and the W4 image), with npm and `npx`,
  for Meterless's conformance suites. A C compiler and libc headers for cgo, in both lanes. Read A13's Node requirement
  from the suite's own `package.json` before you pin the major.
  **dod addition:** `node --version`, `npm --version`, `gcc --version` and `go env CGO_ENABLED` on runner-cursor, each
  pasted; the Node major is equal in both lanes.
- **dod:** `go version`, `python3 --version`, `docker version` and `agent --version` on runner-cursor, each pasted; the
  `agent` version equals the pin. Mutation: a pin file with a wrong version makes the install script refuse by name.
  Rehearse it with `--dry-run` and paste the output.
- **review:** master-verify-only

### W4 — the Go + Python image, and its Dockerfile in proven
- **lands:** one Dockerfile in proven at the path the master proposes (advisor's suggestion: `fleet/infra/go-python/Dockerfile`
  with a README). `ARG BASE` is our `dev-go` image pinned by digest. One `RUN` layer adds Python (the minor of W3, plus venv
  and pip, as the engine's `dev-python` layer does). No vendor CLI, because we do not redistribute one. The file and its README
  say that the base image is not published (A10), so a stranger cannot build it as it stands.
- **added 2026-10-17 (A13, A14):** the image carries W3's Node major and a C compiler: Node as one more pinned layer after
  Python. If `dev-go` already has gcc (A14), say so and add nothing.
  **dod addition:** `docker run --rm <image> sh -c 'node --version; gcc --version; go env CGO_ENABLED'` pasted, equal to W3.
- **dod:** built on runner-cursor; `docker run --rm <image> sh -c 'go version; python3 --version'` gives W3's versions;
  the digest is recorded. The file passes the redaction rules of W7: no internal registry host, path or machine name.
- **owner gate:** D2 below. The file waits for it.
- **review:** master-verify-only

### W5 — the Cursor seats the engine can run today are admitted
- **lands:** owner-act scripts, each rehearsed (law #33):
  - **macm5 (local):** reuse `████████████████`. No new seat. Its API key per `████████████████████████` §7.9: a dashboard
    key, `████████████████` (0600), `saphan config set ████████████████████████████████ <NAME>`. The script never prints
    the key.
  - **runner-cursor native:** `saphan machine admit` (ssh transport, slots for the seats), `machine set-binary
    --binary cursor=<absolute path>`, `machine set-artifact-root`, `machine probe-confinement`, `machine seat --backend
    cursor --role executor --billing-class subscription --config-dir <the seat HOME>`. The login is
    `env HOME=<seat home> NO_OPEN_BROWSER=1 <agent> login`. That is the first measurement of A6.
  - **macmini and the runner-cursor container seat: NOT admitted by this train** (A4, A5). You write the two engine
    requests (§3, R1). The seats are admitted when the engine can run them, or replaced by D1's fallback.
- **dod:** `saphan machine list` shows each admitted seat; a zero-cost compose for each seat does not refuse, or its
  refusal is pasted with the act that clears it.
- **review:** master-verify-only

### W6 — each admitted seat takes one turn
- **lands:** one smoke turn per admitted seat, in a scratch repository. Nothing lands. This is a priced act. It runs only
  behind the owner's explicit consent at the call (law #49), with a quote and a cap of twice the quote (law #16).
- **dod:** per seat: the turn's return; **the model that really ran** (Auto's choice included); and the share of the
  monthly pool before and after. The engine reads usage as unknown until W8 lands (A7), so the pool share comes from where
  the owner can read it. Write `UNMEASURED BY THE ENGINE` beside it. proven's AGENTS.md requires both figures for every run
  from 2026-10-22.
- **review:** master-verify-only

### W7 — the public record in proven
- **lands:** a public copy of each wagon's order and status, under proven's `fleet/` templates (proven `AGENTS.md` §1).
  Add one public page that says where the Cursor executors work and why. The advisor suggests `fleet/SEATS.md`; you
  choose the path. The README of W4's Dockerfile. All in English only.
- **redaction (the owner's rule, §0):** visible: **macm5, macmini, runner-cursor**, Cursor, Go, Python, Docker, slugs, law
  numbers, this repository's commit hashes. Black out with █, one block per character, and list each kind in a `REDACTED`
  section at the end of the document: every other machine name (████████████ and the rest), cluster and node names, IP
  addresses, internal domain and host names, user names, internal paths, private repository names (the Gitea repo and its
  owner), the machines' location, and everything in proven `AGENTS.md` §2. The advisor kept two product names visible:
  "Proxmox" and "Tailscale". The owner may overrule that, and if he does, black them out too.
- **dod:** a deny-list grep over the changed proven files for every name, address and path on the black-out list → zero
  hits. Paste the command and its output. The owner reviews and pushes. **The fleet never pushes** (proven `AGENTS.md` §3).
- **review:** master-verify-only; the owner reviews before his push

### W8 — the legs reach the build's data sources (added 2026-10-17)
- **lands:** A15 measured for a Cursor leg on each admitted seat: for every host, admitted / refused / unmeasured, with the
  reading. Where a host is refused, the change that admits it, written as the property that must hold: "a Cursor executor
  leg on proven reaches these hosts and no others". Who may change the egress policy is not the advisor's call: if it is
  an owner act, a rehearsed `.sh`; if it is engine work, a request to the master of `saphan-agent-cursor` (law #59). The
  Zetaris host is added the day the instance arrives.
- **dod:** per host, the reading pasted. Negative control (law #27): one host not on the list, refused, pasted. If a
  reading needs a live turn, it rides W6's smoke turn under the same consent; it never runs as a separate priced act.
- **routing:** egress policy is network work and security work, so this wagon runs on claude-code (law #9 (a), (b)).
- **review:** master-verify-only

### W9 — a daily rehearsal, the way a judge will run proven (added 2026-10-17)
- **lands:** one script in proven, at the path the master proposes (advisor's suggestion: `scripts/judge-rehearsal.sh`).
  From a fresh clone of proven's trunk, in a scratch directory: `docker build`; `docker run` with an EMPTY environment and
  `SAMPLE_MODE=true`; `POST /run` with each file in `input_examples/`. It checks that each answer is structured JSON, that
  the answers differ from each other, and that `logs/trace-<run_id>.jsonl` exists for each run and names at least two
  agents. It stops at the first failing step and names it. It runs on runner-cursor every day before the owner's 15:00 UTC
  push, and its result is published in proven (advisor's suggestion: `fleet/rehearsals/REHEARSAL_<date>.md`).
- **dod:** the script rehearsed against a minimal skeleton in a scratch directory before 2026-10-22, output pasted.
  Unplug test (law #41): with one input example removed, and again with the Dockerfile broken, the script goes red and
  names the step. First real run: 2026-10-22, before the push.
- **owner gate:** D4 (Docker on runner-cursor). Without it this wagon waits.
- **review:** master-verify-only

## §3 ORDER OF WORK

1. W1 first: it re-measures the ground every other wagon stands on.
2. W2 → W3 → (W4 ∥ W5) → W6. W7 goes with each wagon, in the same change as the work it describes.
3. **Target:** W1–W5 done by **2026-10-21 12:00 UTC**, before the onboarding call. The first Cursor leg of the build is
   2026-10-22.
4. Added 2026-10-17: the W3 and W4 additions travel with W3 and W4. W8 after W5. W9 after W4 and D4: its script by
   2026-10-21 12:00 UTC, its first real run on 2026-10-22 before the push.

**Two rules for every order of the build (added 2026-10-17):**
- **Live keys stay off far legs until A16 is answered.** Cursor legs on runner-cursor build and test in `SAMPLE_MODE=true`.
  Tests that need a live NVIDIA, Zetaris or model key run on macm5, or the master runs them. Every order says which.
- **Work on vulnerability data is routed and worded as security work.** It reads advisories and vulnerability catalogues,
  so a codex lane may refuse it as cybersecurity and return `ok` with nothing. Its reviews go to claude-code (law #9).
  Every order on it names the defect class, the asset, the proof (a failing test, or a query and its rows) and the
  boundary, and never asks for an attack (law #56). proven writes no exploit code; the hackathon rules forbid it too.

**R1 — two engine requests the master sends to the master of `saphan-agent-cursor`.** That is another track: you send a
message, you do no work on it (law #59). They are written as properties that must hold:
- *A far Cursor seat authenticates.* A Cursor API key reaches a leg on an ssh seat in that leg's environment only, never in
  argv, a file or a log. Without it, macmini cannot hold a Cursor seat (A4).
- *A Cursor leg runs inside a container on a far host.* A `█████████████████` vendor layer, the container sentence composed
  for the ssh transport, and a seated container grant (A5).
Ask for an estimate against 2026-10-21. Record the answer in this file.

**Decisions owed by the owner, in the order they block work:**
- **D1 — the four seats.** Recommended: admit now what runs today (macm5 `████████████████`, runner-cursor native). If R1's
  two changes have not landed by 2026-10-21 12:00 UTC, the fallback is a second native seat on runner-cursor and
  `████████████████` on macm5. That still gives four seats, all of them runnable. *Cost of the fallback:* the public
  record shows no far macOS seat and no container seat. *Cost of waiting:* the build may start with two seats. One more
  call is owed whichever way he goes: retire whichever macm5 seat is not used. Placement is derived (law #45), so a fifth
  admitted Cursor seat can be picked by the engine, and the team the owner declared is four Cursor agents.
- **D2 — the Dockerfile before 2026-10-22.** proven `AGENTS.md` §3: *"All project code is written between 22 and 27 October
  2026. Nothing written before 22 October may be copied in."* Recommended: land it now as fleet tooling under `fleet/infra/`,
  dated, with a line saying it is environment, not product code. Alternative: hold it in the bridge and copy it on
  2026-10-22.
- **D3 — Cursor accounts.** One seat is one vendor account. How many Cursor accounts or API keys does the owner hold for
  this, and which seat gets which? Only the owner knows.
- **D4 — Docker group on runner-cursor.** W3, above.

## §4 DEFINITION OF DONE

- runner-cursor exists, sized by printed arithmetic, wired like ████████████, with its brief and report in ████████████.
- Go 1.26.5 and one Python minor in both lanes, the same versions, pasted.
- The Dockerfile in proven (after D2), built once, with its digest recorded.
- Every Cursor seat the engine can run is admitted and has taken one turn, with its real model and the pool share
  (`UNMEASURED BY THE ENGINE` where it is).
- The seats not admitted are named in this file, each with the R1 change it waits for or D1's fallback.
- Every owner-act script was rehearsed by its author, and the output is pasted beside it.
- proven's public record lands in the same change as the work. The deny-list grep shows zero hits. The owner pushes.
- **Every hackathon document is saved in BOTH places** (the owner, 2026-10-17: *"We save in both places all hackathon
  documents."*). The original goes in the bridge. A redacted public copy goes in proven, this train order included, and it
  is updated whenever the original changes. The public copy of this file: proven `fleet/orders/TRAIN_ORDER_proven-cursor-infra.md`.
- Added 2026-10-17: Node and a C toolchain in both lanes, the versions pasted (A13, A14). Every A15 host admitted, or
  named with the change that admits it (W8). The judge rehearsal exists, has gone red when unplugged, and runs before
  every push (W9). Every build order states where its live-key tests run (§3).
- The closing line of law #38 in this file.

## §5 OUT OF SCOPE

- Engine changes: the ssh key carrier, the far container lane for Cursor, `█████████████████`, W4/W6/W8. These belong to
  the train `saphan-agent-cursor` (R1).
- proven's product code: not before 2026-10-22, and not in this train.
- proven's `.github/`, `LICENSE` and `AGENTS.md`: the owner's.
- Any push. Any change to ████████████ itself, apart from correcting its stale row in `████████████`.

---

## REDACTED

Parts of this file are blacked out (█). Nothing else was changed, except one sentence of the owner's, which he reworded himself.

| what was blacked out | times | what it was |
|---|---|---|
| machine names | 13 | the machine runner-cursor is modelled on |
| cluster node names | 9 | the node that will hold runner-cursor |
| account names | 7 | the account used to reach the node |
| infrastructure identifiers | 2 | the numbers of the node's virtual machines |
| private repository names | 4 | our infrastructure repository |
| internal paths | 11 | files in our infrastructure repository, a local checkout path, a local tool folder, a vendor folder on a machine |
| locations in our private code | 22 | file names and line numbers in Saphan Studio's code, and one field name |
| seat identities | 6 | the names of two Cursor seats on macm5 |
| credentials | 2 | the name of the credentials file and of the setting that names a seat's key |
| private commit | 1 | a commit of Saphan Studio's private engine |
| internal design references | 1 | the number of one of our infrastructure decision records |
| private image names | 1 | the internal name of our base image |
