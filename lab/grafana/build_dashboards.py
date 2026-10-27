#!/usr/bin/env python3
"""Writes the lab's six Grafana dashboards (lab/grafana/dashboards/*.json) so they share one style.

Run it from anywhere:  python3 lab/grafana/build_dashboards.py
The JSON it writes is what Grafana provisions (lab/grafana/proven-dashboards.yaml); a test regenerates it and holds the
committed files equal, so a change is made here and the JSON follows, never the other way.

The style, the same on all six (lab/dashboards_test.go and lab/dashboards_style_test.go hold each line of it):
- the lab's default is the light theme, with colours that mean the same on every dashboard and read in dark too: blue is a
  plain count, green is good, red is bad, yellow is "look at this". Colour always has a word beside it;
- every dashboard opens with ONE sentence a visitor reads in three seconds ("5 of 7 pull requests blocked"), a stat whose
  unit is the rest of the sentence; then at most seven more panels, each with a plain-words description;
- every number counts one thing. A review is named by its pull request: the lab gives each pull request one instance id
  ("pr<number>-<fixture>", lab/seed/seed.sh), so a pull request reviewed again keeps one series, and the latest review is
  the one counted. A count of reviews cannot pass the number of pull requests;
- a stat whose zero means "nothing yet" is blue, never red or yellow, and a stat that says a plural has the singular for 1;
- a bar chart names every bar by its series legend (never "Value #A"), has big labels and values, and when the query has no
  series it says why in one sentence (noValue), never a grey bar with a zero;
- a table has no noValue sentence (Grafana would print it in every empty cell); it is tall enough for every row it can
  hold, and has no column a stranger cannot read (only the columns named in the panel are kept);
- a duration column's unit is the unit its data source returns (Tempo: nanoseconds, Prometheus histograms: seconds, the
  Nightshift importer: milliseconds); in sample mode the agents answer from recorded answers, which the panel says;
- the default time range is half an hour, and a chart steps in minutes: the lab's reviews and the replayed receipts all
  happen within a quarter of an hour, so they spread over the chart instead of standing in one bar.

What each dashboard binds to is what Proven emits today (internal/telemetry/telemetry.go) and what the lab's importer loads
into Loki (lab/seed/nightshift-import.sh). A metric that does not exist yet is named in a text panel, never queried.
"""
import json
import math
import os

HERE = os.path.dirname(os.path.abspath(__file__))
OUT = os.environ.get("DASHBOARDS_OUT") or os.path.join(HERE, "dashboards")
PRS = os.path.join(HERE, "..", "seed", "prs.tsv")

PROM = {"type": "prometheus", "uid": "${ds_prom}"}
TEMPO = {"type": "tempo", "uid": "${ds_tempo}"}
LOKI = {"type": "loki", "uid": "${ds_loki}"}

GREEN, RED, YELLOW, BLUE = "green", "red", "yellow", "blue"


def demo_rows():
    """(number, fixture, title) of each demo pull request, in the order the seed opens them."""
    out = []
    with open(PRS, encoding="utf-8") as f:
        for line in f:
            if not line.strip() or line.startswith("#"):
                continue
            cols = line.rstrip("\n").split("\t")
            out.append((len(out) + 1, cols[1], cols[4]))
    return out


ROWS = demo_rows()

# Every raw value Proven emits on a label a dashboard shows, and the words a visitor reads instead (the owner's rule of
# 27 Oct: no raw field or enum name reaches a screen). The values come from internal/telemetry/telemetry.go (reviewers,
# drop reasons), internal/pipeline/pipeline.go (agents, reviewer verdicts, disagreement kinds) and the proof levels of the
# output formats; lab/dashboards_words_test.go holds this table to those sources and to every dashboard.
HUMAN = {
    "proven_agent": {"planner": "Planner", "investigator": "Investigator", "prover": "Prover", "skeptic": "Skeptic",
                     "fixer": "Fixer", "judge": "Judge", "writer": "Writer", "autofix": "Autofix"},
    "proven_reviewer": {"default": "default reviewer", "special": "second reviewer", "shared": "both reviewers"},
    "proven_proof_level": {"proved-by-tool": "Proved by a tool", "proved-by-test": "Proved by a test",
                           "proved-by-data": "Proved by data",
                           "agreed-by-two-models-not-proven": "Agreed by two models, not proven"},
    "reason": {"no_evidence": "No evidence", "refuted": "Refuted by the Skeptic", "suppressed": "Already refuted earlier",
               "other": "Another reason"},
    "verdict": {"upheld": "upheld", "refuted": "refuted", "not_attempted": "not attempted"},
    "kind": {"verdict": "Judged differently", "one_reviewer": "Named by one reviewer only"},
    "proven_class": {"security": "Security", "backend": "Backend", "mobile": "Mobile", "user interface": "User interface",
                     "docs": "Documentation"},
    "change": {fx: title for _, fx, title in ROWS},
}

# A review's instance id ("pr7-project-health") as a visitor reads it: "#7 Add a very young dependency".
REVIEW_TEXT = {"pr%d-%s" % (n, fx): "#%d %s" % (n, title) for n, fx, title in ROWS}
REVIEW_WORDS = [{"type": "value", "options": {k: {"text": v} for k, v in REVIEW_TEXT.items()}}]


