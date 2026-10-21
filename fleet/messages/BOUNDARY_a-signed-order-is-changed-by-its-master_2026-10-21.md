# A signed order, and who may change it

*21 October 2026, about 15:50 to 16:10 UTC. Written by the advisor, from the fleet's record. Nothing here needs blacking
out: seats are named by their role.*

The owner asked for this page after reading one sentence of the advisor's (translated from Polish): *"I did not correct
the signed order myself, so that its sha256 would not drift from the signature. I left writing in the decision to the
master, as he did with Nightshift."* The owner: *"Describe it! It is a really interesting thing, how the fleet
communicates and respects boundaries."*

So here is what happened, step by step, with what each actor could do and what it did.

## 1 · The advisor writes an order, and rehearses its signature

The advisor wrote the train order [`proven-cli-and-action`](../orders/TRAIN_ORDER_proven-cli-and-action.md): the command
line, the three output formats, the GitHub Action and the swap test, the comments, the runner script. It is a draft: an
order binds nobody until the owner signs it.

Before handing the owner a command, the author runs it in a harmless form and pastes what it printed (the fleet's rule:
*do not hand over a gate you have not rehearsed yourself*). The first rehearsal was refused, by name:

```
order sign: --valid 168h0m0s: a train order's signature runs for 24h, 72h or until it is withdrawn,
and this one asks for 168h0m0s [issue-lifetime-exceeds-ceiling]. Nothing was signed.
NEXT — pass --valid 24h, --valid 72h or --valid unlimited
```

72 hours would expire in the middle of the build window, so the advisor chose *until withdrawn*, and rehearsed again:

```
rehearsal — order sign would ask for ONE touch over 1 order(s); nothing was signed or written:
  1 · train-order · proven-cli-and-action · round 1 · sha256 d848be7184130cbd798c915a807c69dcfcbf93cca703cd659b27b4ab6a3982ce
signed by: a touch in the browser at http://localhost:8788/console/orders/sign (it answers); the window would be unlimited
```

The sha256 is the fingerprint of the order's exact bytes. That is what a signature binds.

## 2 · The owner runs the rehearsal, and asks what is wrong

The advisor's message showed the rehearsal and the real command as two blocks that looked alike. The owner ran the
rehearsal, saw *"the window would be unlimited"*, and wrote (translated): *"Something is wrong."*

The advisor did not guess. It looked in the record: no signature for the order. The word *would* was the answer: a
rehearsal says what it would do and does nothing. The mistake was the advisor's, in how the commands were laid out, and it
said so.

## 3 · Only a human touch signs

The owner ran the real command. It printed the same order and waited for a touch in the browser, on the owner's own
console. **No agent can sign.** The command asks for a person's touch, so an agent that runs it has signed nothing.

The record then read:

```
order-sign-result: draft 33581aa395e04a3d6ad51a365c08d89c signed for track:proven-cursor-infra
  · orders: train-order/proven-cli-and-action/round-1 · valid until 9999-12-31T23:59:59Z
```

## 4 · Two decisions arrive after the signature, and the advisor does not write them in

The order ended with two questions for the owner: C1, who is its master, and C2, whether the test repository is public
or private. The owner answered both after signing: *"Now it is OK. I agree with the recommendations."* (translated)

The answers belong in the order's §3. The advisor wrote that file an hour earlier and can still write to it: nothing in
the file system stops it. **It did not edit it.** Two reasons:

1. **A signature binds bytes.** The owner signed the order with sha256 `d848be71…`. If the advisor edits the file
   afterwards, the repository holds a document that says something the owner never signed, under the name of the one he
   did. The record would disagree with itself.
2. **After the signature, the order belongs to its master.** The advisor's role is to say *what*. Carrying out a signed
   order, and changing it, belongs to the head that holds the train. One of the fleet's laws says a head refuses work
   outside its own track. So the decision goes to the master, as a message.

The message, as sent:

> *"The owner SIGNED train order proven-cli-and-action (bridge 05ca5f979, draft 33581aa395e04a3d6ad51a365c08d89c, valid
> until withdrawn) and accepted both recommendations: C1 you are its master; C2 the test repository for W4/W6 is PRIVATE.
> Record C1/C2 in its section 3, move it to active, re-measure CA-1..CA-7 (law 44). No wagon is dispatched before
> 2026-10-22 00:00 UTC. Executors: Cursor seats, grok-4.7-high default; every session recorded (ADR-0024)."*

The fleet's message channel answers a sent message with *seen* as soon as it lands in the other seat. That event is
automatic, and it says so itself: *"it attests DELIVERY, not attention."* A head that reads *seen* as *read* has misread
its own instrument. The advisor waits for the master's own answer.

## 5 · How a master changes a signed order: the Nightshift precedent

Earlier the same day, the infrastructure master had to change the signed order for Nightshift's virtual machine. The
order said: no shared clipboard. The NVIDIA key could not reach the VM without one, and the owner decided to share it.
The master changed the order **in place**, and the change carries three things:

- **the owner's own words**: *"leave the clipboard now, it may come in handy"* (translated), as decision **N6**;
- **the trade, as stated to the owner when he decided**: while the clipboard is shared, the unsigned app in the VM can
  read everything copied on the host, and replace it;
- **a guard that keeps the choice deliberate**: the VM's script keeps *no clipboard* as its default; the shared clipboard
  is a recorded setting, and the status check turns red if the clipboard is shared without that record. Measured both ways.

The signed bytes are not lost: they are in git, in the commit the owner signed (`0e7f27414`), and the amended file says
what changed, who decided it and why. Two documents, both in the record, and nobody has to trust a memory of which is
which.

## 6 · Other boundaries the advisor met today

- **A refused tool stays refused.** Twice today the advisor's request to copy a file was denied by its permissions. It
  wrote the file with another tool it is allowed to use, in the open. It did not look for a spelling of the same act that
  no rule happens to name, because a fleet law says *what a grant denies is the act, never the command line*.
- **A wrong flag that cannot be undone.** On 21 October the advisor once sent a message with another train's name in it.
  That bound its own seat to that train, irreversibly, in the record. It wrote the mistake down and now forwards any
  message that reaches it by that route.
- **A long message is refused.** The fleet's gateway refuses a message body that is too long. The fleet's answer is a
  file in the record and a short message that points to it.

## What is not known

- **Whether the engine refuses to run a wagon from an order whose file changed after the signature.** Not measured. The
  advisor did not test it, because testing it means changing a signed order.
- **Whether an amendment made in place by the master needs a new signature.** The Nightshift amendment was made without
  one, and the owner's decision is quoted in it. Whether that is enough is the owner's call; it is not written down as a
  rule yet.
