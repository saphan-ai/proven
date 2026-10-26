package review

import (
	"github.com/saphan-ai/proven/internal/change"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The tests of ADR-0036: a policy sets the floor of every repository, and an exception accepts a known
// finding only with an owner and an expiry. The sample review of the discount fixture reports four
// findings; the tests below name two of them by rule and location.

// day is a clock fixed on one date, so an exception expires on the day a test chose.
func day(date string) func() time.Time {
	return func() time.Time {
		t, err := time.Parse(dateLayout, date)
		if err != nil {
			panic(err)
		}
		return t.Add(13 * time.Hour)
	}
}

// writeFile writes body in a new temporary directory, outside the module, and returns its path.
func writeFile(t *testing.T, name, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// reviewUnderPolicy reviews the discount fixture with the policy body and, when local is not empty, a
// proven.yaml of that body. now fixes the clock.
func reviewUnderPolicy(t *testing.T, policyBody, local, now string) (Result, error) {
	t.Helper()
	opts := Options{Sample: true, LogDir: t.TempDir(), PolicyPath: writeFile(t, "policy.yaml", policyBody), Now: day(now)}
	if local != "" {
		opts.ConfigPath = writeFile(t, "proven.yaml", local)
	}
	rev, err := Open(opts)
	if err != nil {
		return Result{}, err
	}
	return rev.Review(t.Context(), Request{Target: filepath.Join(moduleRoot(t), "testdata", "fixture", "discount")})
}

const discountException = `exception.discount.rule: discount-applied-twice
exception.discount.location: billing/discount.go:5
exception.discount.reason: the discount is reviewed by finance each quarter
exception.discount.owner: finance-platform
exception.discount.expires: 2026-12-31
`

// TestAWeakerLocalRuleIsRefusedByName: the policy doubles two classes and a repository doubles one.
// The refusal names the key, the local value, the policy's value and the policy file.
func TestAWeakerLocalRuleIsRefusedByName(t *testing.T) {
	_, err := reviewUnderPolicy(t, "doubled_classes: security, backend\n", "doubled_classes: security\n", "2026-10-24")
	ce, ok := err.(*ConfigError)
	if !ok {
		t.Fatalf("want a *ConfigError, got %v", err)
	}
	if ce.Field != "doubled_classes" || ce.Value != "security" {
		t.Fatalf("field %q value %q", ce.Field, ce.Value)
	}
	for _, want := range []string{"weaker than the organisation policy", "policy.yaml", `"security, backend"`, "stops doubling backend"} {
		if !strings.Contains(ce.Reason, want) {
			t.Errorf("reason lacks %q: %s", want, ce.Reason)
		}
	}
}

// TestEachPolicyKeyRefusesAWeakerValueAndAcceptsAnEqualOrStricterOne is one subtest per policy key: the
// stricter-than table of docs/POLICY.md, with a value that is weaker and one that is stricter or equal.
func TestEachPolicyKeyRefusesAWeakerValueAndAcceptsAnEqualOrStricterOne(t *testing.T) {
	cases := []struct {
		name, policy, weaker, why, ok string
	}{
		{"path_rules dropped", "path_rules: docs/**=security\n", "path_rules: go.mod=security\n", "leaves out the rule docs/**=security", "path_rules: docs/**=security, go.mod=security\n"},
		{"path_rules lowered", "path_rules: docs/**=security\n", "path_rules: docs/**=docs\n", "puts docs/** in docs, below security", "path_rules: docs/**=security\n"},
		{"fallback_class lowered", "fallback_class: backend\n", "fallback_class: docs\n", "falls back to docs, below backend", "fallback_class: security\n"},
		{"doubled_classes dropped", "doubled_classes: security\n", "doubled_classes: off\n", "stops doubling security", "doubled_classes: security, docs\n"},
		{"author_commit_trailers off", "author_commit_trailers: on\n", "author_commit_trailers: off\n", "turns the reading of commit trailers off", "author_commit_trailers: on\n"},
		{"author_description off", "author_description: on\n", "author_description: off\n", "turns the reading of the description off", "author_description: on\n"},
		{"raise_class off", "raise_class: on\n", "raise_class: off\n", "turns raise_class off", "raise_class: on\n"},
		{"author_trailer_keys dropped", "author_trailer_keys: Assisted-by\n", "author_trailer_keys: Co-authored-by\n", "drops the key Assisted-by", "author_trailer_keys: co-authored-by, ASSISTED-BY\n"},
		{"class_order changed", "class_order: security, backend, mobile, user interface, docs\n", "class_order: docs, security, backend, mobile, user interface\n", "fixes the order of the classes", "class_order: security, backend, mobile, user interface, docs\n"},
		{"class_order under path_rules", "path_rules: go.mod=security\n", "class_order: docs, security, backend, mobile, user interface\n", "fixes the order of the classes", "path_rules: go.mod=security\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p, err := loadPolicy(change.ProcessPolicy(), writeFile(t, "policy.yaml", c.policy))
			if err != nil {
				t.Fatal(err)
			}
			_, err = p.parseLocal([]byte(c.weaker))
			ce, isCfg := err.(*ConfigError)
			if !isCfg || !strings.Contains(ce.Reason, c.why) || !strings.Contains(ce.Reason, p.path) {
				t.Fatalf("weaker value: %v", err)
			}
			if _, err := p.parseLocal([]byte(c.ok)); err != nil {
				t.Fatalf("an equal or stricter value was refused: %v", err)
			}
		})
	}
}

// TestNoPathLoosensTheClassAcrossFields compares the class every path gets, not one key at a time. Each
// case is weaker only across fields: the repository's rule or fallback is fine against its own key and
// lowers a path below what the policy's other key gives it. The refusal names the key, the pattern, both
// classes and the policy file.
func TestNoPathLoosensTheClassAcrossFields(t *testing.T) {
	cases := []struct {
		name, policy, local string
		key                 string
		why                 []string
	}{
		{
			// The reviewer's case: every unmatched path now takes the docs rule.
			"a local wildcard rule under a policy fallback", "fallback_class: security\n", "path_rules: **=docs\n",
			"path_rules", []string{"**=docs", "fallback gives security", "policy's"},
		},
		{
			"a local fallback under a policy that only sets rules", "path_rules: docs/**=security\n", "fallback_class: docs\n",
			"fallback_class", []string{"falls back to docs", "falls back to docs, below backend"},
		},
		{
			"a local rule below the fallback beside the policy's own rules", "path_rules: go.mod=security\nfallback_class: backend\n", "path_rules: go.mod=security, **/*.md=docs\n",
			"path_rules", []string{"**/*.md=docs", "fallback gives backend", "policy's"},
		},
		{
			"a local rule below the default fallback when the policy sets only the order", "class_order: security, backend, mobile, user interface, docs\n", "path_rules: **/auth/**=security, **=docs\n",
			"path_rules", []string{"**=docs", "fallback gives backend", "policy's"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p, err := loadPolicy(change.ProcessPolicy(), writeFile(t, "policy.yaml", c.policy))
			if err != nil {
				t.Fatal(err)
			}
			_, err = p.parseLocal([]byte(c.local))
			ce, ok := err.(*ConfigError)
			if !ok || ce.Field != c.key {
				t.Fatalf("want a *ConfigError on %s, got %v", c.key, err)
			}
			for _, want := range append([]string{"weaker than the organisation policy", p.path}, c.why...) {
				if !strings.Contains(ce.Reason, want) {
					t.Errorf("reason lacks %q: %s", want, ce.Reason)
				}
			}
		})
	}
	// The same rule is free when the policy fixes no class and names no doubled class: it says nothing
	// about paths.
	p, err := loadPolicy(change.ProcessPolicy(), writeFile(t, "policy.yaml", "author_commit_trailers: on\n"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.parseLocal([]byte("path_rules: **=docs\n")); err != nil {
		t.Fatalf("a policy that fixes no class refused a local rule: %v", err)
	}
	// A policy that doubles a class says which paths it doubles, so the same rule is refused: it would
	// review go.mod, which the policy doubles as security, once as docs.
	p, err = loadPolicy(change.ProcessPolicy(), writeFile(t, "policy.yaml", "doubled_classes: security\n"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.parseLocal([]byte("path_rules: **=docs\n")); err == nil {
		t.Fatal("a policy that doubles security accepted a local rule that puts every path in docs")
	}
	// A rule at or above the policy's fallback, and the policy's own low rule, are accepted.
	p, err = loadPolicy(change.ProcessPolicy(), writeFile(t, "policy.yaml", "path_rules: docs/**=docs\nfallback_class: security\n"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.parseLocal([]byte("path_rules: docs/**=docs, **/*.go=security\n")); err != nil {
		t.Fatalf("an equal or stricter rule set was refused: %v", err)
	}
}

// TestALocalValueOfAKeyThePolicyDidNotSetIsFree:the policy sets one key; the repository changes another.
func TestALocalValueOfAKeyThePolicyDidNotSetIsFree(t *testing.T) {
	p, err := loadPolicy(change.ProcessPolicy(), writeFile(t, "policy.yaml", "author_commit_trailers: on\n"))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := p.parseLocal([]byte("fallback_class: docs\nraise_class: off\n"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Fallback != "docs" || !cfg.ReadCommits {
		t.Fatalf("fallback %q read commits %v", cfg.Fallback, cfg.ReadCommits)
	}
}

// TestARepositoryInheritsThePolicyValueItDoesNotSet: no local file, and a local file that is silent.
func TestARepositoryInheritsThePolicyValueItDoesNotSet(t *testing.T) {
	policyBody := "fallback_class: security\n"
	res, err := reviewUnderPolicy(t, policyBody, "", "2026-10-24")
	if err != nil {
		t.Fatal(err)
	}
	if res.PolicyStatement == "" || !strings.Contains(res.PolicyStatement, "policy.yaml") {
		t.Fatalf("statement %q", res.PolicyStatement)
	}
	// The inherited value is the one the review used: nothing in the policy's own rules matches the
	// fixture's paths, so its class is the policy's fallback and not the default backend.
	res, err = reviewUnderPolicy(t, "fallback_class: security\npath_rules: nothing-matches/**=docs\n", "timeout: 30s\n", "2026-10-24")
	if err != nil {
		t.Fatal(err)
	}
	if res.Class != "security" || res.ClassRule != "fallback" {
		t.Fatalf("class %q rule %q, want the policy's fallback security", res.Class, res.ClassRule)
	}
	p, err := loadPolicy(change.ProcessPolicy(), writeFile(t, "policy.yaml", policyBody))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := p.parseLocal([]byte("timeout: 30s\n"))
	if err != nil || cfg.Fallback != "security" || cfg.Timeout != 30*time.Second {
		t.Fatalf("cfg %+v err %v", cfg, err)
	}
}

// TestAnAcceptedFindingIsShownWithItsOwnerAndDoesNotCount: until the expiry date the finding is moved
// out of Findings, so it is not counted, and carries the four fields of its exception.
func TestAnAcceptedFindingIsShownWithItsOwnerAndDoesNotCount(t *testing.T) {
	res, err := reviewUnderPolicy(t, discountException, "", "2026-12-31")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range res.Findings {
		if f.ID == "F1" {
			t.Fatal("F1 is accepted and must not count")
		}
	}
	if len(res.Findings) != 3 || len(res.Accepted) != 1 || len(res.Expired) != 0 {
		t.Fatalf("findings %d accepted %d expired %d", len(res.Findings), len(res.Accepted), len(res.Expired))
	}
	got := res.Accepted[0]
	if got.Finding.ID != "F1" || got.Owner != "finance-platform" || got.Expires != "2026-12-31" ||
		got.Reason != "the discount is reviewed by finance each quarter" || got.ExceptionID != "discount" {
		t.Fatalf("accepted %+v", got)
	}
}

// TestAnExpiredExceptionNoLongerAcceptsAndTheReviewSaysSo: the day after the expiry the finding counts
// again and the exception is named with its date.
func TestAnExpiredExceptionNoLongerAcceptsAndTheReviewSaysSo(t *testing.T) {
	res, err := reviewUnderPolicy(t, discountException, "", "2027-01-01")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Findings) != 4 || len(res.Accepted) != 0 {
		t.Fatalf("findings %d accepted %d", len(res.Findings), len(res.Accepted))
	}
	if len(res.Expired) != 1 || res.Expired[0].Expires != "2026-12-31" || res.Expired[0].ExceptionID != "discount" || res.Expired[0].Finding != "F1" {
		t.Fatalf("expired %+v", res.Expired)
	}
}

// TestTheExpiryDayIsTheUTCDateInEveryZone: the exception of discountException holds through 2026-12-31 by
// the UTC date of the clock. Two zones, an hour either side of that date's end, as written in the zone
// and as the same instant in UTC: the answer follows the UTC date, never the zone's wall clock.
func TestTheExpiryDayIsTheUTCDateInEveryZone(t *testing.T) {
	bangkok := time.FixedZone("Asia/Bangkok", 7*3600)
	newYork := time.FixedZone("America/New_York", -5*3600)
	at := func(loc *time.Location, wall string) time.Time {
		ts, err := time.ParseInLocation("2006-01-02 15:04", wall, loc)
		if err != nil {
			t.Fatal(err)
		}
		return ts
	}
	cases := []struct {
		name     string
		now      time.Time
		accepted bool
	}{
		// Bangkok (UTC+7): the wall clock is already on 1 Jan, yet the UTC date is still 31 Dec until 07:00.
		{"Bangkok 06:59 on 1 Jan, UTC 23:59 on 31 Dec", at(bangkok, "2027-01-01 06:59"), true},
		{"Bangkok 07:00 on 1 Jan, UTC 00:00 on 1 Jan", at(bangkok, "2027-01-01 07:00"), false},
		// New York (UTC-5): the wall clock is still on 31 Dec when the UTC date has turned, at 19:00.
		{"New York 18:59 on 31 Dec, UTC 23:59 on 31 Dec", at(newYork, "2026-12-31 18:59"), true},
		{"New York 19:00 on 31 Dec, UTC 00:00 on 1 Jan", at(newYork, "2026-12-31 19:00"), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p, err := loadPolicy(change.ProcessPolicy(), writeFile(t, "policy.yaml", discountException))
			if err != nil {
				t.Fatal(err)
			}
			out := Result{Findings: []Finding{{ID: "F1", RuleID: "discount-applied-twice", Location: Location{Path: "billing/discount.go", StartLine: 5}}}}
			p.applyExceptions(&out, c.now)
			if got := len(out.Accepted) == 1; got != c.accepted || len(out.Findings) == 1 != !c.accepted {
				t.Fatalf("now %s (UTC %s): accepted %d findings %d, want accepted %v", c.now, c.now.UTC().Format(dateLayout), len(out.Accepted), len(out.Findings), c.accepted)
			}
			// The same instant in UTC gives the same answer.
			again := Result{Findings: []Finding{{ID: "F1", RuleID: "discount-applied-twice", Location: Location{Path: "billing/discount.go", StartLine: 5}}}}
			p.applyExceptions(&again, c.now.UTC())
			if len(again.Accepted) != len(out.Accepted) {
				t.Fatalf("zone %s and UTC disagree", c.now.Location())
			}
		})
	}
}

// TestAnExceptionCoversOnlyItsRuleAndItsLocation:the same path and line with another rule, and the
// rule with another path or another line, are not accepted.
func TestAnExceptionCoversOnlyItsRuleAndItsLocation(t *testing.T) {
	for name, edit := range map[string][2]string{
		"another rule": {"rule: discount-applied-twice", "rule: another-rule"},
		"another path": {"billing/discount.go:5", "billing/other.go:5"},
		"another line": {"billing/discount.go:5", "billing/discount.go:6"},
	} {
		t.Run(name, func(t *testing.T) {
			res, err := reviewUnderPolicy(t, strings.Replace(discountException, edit[0], edit[1], 1), "", "2026-10-24")
			if err != nil {
				t.Fatal(err)
			}
			if len(res.Findings) != 4 || len(res.Accepted) != 0 {
				t.Fatalf("findings %d accepted %d", len(res.Findings), len(res.Accepted))
			}
		})
	}
	// A location with no line covers every line of the path.
	res, err := reviewUnderPolicy(t, strings.Replace(discountException, "billing/discount.go:5", "billing/discount.go", 1), "", "2026-10-24")
	if err != nil || len(res.Accepted) != 1 {
		t.Fatalf("accepted %d err %v", len(res.Accepted), err)
	}
}

// TestAnExceptionWithoutAnOwnerOrAnExpiryIsRefused: each missing field is refused on load, by its key.
func TestAnExceptionWithoutAnOwnerOrAnExpiryIsRefused(t *testing.T) {
	for _, field := range exceptionFields {
		t.Run(field, func(t *testing.T) {
			var kept []string
			for _, line := range strings.Split(strings.TrimSpace(discountException), "\n") {
				if !strings.HasPrefix(line, "exception.discount."+field+":") {
					kept = append(kept, line)
				}
			}
			_, err := loadPolicy(change.ProcessPolicy(), writeFile(t, "policy.yaml", strings.Join(kept, "\n")+"\n"))
			ce, ok := err.(*ConfigError)
			if !ok || ce.Field != "exception.discount."+field || !strings.Contains(ce.Reason, "missing") {
				t.Fatalf("err %v", err)
			}
		})
	}
	// An empty value is as missing as a missing line.
	_, err := loadPolicy(change.ProcessPolicy(), writeFile(t, "policy.yaml", strings.Replace(discountException, "finance-platform", "''", 1)))
	if ce, ok := err.(*ConfigError); !ok || ce.Field != "exception.discount.owner" {
		t.Fatalf("empty owner: %v", err)
	}
}

// TestAnExceptionWithABadDateOrKeyIsRefused covers the rest of the exception grammar.
func TestAnExceptionWithABadDateOrKeyIsRefused(t *testing.T) {
	cases := map[string]string{
		"date":       strings.Replace(discountException, "2026-12-31", "31/12/2026", 1),
		"field":      discountException + "exception.discount.colour: red\n",
		"id":         strings.Replace(discountException, "exception.discount.", "exception.Disc_ount.", 1),
		"location":   strings.Replace(discountException, "billing/discount.go:5", ":5", 1),
		"duplicate":  discountException + "exception.discount.owner: someone-else\n",
		"policy key": "timeout: 30s\n",
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := loadPolicy(change.ProcessPolicy(), writeFile(t, "policy.yaml", body)); err == nil {
				t.Fatal("not refused")
			} else if _, ok := err.(*ConfigError); !ok || !strings.Contains(err.Error(), "policy.yaml") {
				t.Fatalf("err %v", err)
			}
		})
	}
}

// TestAnExceptionInProvenYamlIsRefused: exceptions live in the policy file only.
func TestAnExceptionInProvenYamlIsRefused(t *testing.T) {
	_, err := parseConfig([]byte(discountException))
	ce, ok := err.(*ConfigError)
	if !ok || !strings.Contains(ce.Reason, "policy file only") {
		t.Fatalf("err %v", err)
	}
	if _, err := reviewUnderPolicy(t, "fallback_class: backend\n", discountException, "2026-10-24"); err == nil {
		t.Fatal("a repository carried its own exception")
	}
}

// TestAMissingOrUnreadablePolicyIsRefusedAtLoad: Open refuses before any review.
func TestAMissingOrUnreadablePolicyIsRefusedAtLoad(t *testing.T) {
	for name, path := range map[string]string{
		"missing":   filepath.Join(t.TempDir(), "absent.yaml"),
		"directory": t.TempDir(),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Open(Options{Sample: true, LogDir: t.TempDir(), PolicyPath: path})
			ce, ok := err.(*ConfigError)
			if !ok || ce.Field != EnvPolicy || ce.Value != path {
				t.Fatalf("err %v", err)
			}
		})
	}
}

// TestAPolicyInsideTheTargetIsRefused: the change under review could edit it. A link into the target
// is as inside as the file itself.
func TestAPolicyInsideTheTargetIsRefused(t *testing.T) {
	target := t.TempDir()
	inside := filepath.Join(target, "policy.yaml")
	if err := os.WriteFile(inside, []byte("fallback_class: backend\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "link.yaml")
	if err := os.Symlink(inside, link); err != nil {
		t.Fatal(err)
	}
	for name, path := range map[string]string{"file": inside, "link": link} {
		t.Run(name, func(t *testing.T) {
			rev, err := Open(Options{Sample: true, LogDir: t.TempDir(), PolicyPath: path})
			if err != nil {
				t.Fatal(err)
			}
			_, err = rev.Review(t.Context(), Request{Target: target})
			ce, ok := err.(*ConfigError)
			if !ok || ce.Field != EnvPolicy || !strings.Contains(ce.Reason, "inside the repository under review") {
				t.Fatalf("err %v", err)
			}
		})
	}
}

// TestTheExampleOfPolicyMdLoads reads the example out of docs/POLICY.md and loads it, so the page cannot
// teach a policy the reader refuses.
func TestTheExampleOfPolicyMdLoads(t *testing.T) {
	page, err := os.ReadFile(filepath.Join(moduleRoot(t), "docs", "POLICY.md"))
	if err != nil {
		t.Fatal(err)
	}
	_, after, ok := strings.Cut(string(page), "## Example")
	if !ok {
		t.Fatal("no example")
	}
	parts := strings.Split(after, "```")
	if len(parts) < 3 {
		t.Fatal("no fenced example")
	}
	p, err := loadPolicy(change.ProcessPolicy(), writeFile(t, "policy.yaml", parts[1]))
	if err != nil {
		t.Fatal(err)
	}
	if len(p.exceptions) != 1 || p.exceptions[0].Line != 5 || p.exceptions[0].Owner != "payments-platform" || !p.settings.Doubled["security"] {
		t.Fatalf("%+v", p)
	}
	if _, err := p.parseLocal([]byte("doubled_classes: off\n")); err == nil {
		t.Fatal("the weaker value of the page was applied")
	}
	if _, err := p.parseLocal([]byte("doubled_classes: security, docs\n")); err != nil {
		t.Fatal(err)
	}
}

// TestNoPolicyChangesNothing: with no policy a review has no policy line, no accepted finding and no
// expired exception.
func TestNoPolicyChangesNothing(t *testing.T) {
	res := reviewSample(t, "", moduleRoot(t))
	if res.PolicyStatement != "" || len(res.Accepted) != 0 || len(res.Expired) != 0 || len(res.Findings) != 4 {
		t.Fatalf("%q accepted %d expired %d findings %d", res.PolicyStatement, len(res.Accepted), len(res.Expired), len(res.Findings))
	}
}