def human(expr, *labels):
    """expr with each raw value of each label replaced by the words in HUMAN, in PromQL, so legends, bars and tooltips read
    as words. A value not in the table is left as it is, and the words test fails on it."""
    for label in labels:
        for raw, words in HUMAN[label].items():
            expr = 'label_replace(%s, "%s", "%s", "%s", "%s")' % (expr, label, words, label, raw)
    return expr


def value_words(*labels):
    """The same table as Grafana value mappings, for a table column that shows a label's raw value."""
    opts = {}
    for label in labels:
        for raw, words in HUMAN[label].items():
            opts[raw] = {"text": words}
    return [{"type": "value", "options": opts}]


def demo_fixtures():
    """The fixtures of the demo pull requests, by the verdict the manifest gives them."""
    clean, defect = [], []
    with open(PRS, encoding="utf-8") as f:
        for line in f:
            if not line.strip() or line.startswith("#"):
                continue
            cols = line.rstrip("\n").split("\t")
            (clean if cols[3] == "clean" else defect).append(cols[1])
    return clean, defect


CLEAN, DEFECT = demo_fixtures()

N_PR = len(CLEAN) + len(DEFECT)
# The review's instance id is "pr<number>-<fixture>" (the workflow sets it, lab/seed/seed.sh): one per pull request, so a
# fixture name in the id is how a review is known for a clean or a defect pull request.
ALL_RE = "pr[0-9]+-.+"
CLEAN_RE = "pr[0-9]+-(" + "|".join(CLEAN) + ")"
RANGE = "[$__range]"
SIZE = {"titleSize": 16, "valueSize": 16}   # the labels and values of a bar: readable in five seconds on a video frame


def thresholds(*steps):
    """steps: (value or None, colour) from the base up."""
    return {"mode": "absolute", "steps": [{"color": c, "value": v} for v, c in steps]}


def colour_at(steps, v):
    c = steps[0][1]
    for val, col in steps[1:]:
        if val is not None and v >= val:
            c = col
    return c


def target(expr, ds, legend=None, instant=True, ref="A", **extra):
    t = {"refId": ref, "datasource": ds, "expr": expr, "instant": instant, "range": not instant}
    if legend is not None:
        t["legendFormat"] = legend
    t.update(extra)
    return t


# What a stat says when its query has no answer, instead of a zero: the source has not reported to this lab.
NOT_MEASURED = "not measured yet: no review has reported"
NOT_LOADED = "none loaded: make nightshift-receipts"


def panel(pid, ptype, title, desc, x, y, w, h, targets=None, **kw):
    p = {"id": pid, "type": ptype, "title": title, "description": desc,
         "gridPos": {"x": x, "y": y, "w": w, "h": h}, "transparent": False}
    if targets is not None:
        p["datasource"] = targets[0]["datasource"]
        p["targets"] = targets
    p.update(kw)
    return p


def stat(pid, title, desc, expr, x, y, w, h, steps, word, zero=None, one=None, ds=PROM, size=40, absent=NOT_MEASURED):
    """A big number. word is what follows the number ("blocked", "finished"); zero is the words shown instead of a bare 0,
    in the colour of the base step (so "0 · clean" is green and "0 · none loaded" is blue); one is the sentence for exactly 1
    where word is a plural. absent is what the panel says when its query has no answer: no data is never a zero."""
    mappings = []
    if zero:
        mappings.append({"type": "value", "options": {"0": {"text": zero, "color": steps[0][1]}}})
    if one:
        mappings.append({"type": "value", "options": {"1": {"text": one, "color": colour_at(steps, 1)}}})
    return panel(
        pid, "stat", title, desc, x, y, w, h, [target(expr, ds)],
        fieldConfig={"defaults": {"unit": "suffix:" + word, "color": {"mode": "thresholds"}, "thresholds": thresholds(*steps),
                                  "mappings": mappings, "decimals": 0, "noValue": absent}, "overrides": []},
        options={"reduceOptions": {"calcs": ["lastNotNull"], "values": False}, "colorMode": "value", "graphMode": "none",
                 "textMode": "value", "justifyMode": "center", "orientation": "auto",
                 "text": {"valueSize": size, "titleSize": 15}},
    )


def headline(pid, title, desc, expr, steps, word, zero, one=None, ds=PROM, x=0, w=24, absent=NOT_MEASURED):
    """The sentence a dashboard opens with: a stat whose number and unit read as one line."""
    return stat(pid, title, desc, expr, x, 0, w, 5, steps, word, zero=zero, one=one, ds=ds, size=48, absent=absent)


def measured(expr, presence):
    """expr's answer, or a 0 only when presence answers: a series that exists once the source has reported (a review
    finished, a model was called, a receipt was loaded). With neither, the query has no answer and the panel says absent:
    a metric that never reached the lab must not read as a measured "none"."""
    return "%s or (%s * 0)" % (expr, presence)


def reviews_done(instance):
    """The reviews whose last agent, the Writer, finished: the presence behind a count a finished review would have
    emitted (a clean review emits no finding series, so its zero is measured only once it finished)."""
    return 'count(last_over_time(proven_agent_state{proven_agent="writer",instance%s}%s) == 2)' % (instance, RANGE)


def agents_seen(instance):
    return 'count(last_over_time(proven_agent_state{instance%s}%s))' % (instance, RANGE)


