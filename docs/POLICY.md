# The organisation policy

One file sets the floor for every repository's `proven.yaml`, and a known risk is accepted only by an exception with an
owner and an expiry date ([ADR-0036](adr/ADR-0036-an-organisation-policy-and-exceptions-that-expire.md)).

## The setting

| Environment variable | Default | What it changes |
|---|---|---|
| `PROVEN_POLICY` | none | a path to the policy file, on the machine that runs Proven |

- The file is read once, when Proven starts, **before** any repository's `proven.yaml`.
- A path that is missing, a directory, unreadable, or larger than `PROVEN_MAX_CONFIG_BYTES` is refused on load: proven
  exits 7 and the refusal names `PROVEN_POLICY` and the path.
- A path **inside the repository under review** is refused, after links are resolved, because the change under review
  could edit it. Keep the policy in another repository or on the runner. A policy read from a repository is not
  built yet; this version takes a local path only.
- Sample mode reads the policy too.

## The format

The same flat format as `proven.yaml`: one `key: value` per line, comments, quotes, lists as comma-separated scalars, a
duplicate key refused. A policy file holds two kinds of key.

1. **Policy keys**: the `proven.yaml` keys in the table below, and no others. A `proven.yaml` key that is not in the table
   is refused in the policy, because no order says which of its values is stricter.
2. **Exception keys**: `exception.<id>.<field>`, described under *Exceptions*.

A policy key the repository does not set keeps the policy's value. A key the policy does not set is the repository's own.

## Stricter than

A repository may make a policy setting **stricter or leave it equal**. A weaker value is refused on load, and proven
exits 7. The refusal names the key, the repository's value, the policy's value and the policy file:

```
configuration refused: doubled_classes="security": weaker than the organisation policy /etc/proven/policy.yaml,
which has doubled_classes="security, backend": it stops doubling backend; allowed: a value as strict as the policy's, or stricter
```

| Key | Stricter means | A weaker value, refused |
|---|---|---|
| `path_rules` | the repository's list keeps every rule of the policy at the same class or a higher one; an extra rule is allowed only at or above the policy's fallback class (see *The effective class*) | a rule of the policy left out, its pattern in a lower class, or an extra rule below the policy's fallback |
| `class_order` | the same order as the policy's. A policy that sets `class_order`, `path_rules` or `fallback_class` fixes the order, because the order ranks the classes those keys name | any other order |
| `fallback_class` | the same class or a higher one in the order than the policy's fallback, which is `backend` when the policy sets a class key but not this one | a lower class |
| `doubled_classes` | a superset of the policy's classes (`off` is the empty set) | a class of the policy no longer doubled |
| `author_commit_trailers` | `on` when the policy says `on` | `off` |
| `author_description` | `on` when the policy says `on` | `off` |
| `author_trailer_keys` | a superset of the policy's keys, compared without case | a key of the policy dropped |
| `raise_class` | `on` when the policy says `on` | `off` |

### The effective class

The three keys that decide the class of a path (`path_rules`, `class_order`, `fallback_class`) are not compared one by
one. When the policy sets any of them, a repository may not give **any path** a lower class than the policy gives it.
A path takes the highest class of the rules that match it, else the fallback, so:

- the repository keeps every rule of the policy (or of the defaults, when the policy sets only the fallback or the order)
  at the same class or a higher one;
- a rule of the repository that is not one of the policy's must sit at or above the policy's fallback class. Below it,
  the rule could give an unmatched path a lower class than the fallback gives it. A pattern cannot be shown to match
  only paths a policy rule already covers, so the refusal takes the stricter reading;
- the repository's own fallback must sit at or above the policy's.

Example. The policy says `fallback_class: security`. A repository says `path_rules: **=docs`. Each value passes when it
is read alone against its own key, but every path that no policy rule covers would be reviewed as `docs`, not
`security`. It is refused, and the refusal names the key, the rule and both classes:

```
configuration refused: path_rules="**=docs": weaker than the organisation policy /etc/proven/policy.yaml, which has
path_rules="(not set; the policy's rules still apply)": its rule **=docs gives paths docs where the policy's fallback
gives security, below it; allowed: a value as strict as the policy's, or stricter
```

A policy that sets none of the three keys says nothing about the class of a path, and a repository's rules are its own.

Not policy keys, because their values have no order: `timeout`, `default_reviewer`, `special_reviewer`, `families`,
`author_record`. The shell's own settings (`--severity`, `--error`, `--include`) are not `proven.yaml` keys and a policy
cannot set them in this version.

## Exceptions

An exception accepts one known finding until a date. It lives in the policy file only; an `exception.` key in a
repository's `proven.yaml` is refused. Each exception is five keys with one id of `a-z`, `0-9` and `-`:

| Key | Value |
|---|---|
| `exception.<id>.rule` | the finding's rule id, for example `known-exploited-dependency` |
| `exception.<id>.location` | a path, or a path and a line: `go.mod` or `go.mod:5`. A path alone covers every line of it; a line covers a finding whose lines include it |
| `exception.<id>.reason` | why the risk is accepted |
| `exception.<id>.owner` | who answers for it |
| `exception.<id>.expires` | the last day it holds, `YYYY-MM-DD` |

**An exception without an owner or an expiry is refused on load**, and so is one with any other field missing or empty,
an expiry that is not a `YYYY-MM-DD` date, or a key that is not one of the five. Proven exits 7 and names the key.

In a review:

- **Until it expires** the finding is *accepted*: it is printed with its reason, owner and expiry, it is never hidden, and
  it does not count. `--error` does not fail for it. The exception holds through the whole of its expiry day, by the
  UTC date of the review's clock.
- **After it expires** the finding counts again, as if there were no exception, and the review prints
  `exception <id> expired on <date>: finding <name> counts again (rule …, owner …, reason: …)`.

In this build the text output prints both. The JSON, SARIF and pull-request-comment writers are not in this build; the
result they will read (`Result.Accepted`, `Result.Expired`) carries the four fields, so each writer needs only to print
them.

## Example

`/etc/proven/policy.yaml`, kept outside the repositories:

```
# Every repository: dependency manifests are security, and security is reviewed twice.
path_rules: go.mod=security, go.sum=security, package-lock.json=security
doubled_classes: security
fallback_class: backend
author_commit_trailers: on

exception.leftpad.rule: known-exploited-dependency
exception.leftpad.location: go.mod:5
exception.leftpad.reason: not reachable; the upgrade is planned for the next release
exception.leftpad.owner: payments-platform
exception.leftpad.expires: 2026-12-31
```

Run with `PROVEN_POLICY=/etc/proven/policy.yaml proven ci`. A repository whose `proven.yaml` says `doubled_classes: off`
is refused; one that says `doubled_classes: security, docs` is applied.