def bars(pid, title, desc, expr, ds, label, reason, x, y, w, h, color=BLUE, steps=None, unit="short"):
    """A bar chart of one label. Every bar is named by the series legend ({{label}}); when the query has no series the panel
    says `reason` in one sentence (noValue), and draws no bar. steps colours the bars by value instead (green below one, red
    from one, for a count of failures)."""
    if steps:
        colour, thr = {"mode": "thresholds"}, thresholds(*steps)
    else:
        colour, thr = {"mode": "fixed", "fixedColor": color}, thresholds((None, color))
    return panel(
        pid, "bargauge", title, desc, x, y, w, h, [target(expr, ds, legend="{{%s}}" % label)],
        fieldConfig={"defaults": {"unit": unit, "min": 0, "noValue": reason, "color": colour, "thresholds": thr,
                                  "displayName": "${__field.labels.%s}" % label}, "overrides": []},
        options={"reduceOptions": {"calcs": ["lastNotNull"], "values": False}, "orientation": "horizontal",
                 "displayMode": "gradient", "valueMode": "color", "showUnfilled": False, "minVizHeight": 32, "maxVizHeight": 48,
                 "minVizWidth": 8, "namePlacement": "top", "sizing": "manual", "legend": {"showLegend": False}, "text": SIZE},
    )


def series(pid, title, desc, targets, x, y, w, h, no_value, unit="short", style="bars"):
    return panel(
        pid, "timeseries", title, desc, x, y, w, h, targets,
        fieldConfig={"defaults": {"unit": unit, "min": 0, "noValue": no_value, "color": {"mode": "palette-classic"},
                                  "custom": {"drawStyle": style, "lineWidth": 2, "fillOpacity": 60, "showPoints": "never",
                                             "stacking": {"mode": "normal", "group": "A"}, "axisPlacement": "auto"}},
                     "overrides": []},
        options={"legend": {"displayMode": "list", "placement": "bottom", "showLegend": True},
                 "tooltip": {"mode": "multi", "sort": "desc"}},
    )


def text(pid, title, desc, markdown, x, y, w, h):
    return {"id": pid, "type": "text", "title": title, "description": desc, "gridPos": {"x": x, "y": y, "w": w, "h": h},
            "options": {"mode": "markdown", "content": markdown}}


def table_h(rows):
    """The panel height (grid units: 30 px and an 8 px gap) that shows `rows` rows, the header and the title, with room to
    spare: a row is about 34 px, so a table never cuts its last row in half."""
    return int(math.ceil(((rows + 1) * 34 + 58) / 38.0))


YES_NO = [{"type": "value", "options": {"true": {"text": "yes"}, "false": {"text": "no"}}},
          {"type": "special", "options": {"match": "null+nan", "result": {"text": "not recorded"}}}]


def spans_table(pid, title, desc, traceql, x, y, w, limit, select):
    """A Tempo span search shown as a table, the first `limit` spans. The duration is a bar, so the table reads as a
    timeline. Tempo returns a duration in nanoseconds. Only the columns named in select are kept (Tempo adds raw ones, such
    as the service instance id), under the names select gives. The table has no noValue sentence: Grafana prints it in every
    empty cell. The panel is tall enough for all `limit` rows."""
    yes_no = [{"matcher": {"id": "byName", "options": n}, "properties": [{"id": "mappings", "value": YES_NO}]}
              for n in select.values() if n in ("Two reviewers", "Cut short")]
    yes_no += [{"matcher": {"id": "byName", "options": n}, "properties": [{"id": "mappings", "value": value_words(label)}]}
               for n, label in (("Agent", "proven_agent"), ("Reviewer", "proven_reviewer"), ("Review class", "proven_class"))
               if n in select.values()]
    if "Review" in select.values():
        yes_no.append({"matcher": {"id": "byName", "options": "Review"}, "properties": [{"id": "mappings", "value": REVIEW_WORDS}]})
    return panel(
        pid, "table", title, desc, x, y, w, table_h(limit),
        [{"refId": "A", "datasource": TEMPO, "queryType": "traceql", "query": traceql, "tableType": "spans", "limit": limit}],
        fieldConfig={"defaults": {"custom": {"align": "left", "cellOptions": {"type": "auto", "wrapText": True}}},
                     "overrides": [{"matcher": {"id": "byName", "options": "Duration"},
                                    "properties": [{"id": "unit", "value": "ns"}, {"id": "decimals", "value": 1}]}] + yes_no},
        options={"showHeader": True, "cellHeight": "sm", "sortBy": [{"displayName": "Start time", "desc": False}]},
        transformations=[
            {"id": "filterFieldsByName", "options": {"include": {"names": list(select.keys())}}},
            {"id": "organize", "options": {"renameByName": select}}],
    )


def dashboard(uid, title, desc, panels, variables=None, tr="now-30m"):
    assert len(panels) <= 8, "%s has %d panels" % (uid, len(panels))
    datasources = [
        {"name": "ds_prom", "type": "datasource", "query": "prometheus", "hide": 2, "current": {}},
        {"name": "ds_tempo", "type": "datasource", "query": "tempo", "hide": 2, "current": {}},
        {"name": "ds_loki", "type": "datasource", "query": "loki", "hide": 2, "current": {}},
    ]
    return {
        "uid": uid, "title": title, "description": desc, "tags": ["proven"], "schemaVersion": 39, "version": 1,
        "editable": False, "graphTooltip": 1, "refresh": "10s", "time": {"from": tr, "to": "now"},
        "timepicker": {"refresh_intervals": ["5s", "10s", "30s", "1m"]},
        "style": "light",
        "links": [{"type": "dashboards", "title": "Proven dashboards", "tags": ["proven"], "asDropdown": False,
                   "includeVars": False, "keepTime": True, "icon": "external link", "targetBlank": False}],
        "templating": {"list": datasources + (variables or [])},
        "panels": panels,
    }


def write(d):
    os.makedirs(OUT, exist_ok=True)
    with open(os.path.join(OUT, d["uid"] + ".json"), "w", encoding="utf-8") as f:
        json.dump(d, f, indent=2, ensure_ascii=False)
        f.write("\n")


# The words that follow a number, by what a count means. "bad" counts are green at zero and say so in words.
BAD = " · bad"
SAMPLE_NOTE = ("In a sample-mode lab the agents answer from recorded answers and the Prover runs no test, so times of a few "
               "microseconds are the recorded answers, not real work. In a live lab they are real.")


def by_pr(expr):
    """expr counted once per review: the latest review of each pull request (one instance id each)."""
    return "max by (instance) (%s)" % expr


def proof_bars(pid, scope, x, y, w, h):
    return bars(pid, "Findings by level of proof", "How well each kept finding is proved: the strongest evidence it carries.",
                human("sum by (proven_proof_level) (last_over_time(proven_review_findings_total%s%s))" % (scope, RANGE), "proven_proof_level"), PROM,
                "proven_proof_level", "No finding kept: the reviews are still running, or every change was clean", x, y, w, h)


# ---- 1. Overview -----------------------------------------------------------------------------------------------------

def review_labels(expr):
    """expr with the pull request number and the change name cut out of the instance id, as labels pr and change."""
    return ('label_replace(label_replace(%s, "pr", "$1", "instance", "pr([0-9]+)-.*"), '
            '"change", "$1", "instance", "pr[0-9]+-(.*)")' % expr)


def overview():
    n_def, n_clean = len(DEFECT), len(CLEAN)
    sel = '{instance=~"%s"}' % ALL_RE
    findings = "sum by (instance) (last_over_time(proven_review_findings_total%s%s))" % (sel, RANGE)
    per_pr = lambda e: "max by (pr, change) (%s)" % review_labels(e)
    table = panel(
        5, "table", "Every pull request: how its review ended",
        "One row per demo pull request, its latest review. Verdict: clean when no finding was kept, blocked when one was. "
        "Findings and Claims dropped count that one review.",
        0, 9, 16, table_h(N_PR + 1),
        [target(per_pr('last_over_time(proven_agent_state{proven_agent="writer",instance=~"%s"}%s)' % (ALL_RE, RANGE)), PROM, ref="A", format="table"),
         target(per_pr(findings), PROM, ref="B", format="table"),
         target(per_pr("sum by (instance) (last_over_time(proven_review_claims_dropped_total%s%s))" % (sel, RANGE)), PROM, ref="C",
                format="table"),
         target("clamp_max(%s, 1)" % per_pr(findings), PROM, ref="D", format="table")],
        fieldConfig={"defaults": {"custom": {"align": "left"}},
                     "overrides": [
                         {"matcher": {"id": "byName", "options": "Verdict"}, "properties": [
                             {"id": "mappings", "value": [{"type": "value", "options": {"0": {"text": "clean", "color": GREEN},
                                                                                         "1": {"text": "blocked", "color": RED}}}]},
                             {"id": "custom.cellOptions", "value": {"type": "color-text"}}]},
                         {"matcher": {"id": "byName", "options": "State"}, "properties": [
                             {"id": "mappings", "value": [
                                 {"type": "value", "options": {"0": {"text": "waiting", "color": YELLOW}, "1": {"text": "working", "color": BLUE},
                                                               "2": {"text": "finished", "color": GREEN}, "3": {"text": "failed", "color": RED}}}]},
                             {"id": "custom.cellOptions", "value": {"type": "color-text"}}]},
                         {"matcher": {"id": "byName", "options": "Pull request"}, "properties": [{"id": "decimals", "value": 0}]},
                         {"matcher": {"id": "byName", "options": "Change"}, "properties": [
                             {"id": "mappings", "value": value_words("change")}, {"id": "custom.width", "value": 340}]},
                         {"matcher": {"id": "byName", "options": "Claims dropped"}, "properties": [
                             {"id": "mappings", "value": [{"type": "special", "options": {"match": "null+nan", "result": {"text": "none recorded"}}}]}]}]},
        options={"showHeader": True, "cellHeight": "sm", "sortBy": [{"displayName": "Pull request", "desc": False}]},
        transformations=[
            {"id": "merge", "options": {}},
            {"id": "convertFieldType", "options": {"conversions": [{"targetField": "pr", "destinationType": "number"}]}},
            {"id": "organize", "options": {
                "excludeByName": {"Time": True},
                "indexByName": {"pr": 0, "change": 1, "Value #D": 2, "Value #B": 3, "Value #C": 4, "Value #A": 5},
                "renameByName": {"pr": "Pull request", "change": "Change", "Value #D": "Verdict", "Value #B": "Findings",
                                 "Value #C": "Claims dropped", "Value #A": "State"}}},
        ],
    )
    return dashboard("proven-overview", "Proven: every pull request at a glance",
                     "Every pull request the lab reviewed (its latest review), whether Proven blocked the broken ones, "
                     "and whether it raised a false alarm on the clean ones.", [
        headline(1, "Pull requests blocked",
                 "Pull requests on which Proven kept at least one finding, out of the %d demo pull requests (%d broken, %d clean). "
                 "Every kept finding carries evidence." % (N_PR, n_def, n_clean),
                 measured("count(%s > 0)" % findings, reviews_done('=~"%s"' % ALL_RE)),
                 [(None, BLUE), (1, YELLOW), (n_def, GREEN)], " of %d pull requests blocked" % N_PR, zero="0 of %d pull requests blocked" % N_PR),
        stat(2, "Reviews finished", "Pull requests whose review is done: its last agent, the Writer, finished.",
             measured('count(%s == 2)' % by_pr('last_over_time(proven_agent_state{proven_agent="writer",instance=~"%s"}%s)' % (ALL_RE, RANGE)),
                       agents_seen('=~"%s"' % ALL_RE)),
             0, 5, 8, 4, [(None, BLUE), (1, YELLOW), (N_PR, GREEN)], " of %d finished" % N_PR, zero="0 of %d finished" % N_PR),
        stat(3, "False alarms on the clean changes",
             "Findings Proven kept on the %d clean demo pull requests. None is good: a clean change must get a clean verdict." % n_clean,
             measured("sum(last_over_time(proven_review_findings_total{instance=~\"%s\"}%s))" % (CLEAN_RE, RANGE),
                       reviews_done('=~"%s"' % CLEAN_RE)),
             8, 5, 8, 4, [(None, GREEN), (1, RED)], BAD, zero="0 · none, good"),
        stat(4, "Agents that failed",
             "Agent stages that ended failed in the latest review of any pull request. None is good: a failed agent means part of the review did not run.",
             measured('count(last_over_time(proven_agent_state{instance=~"%s"}%s) == 3)' % (ALL_RE, RANGE), agents_seen('=~"%s"' % ALL_RE)),
             16, 5, 8, 4, [(None, GREEN), (1, RED)], BAD, zero="0 · none, good"),
        text(7, "What the numbers count", "Every number on this page counts the same thing.",
             "Each number counts **pull requests**, using the **latest review** of each. %d demo pull requests: %d broken, %d clean. "
             "A review that has not finished is counted as it stands." % (N_PR, n_def, n_clean), 0, 9 + table_h(N_PR + 1), 24, 3),
        table,
        proof_bars(6, "{instance=~\"%s\"}" % ALL_RE, 16, 9, 8, table_h(N_PR + 1)),
    ])


# ---- 2. One review ---------------------------------------------------------------------------------------------------

def one_review():
    # A custom variable: its options are the demo pull requests, each shown by number and title, valued by its instance id.
    var = {"name": "run", "label": "Pull request", "type": "custom", "includeAll": False, "multi": False,
           "query": ",".join("%s : %s" % (v.replace(",", " "), k) for k, v in REVIEW_TEXT.items()),
           "options": [{"text": v, "value": k, "selected": i == len(REVIEW_TEXT) - 1}
                       for i, (k, v) in enumerate(REVIEW_TEXT.items())],
           "current": {"text": list(REVIEW_TEXT.values())[-1], "value": list(REVIEW_TEXT.keys())[-1]}}
    spans = spans_table(
        4, "The agents, in order, and how long each took",
        "Each row is one agent of this review, from the trace. The bar is how long it ran. The two reviewers each run their own Planner and Skeptic. "
        + SAMPLE_NOTE,
        '{ resource.service.instance.id = "$run" && span.proven.agent != "" } | select(span.proven.agent, span.proven.reviewer, span.gen_ai.request.model)',
        0, 5, 14, 14,
        {"Start time": "Start time", "proven.agent": "Agent", "proven.reviewer": "Reviewer", "gen_ai.request.model": "Model", "Duration": "Duration"})
    verdicts = bars(
        5, "Skeptic verdicts, by reviewer",
        "How each reviewer's Skeptic judged each finding: refuted, upheld, or not attempted. Where the two differ, the reviewers disagree.",
        'label_join(%s, "who", " · ", "proven_reviewer", "verdict")' % human('sum by (proven_reviewer, verdict) (last_over_time(proven_reviewer_verdicts_total{instance="$run"}%s))' % RANGE, "proven_reviewer", "verdict"),
        PROM, "who", "No finding was kept, so the Skeptic had nothing to judge", 14, 5, 10, 8)
    return dashboard("proven-review", "Proven: one review, step by step",
                     "Pick a review at the top. See each agent in order, how long it took, and how the two reviewers judged each finding.", [
        headline(1, "Findings this review kept", "Findings this review kept. None means a clean verdict.",
                 measured('sum(last_over_time(proven_review_findings_total{instance="$run"}%s))' % RANGE, reviews_done('="$run"')),
                 [(None, GREEN), (1, RED)], " findings kept · blocked", zero="0 findings kept · clean", one="1 finding kept · blocked", x=0, w=12),
        stat(2, "Claims the Judge dropped", "Claims dropped for being refuted, unproved, or already refuted in an earlier review.",
             measured('sum(last_over_time(proven_review_claims_dropped_total{instance="$run"}%s))' % RANGE, reviews_done('="$run"')),
             12, 0, 6, 5, [(None, BLUE)], " dropped", zero="0 · none dropped"),
        stat(3, "Model calls that failed",
             "Calls to a model that returned an error. In sample mode a call the recorded set has no answer for counts here too; the review goes on.",
             measured('sum(last_over_time(proven_model_calls_total{instance="$run",outcome="error"}%s))' % RANGE,
                       'count(last_over_time(proven_model_calls_total{instance="$run"}%s))' % RANGE),
             18, 0, 6, 5, [(None, GREEN), (1, YELLOW)], " · check", zero="0 · none failed"),
        spans,
        verdicts,
        proof_bars(6, '{instance="$run"}', 14, 13, 10, 6),
        bars(7, "Where the two reviewers disagree", "Findings one reviewer named and the other did not, or that they judged differently.",
             human('sum by (kind) (last_over_time(proven_reviewer_disagreements_total{instance="$run"}%s))' % RANGE, "kind"), PROM, "kind",
             "The two reviewers agree on this review", 0, 5 + table_h(14), 12, 6, color=YELLOW),
        text(8, "Each finding, with its evidence", "Where to read the evidence behind a finding.",
             "The evidence of each finding (the query that found it, or the test that fails on the change and passes before it) "
             "is in the **Report** page of its pull request. Open the lab's start page and use the **Report** link in the table.",
             12, 5 + table_h(14), 12, 6),
    ], variables=[var])


# ---- 3. Evidence -----------------------------------------------------------------------------------------------------

def evidence():
    return dashboard("proven-evidence", "Proven: the evidence behind each finding",
                     "Where findings get their proof. What Proven does not yet report as a number is named in the two text panels.", [
        headline(1, "Findings kept, with their evidence",
                 "Findings kept across the latest review of each pull request. The Judge keeps a finding only when a query or a test supports it.",
                 measured('sum(last_over_time(proven_review_findings_total{instance=~"%s"}%s))' % (ALL_RE, RANGE), reviews_done('=~"%s"' % ALL_RE)),
                 [(None, BLUE)], " findings kept, each with a query or a test as evidence", zero="0 findings kept: nothing to prove",
                 one="1 finding kept, with a query or a test as evidence", x=0, w=12),
        stat(2, "Claims dropped", "Claims the Judge dropped for any reason: no query or test supported them, the Skeptic refuted them, "
             "or they were already refuted earlier. This is Proven refusing to guess.",
             measured('sum(last_over_time(proven_review_claims_dropped_total{instance=~"%s"}%s))' % (ALL_RE, RANGE), reviews_done('=~"%s"' % ALL_RE)),
             12, 0, 6, 5, [(None, BLUE)], " dropped", zero="0 · none dropped"),
        stat(3, "Evidence agents that failed",
             "Investigator or Prover stages that ended failed. None is good: a failed one means evidence was not gathered.",
             measured('count(last_over_time(proven_agent_state{proven_agent=~"investigator|prover",instance=~"%s"}%s) == 3)' % (ALL_RE, RANGE),
                       agents_seen('=~"%s"' % ALL_RE)),
             18, 0, 6, 5, [(None, GREEN), (1, RED)], BAD, zero="0 · none, good"),
        text(7, "Zetaris queries: not measured yet",
             "What would be shown here, and the numbers that are missing.",
             "**Not emitted by Proven yet.** Proven asks Zetaris for each advisory check, but does not yet count it. "
             "Missing, so nothing is drawn here:\n\n- queries per table (OSV, KEV, EPSS)\n- rows returned\n- query duration\n- queries Zetaris refused", 0, 5, 12, 6),
        text(4, "Prover runs: not measured yet",
             "What would be shown here, and the numbers that are missing.",
             "**Not emitted by Proven yet.** The Prover runs a test on the base and on the change, with the network off. "
             "Missing, so nothing is drawn here:\n\n- runs, with the outcome on the base and on the change\n- run duration\n- runs that had the network off", 12, 5, 12, 6),
        proof_bars(5, '{instance=~"%s"}' % ALL_RE, 0, 11, 24, 6),
        spans_table(6, "Investigator and Prover runs, and how long each took",
                    "Each row is one run of the Investigator (turns a risk into a query) or the Prover (asks for a test). The bar is how long it ran. "
                    + SAMPLE_NOTE,
                    '{ span.proven.agent = "investigator" || span.proven.agent = "prover" } | select(span.proven.agent, span.proven.reviewer)',
                    0, 17, 24, 20, {"Start time": "Start time", "proven.agent": "Agent", "proven.reviewer": "Reviewer", "Duration": "Duration"}),
    ])


# ---- 4. Models and trust ---------------------------------------------------------------------------------------------

def models():
    both = lambda m: 'label_join(%s, "who", " · ", "proven_agent", "proven_reviewer")' % human(
        "sum by (proven_agent, proven_reviewer) (last_over_time(%s%s))" % (m, RANGE), "proven_agent", "proven_reviewer")
    latency = 'label_join(%s, "who", " · ", "proven_agent", "proven_reviewer")' % human(
        "(sum by (proven_agent, proven_reviewer) (last_over_time(gen_ai_client_operation_duration_seconds_sum%s)) / "
        "sum by (proven_agent, proven_reviewer) (last_over_time(gen_ai_client_operation_duration_seconds_count%s)))" % (RANGE, RANGE),
        "proven_agent", "proven_reviewer")
    # The model asked, and "asked → answered" when the provider names the model that answered. In sample mode no answering
    # model is named, so the bar carries the asked model alone.
    asked_answered = (
        'label_replace(label_join(sum by (gen_ai_request_model, gen_ai_response_model) '
        '(last_over_time(gen_ai_client_operation_duration_seconds_count%s)), "route", " → ", "gen_ai_request_model", "gen_ai_response_model"), '
        '"route", "$1", "route", "(.*) → ")' % RANGE)
    asked_answered = 'label_replace(%s, "route", "model not recorded", "route", "")' % asked_answered
    refused = human("sum by (reason) (last_over_time(proven_review_claims_dropped_total%s))" % RANGE, "reason")
    lat = bars(8, "Model latency (sample mode: recorded answers)",
               "Mean time of a model call. In sample mode the answers are recorded, so this is not a measure of any model; "
               "in a live lab it is the real latency of each agent's calls.",
               latency, PROM, "who", "No model call yet", 16, 13, 8, 8, unit="s")
    return dashboard("proven-models", "Proven: which models answered, and what the guardrail checked",
                     "Calls to models by agent and reviewer, which model answered, what the guardrail checked, and what Proven refused to keep.", [
        headline(1, "Model calls answered", "Calls to a model that returned an answer, across all reviews.",
                 measured('sum(last_over_time(proven_model_calls_total{outcome!="error"}%s))' % RANGE, 'count(last_over_time(proven_model_calls_total%s))' % RANGE),
                 [(None, BLUE)], " model calls answered", zero="0 model calls yet", one="1 model call answered", x=0, w=12),
        stat(2, "Model calls that returned an error",
             "A model call that returned an error. In sample mode a call the recorded answers have no answer for is counted here "
             "(generator: no recorded response); the review goes on without it. In a live lab an error is a failed call to the model.",
             measured('sum(last_over_time(proven_model_calls_total{outcome="error"}%s))' % RANGE, 'count(last_over_time(proven_model_calls_total%s))' % RANGE),
             12, 0, 6, 5, [(None, GREEN), (1, YELLOW)], " · check", zero="0 · none failed"),
        stat(3, "Guardrail checks that said unsafe",
             "Model answers or prompts the guardrail judged unsafe. None is good.",
             measured("sum(last_over_time(proven_guardrail_unsafe_total%s))" % RANGE, "count(last_over_time(proven_guardrail_checks_total%s))" % RANGE),
             18, 0, 6, 5, [(None, GREEN), (1, RED)], BAD, zero="0 · none, good"),
        bars(4, "Calls, by agent and reviewer", "Which agent made the calls, and for which reviewer. Planner and Skeptic run once per reviewer.",
             both("proven_model_calls_total"), PROM, "who", "No model call yet", 0, 5, 12, 8),
        bars(5, "Calls, by the model asked (and the one that answered)",
             "The model a call asked for. When the provider names the model that answered, the bar reads asked → answered; a fallback shows "
             "as two different names. Sample mode names no answering model.",
             asked_answered, PROM, "route", "No model call yet", 12, 5, 12, 8),
        bars(6, "Guardrail checks, by side and verdict", "What the guardrail was asked to check (the prompt or the answer) and what it said.",
             'label_join(sum by (side, verdict) (last_over_time(proven_guardrail_checks_total%s)), "check", " · ", "side", "verdict")' % RANGE,
             PROM, "check", "No guardrail check yet", 0, 13, 8, 8),
        bars(7, "What Proven refused to keep, by reason",
             "Claims the Judge dropped: not supported by a query or a test, refuted by the Skeptic, or already refuted in an earlier review.",
             refused, PROM, "reason", "Nothing was dropped yet", 8, 13, 8, 8),
        lat,
    ])


# ---- 5. Meterless engines --------------------------------------------------------------------------------------------

def engines():
    return dashboard("proven-engines", "Proven: the four Meterless engines and what each did",
                     "H-MEM, Scout Intent, Markovian and the World Model are on in this lab. What each one did is shown where Proven reports it.", [
        headline(1, "H-MEM: claims it kept out of the Judge's hands",
                 "Claims H-MEM recognised as already refuted in an earlier review, so they never reached the Judge.",
                 measured('sum(last_over_time(proven_review_claims_dropped_total{reason="suppressed",instance=~"%s"}%s))' % (ALL_RE, RANGE),
                           reviews_done('=~"%s"' % ALL_RE)),
                 [(None, BLUE)], " claims suppressed as already refuted", zero="0 claims suppressed: nothing was refuted earlier",
                 one="1 claim suppressed as already refuted", x=0, w=12),
        text(2, "Scout Intent: not measured yet", "What would be shown here.",
             "**Not emitted by Proven yet.** Scout plans each review, light or full, and can be overridden by a guard. Missing:\n\n"
             "- plans, by light or full\n- guard overrides", 12, 0, 12, 5),
        text(3, "Markovian: not measured yet", "What would be shown here.",
             "**Not emitted by Proven yet.** Markovian reviews a large change in bounded steps. Missing:\n\n- steps taken\n\n"
             "Small changes, like the demo's, take one step.", 0, 5, 12, 5),
        text(4, "World Model: not measured yet", "What would be shown here.",
             "**Not emitted by Proven yet.** The World Model keeps claims, evidence and verdicts as a graph. Missing:\n\n"
             "- claims\n- evidence\n- verdicts recorded", 12, 5, 12, 5),
        spans_table(5, "What each review ran as",
                    "One row per review: its pull request, the review class Scout chose, whether a second reviewer ran, and whether the review was cut short.",
                    '{ name = "proven.review" } | select(resource.service.instance.id, span.proven.class, span.proven.doubled, span.proven.partial, span.proven.findings)',
                    0, 10, 24, 10,
                    {"Start time": "Start time", "service.instance.id": "Review", "proven.class": "Review class", "proven.doubled": "Two reviewers",
                     "proven.partial": "Cut short", "proven.findings": "Findings kept", "Duration": "Duration"}),
    ])


# ---- 6. Nightshift ---------------------------------------------------------------------------------------------------

NS = '{job="nightshift"}'
NS_EMPTY = "No receipts loaded yet. To load the public ones: make nightshift-receipts (docs/LAB.md)."
NS_NONE = "No receipts loaded: make nightshift-receipts"


def nightshift():
    unwrap = lambda f: '| json | unwrap %s | __error__=""' % f
    tokens = " or ".join(
        'label_replace(sum by (model) (sum_over_time(%s %s [$__range])), "series", "$1 · %s", "model", "(.*)")' % (NS, unwrap(f), d)
        for f, d in (("input_tokens", "tokens in"), ("output_tokens", "tokens out")))
    return dashboard("proven-nightshift", "Proven: the model gateway, from Nightshift's signed receipts",
                     "Every call Proven made through Meterless Nightshift: how many carry a signed receipt, to which model, how fast, and which failed and why. "
                     "The receipts are replayed when the lab starts, so time here is when they were loaded. "
                     "None are loaded until the public receipts file is put in place: make nightshift-receipts.", [
        headline(1, "Model calls with a signed receipt that verifies",
                 "Receipts whose Ed25519 signature verified under the pinned key, counted from the receipts loaded into this lab.",
                 measured('sum(count_over_time({job="nightshift", verified="true"}%s))' % RANGE, "sum(count_over_time(%s%s))" % (NS, RANGE)),
                 [(None, BLUE), (1, GREEN)], " model calls carry a signed receipt that verifies",
                 zero="0 receipts loaded: make nightshift-receipts", one="1 model call carries a signed receipt that verifies", ds=LOKI, x=0, w=9, absent=NOT_LOADED),
        stat(2, "Calls made", "Public receipts of Proven's own calls, loaded into this lab: one per model call.",
             "sum(count_over_time(%s%s))" % (NS, RANGE), 9, 0, 5, 5, [(None, BLUE)], " in all", ds=LOKI, absent=NOT_LOADED),
        stat(3, "Calls that failed", "Calls whose receipt says failed. The reasons are in the panel below. None is good.",
             measured('sum(count_over_time({job="nightshift", status="failed"}%s))' % RANGE, "sum(count_over_time(%s%s))" % (NS, RANGE)), 14, 0, 5, 5,
             [(None, GREEN), (1, RED)], " · failed", zero="0 · none failed", ds=LOKI, absent=NOT_LOADED),
        stat(4, "Verification failures",
             "Receipts whose Ed25519 signature did not verify under the pinned key. None is good.",
             measured('sum(count_over_time({job="nightshift", verified="false"}%s))' % RANGE, "sum(count_over_time(%s%s))" % (NS, RANGE)), 19, 0, 5, 5,
             [(None, GREEN), (1, RED)], " · bad", zero="0 · none failed", ds=LOKI, absent=NOT_LOADED),
        series(5, "Calls over time, by model and app", "Calls per minute, for each model that answered and each app that asked.",
               [target("sum by (model, app) (count_over_time(%s[1m]))" % NS, LOKI, legend="{{model}} · {{app}}", instant=False, interval="1m", step="1m")],
               0, 5, 12, 8, NS_EMPTY),
        bars(6, "Why calls failed", "The reason on each failed receipt: the provider's deadline, an empty answer, a refused key, and the rest.",
             'sum by (reason) (count_over_time({job="nightshift", status="failed"}%s))' % RANGE, LOKI, "reason",
             "No call failed (or no receipts are loaded)", 12, 5, 12, 8, steps=[(None, GREEN), (1, RED)]),
        series(7, "Latency per call, median, by model", "Median time of an answered call, per model.",
               [target("quantile_over_time(0.5, %s %s [1m]) by (model)" % (NS.replace('}', ', status="ok"}'), unwrap("latency_ms")), LOKI,
                       legend="{{model}}", instant=False, interval="1m", step="1m")],
               0, 13, 12, 8, NS_EMPTY, unit="ms", style="line"),
        bars(8, "Tokens, by model", "Input and output tokens of the calls, per model.", tokens, LOKI, "series",
             NS_NONE, 12, 13, 12, 8),
    ])


def main():
    for build in (overview, one_review, evidence, models, engines, nightshift):
        write(build())


if __name__ == "__main__":
    main()
