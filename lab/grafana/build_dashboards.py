#!/usr/bin/env python3
"""Writes the lab's six Grafana dashboards (lab/grafana/dashboards/*.json) so they share one style.

Run it from anywhere:  python3 lab/grafana/build_dashboards.py
The JSON it writes is what Grafana provisions (lab/grafana/proven-dashboards.yaml); a test regenerates it and holds the
committed files equal, so a change is made here and the JSON follows, never the other way.

The style, the same on all six: dark theme (set by the lab), a title that says what you see, at most eight panels, big stat
numbers whose colour is always beside a word (a unit such as "false alarms", or a mapped word such as "clean"), no panel
that can be empty without saying why (noValue), a plain-words description on every panel, and a default time range of the
last hour (six hours for Nightshift, whose receipts are replayed when the lab starts).

What each dashboard binds to is what Proven emits today (internal/telemetry/telemetry.go) and what the lab's importer loads
into Loki (lab/seed/nightshift-import.sh). A metric that does not exist yet is named in a text panel, never queried.
"""
import json
import os

HERE = os.path.dirname(os.path.abspath(__file__))
OUT = os.environ.get("DASHBOARDS_OUT") or os.path.join(HERE, "dashboards")
PRS = os.path.join(HERE, "..", "seed", "prs.tsv")

PROM = {"type": "prometheus", "uid": "${ds_prom}"}
TEMPO = {"type": "tempo", "uid": "${ds_tempo}"}
LOKI = {"type": "loki", "uid": "${ds_loki}"}

GREEN, RED, YELLOW, BLUE = "green", "red", "yellow", "blue"


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
# The review's instance id is "<unix time>-pr<number>-<fixture>" (the workflow sets it, lab/seed/seed.sh), so a fixture
# name in the id is how a review is known for a clean or a defect pull request.
CLEAN_RE = "[0-9]+-pr[0-9]+-(" + "|".join(CLEAN) + ")"
DEFECT_RE = "[0-9]+-pr[0-9]+-(" + "|".join(DEFECT) + ")"
RANGE = "[$__range]"


def thresholds(*steps):
    """steps: (value or None, colour) from the base up."""
    return {"mode": "absolute", "steps": [{"color": c, "value": v} for v, c in steps]}


def target(expr, ds, legend=None, instant=True, ref="A", **extra):
    t = {"refId": ref, "datasource": ds, "expr": expr, "instant": instant, "range": not instant}
    if legend is not None:
        t["legendFormat"] = legend
    t.update(extra)
    return t


def panel(pid, ptype, title, desc, x, y, w, h, targets=None, **kw):
    p = {"id": pid, "type": ptype, "title": title, "description": desc,
         "gridPos": {"x": x, "y": y, "w": w, "h": h}, "transparent": False}
    if targets is not None:
        p["datasource"] = targets[0]["datasource"]
        p["targets"] = targets
    p.update(kw)
    return p


def stat(pid, title, desc, expr, x, y, w, h, unit, steps, no_value="0", ds=PROM, mappings=None):
    return panel(
        pid, "stat", title, desc, x, y, w, h, [target(expr, ds)],
        fieldConfig={"defaults": {"unit": unit, "color": {"mode": "thresholds"}, "thresholds": thresholds(*steps),
                                  "noValue": no_value, "mappings": mappings or [], "decimals": 0}, "overrides": []},
        options={"reduceOptions": {"calcs": ["lastNotNull"], "values": False}, "colorMode": "value", "graphMode": "none",
                 "textMode": "value", "justifyMode": "center", "orientation": "auto",
                 "text": {"valueSize": 46, "titleSize": 15}},
    )


def bars(pid, title, desc, targets, x, y, w, h, no_value, unit="short", color=BLUE):
    return panel(
        pid, "bargauge", title, desc, x, y, w, h, targets,
        fieldConfig={"defaults": {"unit": unit, "min": 0, "color": {"mode": "fixed", "fixedColor": color},
                                  "thresholds": thresholds((None, color)), "noValue": no_value, "decimals": 0}, "overrides": []},
        options={"reduceOptions": {"calcs": ["lastNotNull"], "values": False}, "orientation": "horizontal",
                 "displayMode": "gradient", "valueMode": "color", "showUnfilled": True, "minVizHeight": 16,
                 "minVizWidth": 8, "namePlacement": "left", "sizing": "auto", "legend": {"showLegend": False}},
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


def spans_table(pid, title, desc, traceql, x, y, w, h, select, no_value):
    """A Tempo span search shown as a table; the duration column is a bar, so the table reads as a timeline."""
    return panel(
        pid, "table", title, desc, x, y, w, h,
        [{"refId": "A", "datasource": TEMPO, "queryType": "traceql", "query": traceql, "tableType": "spans", "limit": 100}],
        fieldConfig={"defaults": {"noValue": no_value, "custom": {"align": "left", "cellOptions": {"type": "auto"}}},
                     "overrides": [{"matcher": {"id": "byName", "options": "Duration"},
                                    "properties": [{"id": "unit", "value": "ms"},
                                                   {"id": "custom.cellOptions", "value": {"type": "gauge", "mode": "gradient"}},
                                                   {"id": "min", "value": 0},
                                                   {"id": "color", "value": {"mode": "fixed", "fixedColor": BLUE}}]}]},
        options={"showHeader": True, "cellHeight": "md", "sortBy": [{"displayName": "Start time", "desc": False}]},
        transformations=[{"id": "organize", "options": {
            "excludeByName": {"traceIdHidden": True, "spanID": True, "Trace ID": True, "Span ID": True, "traceService": True,
                              "traceName": True, "nested": True},
            "renameByName": select}}],
    )


def dashboard(uid, title, desc, panels, variables=None, tr="now-1h"):
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
        "style": "dark",
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


# ---- 1. Overview -----------------------------------------------------------------------------------------------------

def overview():
    n_def, n_clean = len(DEFECT), len(CLEAN)
    findings = "sum by (instance) (last_over_time(proven_review_findings_total%s))" % RANGE
    table = panel(
        5, "table", "Every review: how it ended, findings kept, claims dropped",
        "One row per review. A review that kept no finding is clean; one that kept a finding is blocked. "
        "The change column names the demo change the review ran on.",
        0, 5, 16, 10,
        [target('max by (instance) (last_over_time(proven_agent_state{proven_agent="writer"}%s))' % RANGE, PROM, ref="A", format="table"),
         target(findings, PROM, ref="B", format="table"),
         target("sum by (instance) (last_over_time(proven_review_claims_dropped_total%s))" % RANGE, PROM, ref="C", format="table")],
        fieldConfig={"defaults": {"noValue": "0", "custom": {"align": "left"}},
                     "overrides": [
                         {"matcher": {"id": "byName", "options": "Findings"}, "properties": [
                             {"id": "mappings", "value": [{"type": "value", "options": {"0": {"text": "clean", "color": GREEN}}},
                                                           {"type": "range", "options": {"from": 1, "to": 1000, "result": {"color": RED}}}]},
                             {"id": "custom.cellOptions", "value": {"type": "color-text"}}]},
                         {"matcher": {"id": "byName", "options": "State"}, "properties": [
                             {"id": "mappings", "value": [
                                 {"type": "value", "options": {"0": {"text": "waiting", "color": YELLOW}, "1": {"text": "working", "color": BLUE},
                                                               "2": {"text": "finished", "color": GREEN}, "3": {"text": "failed", "color": RED}}}]},
                             {"id": "custom.cellOptions", "value": {"type": "color-text"}}]}]},
        options={"showHeader": True, "cellHeight": "md", "sortBy": [{"displayName": "Review", "desc": True}]},
        transformations=[
            {"id": "joinByField", "options": {"byField": "instance", "mode": "outer"}},
            {"id": "extractFields", "options": {"source": "instance", "format": "regexp",
                                                "regExp": "(?<review>[0-9]+)-pr(?<pr>[0-9]+)-(?<change>.+)", "keepTime": False,
                                                "replace": False}},
            {"id": "organize", "options": {
                "excludeByName": {"Time": True, "Time 1": True, "Time 2": True, "Time 3": True, "instance": True, "review": True},
                "indexByName": {"pr": 0, "change": 1, "Value #A": 2, "Value #B": 3, "Value #C": 4},
                "renameByName": {"pr": "Pull request", "change": "Change", "Value #A": "State", "Value #B": "Findings",
                                 "Value #C": "Claims dropped"}}},
        ],
    )
    return dashboard("proven-overview", "Proven: every review at a glance",
                     "Every review the lab ran, whether Proven flagged the broken changes, and whether it raised a false alarm on the clean ones.", [
        stat(1, "Reviews finished", "Reviews whose last agent, the Writer, is done.",
             'count(last_over_time(proven_agent_state{proven_agent="writer"}%s) == 2) or vector(0)' % RANGE,
             0, 0, 6, 5, "suffix: reviews finished", [(None, YELLOW), (1, GREEN)], no_value="0 reviews finished"),
        stat(2, "Broken changes Proven flagged",
             "Demo pull requests with a real defect (%d in the lab) on which Proven kept at least one finding. All of them is good." % n_def,
             "count(sum by (instance) (last_over_time(proven_review_findings_total{instance=~\"%s\"}%s)) > 0) or vector(0)" % (DEFECT_RE, RANGE),
             6, 0, 6, 5, "suffix: of %d flagged" % n_def, [(None, RED), (1, YELLOW), (n_def, GREEN)], no_value="0 of %d flagged" % n_def),
        stat(3, "False alarms on the clean changes",
             "Findings Proven kept on the %d clean demo pull requests. None is good: a clean change must get a clean verdict." % n_clean,
             "sum(last_over_time(proven_review_findings_total{instance=~\"%s\"}%s)) or vector(0)" % (CLEAN_RE, RANGE),
             12, 0, 6, 5, "suffix: false alarms", [(None, GREEN), (1, RED)], no_value="0 false alarms"),
        stat(4, "Agents that failed",
             "Agent stages that ended failed in any review. None is good: a failed agent means part of the review did not run.",
             "count(last_over_time(proven_agent_state%s) == 3) or vector(0)" % RANGE,
             18, 0, 6, 5, "suffix: agents failed", [(None, GREEN), (1, RED)], no_value="0 agents failed"),
        table,
        bars(6, "Findings by level of proof", "How well each kept finding is proved: the strongest evidence it carries.",
             [target("sum by (proven_proof_level) (last_over_time(proven_review_findings_total%s))" % RANGE, PROM, legend="{{proven_proof_level}}")],
             16, 5, 8, 10, "No finding yet: the reviews are still running, or every change was clean.", color=GREEN),
        series(7, "Reviews started over time", "Reviews that began in each ten-minute step.",
               [target('count(count by (instance) (proven_agent_state{proven_agent="writer",proven_reviewer="shared"}))', PROM, legend="reviews", instant=False,
                       interval="10m")],
               0, 15, 12, 7, "No review has started in this time range."),
        spans_table(8, "Reviews that had a failed service", "A review whose outside service (the guardrail, a model, Zetaris) did not answer. None is good.",
                    '{ name = "proven.review" && span.proven.failed_services > 0 } | select(span.proven.failed_services, span.proven.class)',
                    12, 15, 12, 7, {"Start time": "Start time", "proven.failed_services": "Failed services", "proven.class": "Review class"},
                    "No review had a failed service."),
    ])


# ---- 2. One review ---------------------------------------------------------------------------------------------------

def one_review():
    var = {"name": "run", "label": "Review", "type": "query", "datasource": PROM,
           "query": {"query": "label_values(proven_agent_state, instance)", "refId": "run"},
           "definition": "label_values(proven_agent_state, instance)", "refresh": 2, "sort": 2, "includeAll": False, "multi": False,
           "current": {}}
    matrix = panel(
        5, "table", "The Skeptic's verdicts, both reviewers side by side",
        "How each reviewer's Skeptic judged each finding: refuted, upheld, or not attempted. Where the two columns differ, the reviewers disagree.",
        14, 5, 10, 10,
        [target('sum by (proven_reviewer, verdict) (last_over_time(proven_reviewer_verdicts_total{instance="$run"}%s))' % RANGE, PROM, format="table")],
        fieldConfig={"defaults": {"noValue": "No verdict: this review kept no finding for the Skeptic to judge.", "custom": {"align": "left"}}, "overrides": []},
        options={"showHeader": True, "cellHeight": "md"},
        transformations=[
            {"id": "groupingToMatrix", "options": {"columnField": "proven_reviewer", "rowField": "verdict", "valueField": "Value", "emptyValue": "zero"}},
            {"id": "organize", "options": {"renameByName": {"verdict\\proven_reviewer": "Verdict"}}}],
    )
    return dashboard("proven-review", "Proven: one review, step by step",
                     "Pick a review at the top. See each agent in order, how long it took, and how the two reviewers judged each finding.", [
        stat(1, "Findings kept", "Findings this review kept. None means a clean verdict.",
             'sum(last_over_time(proven_review_findings_total{instance="$run"}%s)) or vector(0)' % RANGE,
             0, 0, 8, 5, "suffix: findings kept", [(None, GREEN), (1, RED)], no_value="0 findings kept: clean"),
        stat(2, "Claims the Judge dropped", "Claims dropped for being refuted, unproved, or already refuted in an earlier review.",
             'sum(last_over_time(proven_review_claims_dropped_total{instance="$run"}%s)) or vector(0)' % RANGE,
             8, 0, 8, 5, "suffix: claims dropped", [(None, BLUE)], no_value="0 claims dropped"),
        stat(3, "Model calls that failed",
             "Calls to a model that returned an error. In sample mode a call the recorded set has no answer for counts here too; the review goes on.",
             'sum(last_over_time(proven_model_calls_total{instance="$run",outcome="error"}%s)) or vector(0)' % RANGE,
             16, 0, 8, 5, "suffix: calls failed", [(None, GREEN), (1, YELLOW)], no_value="0 calls failed"),
        spans_table(4, "The agents, in the order they ran, and how long each took",
                    "Each row is one agent of this review, from the trace. The bar is how long it ran. The two reviewers each run their own Planner and Skeptic.",
                    '{ resource.service.instance.id = "$run" && span.proven.agent != "" } | select(span.proven.agent, span.proven.reviewer, span.gen_ai.request.model)',
                    0, 5, 14, 10,
                    {"Start time": "Start time", "proven.agent": "Agent", "proven.reviewer": "Reviewer", "gen_ai.request.model": "Model", "Duration": "Duration"},
                    "No agent spans for this review yet."),
        matrix,
        bars(6, "Findings by level of proof", "The strongest evidence each finding of this review carries.",
             [target('sum by (proven_proof_level) (last_over_time(proven_review_findings_total{instance="$run"}%s))' % RANGE, PROM, legend="{{proven_proof_level}}")],
             0, 15, 8, 7, "No finding in this review.", color=GREEN),
        bars(7, "Where the two reviewers disagree", "Findings one reviewer named and the other did not, or that they judged differently.",
             [target('sum by (kind) (last_over_time(proven_reviewer_disagreements_total{instance="$run"}%s))' % RANGE, PROM, legend="{{kind}}")],
             8, 15, 8, 7, "The reviewers do not disagree on anything in this review.", color=YELLOW),
        text(8, "Each finding, with its evidence", "Where to read the evidence behind a finding.",
             "The evidence of each finding (the query that found it, or the test that fails on the change and passes before it) "
             "is in the **Report** page of its pull request. Open the lab's start page and use the **Report** link in the table.",
             16, 15, 8, 7),
    ], variables=[var])


# ---- 3. Evidence -----------------------------------------------------------------------------------------------------

def evidence():
    return dashboard("proven-evidence", "Proven: the evidence behind each finding",
                     "Where findings get their proof. What Proven does not yet report as a number is named in the two grey panels.", [
        text(1, "Zetaris queries: not measured yet",
             "What would be shown here, and the numbers that are missing.",
             "**Not emitted by Proven yet.** Proven asks Zetaris for each advisory check, but does not yet count it. "
             "Missing, so nothing is drawn here:\n\n- queries per table (OSV, KEV, EPSS)\n- rows returned\n- query duration\n- queries Zetaris refused\n\n"
             "The proof level of each finding, which comes from these queries, is shown below.", 0, 0, 12, 6),
        text(2, "Prover runs: not measured yet",
             "What would be shown here, and the numbers that are missing.",
             "**Not emitted by Proven yet.** The Prover runs a test on the base and on the change, with the network off. "
             "Missing, so nothing is drawn here:\n\n- runs, with the outcome on the base and on the change\n- run duration\n- runs that had the network off\n\n"
             "How long the Prover and the Investigator ran is in the table below.", 12, 0, 12, 6),
        stat(3, "Agents of evidence that failed",
             "Investigator or Prover stages that ended failed. None is good: a failed one means evidence was not gathered.",
             'count(last_over_time(proven_agent_state{proven_agent=~"investigator|prover"}%s) == 3) or vector(0)' % RANGE,
             0, 6, 6, 5, "suffix: failed", [(None, GREEN), (1, RED)], no_value="0 failed"),
        stat(4, "Claims dropped for no evidence",
             "Claims the Judge dropped because no query or test supported them. This is Proven refusing to guess.",
             'sum(last_over_time(proven_review_claims_dropped_total{reason="no_evidence"}%s)) or vector(0)' % RANGE,
             6, 6, 6, 5, "suffix: claims dropped", [(None, BLUE)], no_value="0 claims dropped"),
        bars(5, "Findings by level of proof", "How strong the proof of each kept finding is, from a failing test down to a query result.",
             [target("sum by (proven_proof_level) (last_over_time(proven_review_findings_total%s))" % RANGE, PROM, legend="{{proven_proof_level}}")],
             12, 6, 12, 5, "No finding yet: the reviews are still running, or every change was clean.", color=GREEN),
        spans_table(6, "How long the Investigator and the Prover ran, in each review",
                    "Each row is one run of the Investigator (turns a risk into a query) or the Prover (asks for a test). The bar is how long it ran.",
                    '{ span.proven.agent = "investigator" || span.proven.agent = "prover" } | select(span.proven.agent, span.proven.reviewer)',
                    0, 11, 24, 10, {"Start time": "Start time", "proven.agent": "Agent", "proven.reviewer": "Reviewer", "Duration": "Duration"},
                    "No Investigator or Prover has run yet."),
    ])


# ---- 4. Models and trust ---------------------------------------------------------------------------------------------

def models():
    latency = bars(
        8, "Model latency, mean per call (sample mode: recorded answers, not real model latency)",
        "Mean time of a model call. In sample mode the answers are recorded, so this is not a measure of any model; "
        "in a live lab it is the real latency of each agent's calls.",
        [target("sum by (proven_agent, proven_reviewer) (last_over_time(gen_ai_client_operation_duration_seconds_sum%s)) / "
                "sum by (proven_agent, proven_reviewer) (last_over_time(gen_ai_client_operation_duration_seconds_count%s))" % (RANGE, RANGE),
                PROM, legend="{{proven_agent}} · {{proven_reviewer}}")],
        16, 13, 8, 8, "No model call yet.", unit="s", color=BLUE)
    return dashboard("proven-models", "Proven: which models answered, and what the guardrail checked",
                     "Calls to models by agent and reviewer, which model answered, what the guardrail checked, and what Proven refused to keep.", [
        stat(1, "Model calls", "Calls to a model across all reviews.",
             "sum(last_over_time(proven_model_calls_total%s)) or vector(0)" % RANGE,
             0, 0, 8, 5, "suffix: model calls", [(None, BLUE)], no_value="0 model calls"),
        stat(2, "Model calls that returned an error",
             "A model call that returned an error. In sample mode a call the recorded answers have no answer for is counted here "
             "(generator: no recorded response); the review goes on without it. In a live lab an error is a failed call to the model.",
             'sum(last_over_time(proven_model_calls_total{outcome="error"}%s)) or vector(0)' % RANGE,
             8, 0, 8, 5, "suffix: calls with an error", [(None, GREEN), (1, YELLOW)], no_value="0 calls with an error"),
        stat(3, "Guardrail checks that said unsafe",
             "Model answers or prompts the guardrail judged unsafe. None is good.",
             "sum(last_over_time(proven_guardrail_unsafe_total%s)) or vector(0)" % RANGE,
             16, 0, 8, 5, "suffix: unsafe", [(None, GREEN), (1, RED)], no_value="0 unsafe"),
        bars(4, "Calls, by agent and reviewer", "Which agent made the calls, and for which reviewer. Planner and Skeptic run once per reviewer.",
             [target("sum by (proven_agent, proven_reviewer) (last_over_time(proven_model_calls_total%s))" % RANGE, PROM, legend="{{proven_agent}} · {{proven_reviewer}}")],
             0, 5, 12, 8, "No model call yet."),
        bars(5, "Calls, by the model asked and the model that answered",
             "The model a call asked for, and the model the provider says answered it. When a fallback answers, the two differ "
             "(NVIDIA first, then the fallback model). Sample mode names no answering model.",
             [target("sum by (gen_ai_request_model, gen_ai_response_model) (last_over_time(gen_ai_client_operation_duration_seconds_count%s))" % RANGE, PROM,
                     legend="{{gen_ai_request_model}} → {{gen_ai_response_model}}")],
             12, 5, 12, 8, "No model call yet."),
        bars(6, "Guardrail checks, by side and verdict", "What the guardrail was asked to check (the prompt or the answer) and what it said.",
             [target("sum by (side, verdict) (last_over_time(proven_guardrail_checks_total%s))" % RANGE, PROM, legend="{{side}} · {{verdict}}")],
             0, 13, 8, 8, "No guardrail check yet.", color=GREEN),
        bars(7, "What Proven refused to keep, by reason",
             "Claims the Judge dropped: refuted by the Skeptic, no evidence, or already refuted in an earlier review (suppressed).",
             [target("sum by (reason) (last_over_time(proven_review_claims_dropped_total%s))" % RANGE, PROM, legend="{{reason}}")],
             8, 13, 8, 8, "Nothing was dropped yet.", color=YELLOW),
        latency,
    ])


# ---- 5. Meterless engines --------------------------------------------------------------------------------------------

def engines():
    return dashboard("proven-engines", "Proven: the four Meterless engines and what each did",
                     "H-MEM, Scout Intent, Markovian and the World Model are on in this lab. What each one did is shown where Proven reports it.", [
        stat(1, "H-MEM: claims it suppressed", "Claims H-MEM recognised as already refuted in an earlier review, so they never reached the Judge.",
             'sum(last_over_time(proven_review_claims_dropped_total{reason="suppressed"}%s)) or vector(0)' % RANGE,
             0, 0, 6, 6, "suffix: claims suppressed", [(None, BLUE)], no_value="0 claims suppressed"),
        text(2, "Scout Intent: not measured yet", "What would be shown here.",
             "**Not emitted by Proven yet.** Scout plans each review, light or full, and can be overridden by a guard. Missing:\n\n"
             "- plans, by light or full\n- guard overrides\n\nWhat each review ran as is in the table below.", 6, 0, 6, 6),
        text(3, "Markovian: not measured yet", "What would be shown here.",
             "**Not emitted by Proven yet.** Markovian reviews a large change in bounded steps. Missing:\n\n- steps taken\n\n"
             "Small changes, like the demo's, take one step.", 12, 0, 6, 6),
        text(4, "World Model: not measured yet", "What would be shown here.",
             "**Not emitted by Proven yet.** The World Model keeps claims, evidence and verdicts as a graph. Missing:\n\n"
             "- claims\n- evidence\n- verdicts recorded", 18, 0, 6, 6),
        spans_table(5, "What each review ran as", "One row per review: the review class Scout chose, whether a second reviewer ran, and whether the review was cut short.",
                    '{ name = "proven.review" } | select(span.proven.class, span.proven.doubled, span.proven.partial, span.proven.findings)',
                    0, 6, 24, 10,
                    {"Start time": "Start time", "proven.class": "Review class", "proven.doubled": "Two reviewers", "proven.partial": "Cut short",
                     "proven.findings": "Findings kept", "Duration": "Duration"},
                    "No review has run yet."),
    ])


# ---- 6. Nightshift ---------------------------------------------------------------------------------------------------

NS = '{job="nightshift"}'
NS_EMPTY = "No receipts loaded. The lab loads the public receipts of Proven's own calls from lab/nightshift/receipts.public.jsonl (make nightshift-receipts)."


def nightshift():
    unwrap = lambda f: '| json | unwrap %s | __error__=""' % f
    return dashboard("proven-nightshift", "Proven: the model gateway, from Nightshift's signed receipts",
                     "Every call Proven made through Meterless Nightshift: how many, to which model, how fast, and which failed and why. "
                     "The receipts are replayed when the lab starts, so time here is when they were loaded.", [
        stat(1, "Receipts loaded", "Public receipts of Proven's own calls, loaded into this lab.",
             "sum(count_over_time(%s%s)) or vector(0)" % (NS, RANGE), 0, 0, 6, 5, "suffix: receipts", [(None, YELLOW), (1, BLUE)],
             no_value="0 receipts: none loaded", ds=LOKI),
        stat(2, "Calls that answered", "Calls whose receipt says ok.",
             'sum(count_over_time({job="nightshift", status="ok"}%s)) or vector(0)' % RANGE, 6, 0, 6, 5, "suffix: ok", [(None, YELLOW), (1, GREEN)],
             no_value="0 ok", ds=LOKI),
        stat(3, "Calls that failed", "Calls whose receipt says failed. The reasons are in the panel below. None is good.",
             'sum(count_over_time({job="nightshift", status="failed"}%s)) or vector(0)' % RANGE, 12, 0, 6, 5, "suffix: failed",
             [(None, GREEN), (1, RED)], no_value="0 failed", ds=LOKI),
        stat(4, "Receipts whose signature verified", "Receipts whose Ed25519 signature verified under the pinned key. All of them is good.",
             'sum(count_over_time({job="nightshift", verified="true"}%s)) or vector(0)' % RANGE, 18, 0, 6, 5, "suffix: verified",
             [(None, RED), (1, GREEN)], no_value="0 verified", ds=LOKI),
        series(5, "Calls over time, by model and app", "Calls per step, for each model that answered and each app that asked.",
               [target("sum by (model, app) (count_over_time(%s[$__interval]))" % NS, LOKI, legend="{{model}} · {{app}}", instant=False)],
               0, 5, 12, 8, NS_EMPTY),
        bars(6, "Why calls failed", "The reason on each failed receipt: the provider's deadline, an empty answer, a refused key, and the rest.",
             [target('sum by (reason) (count_over_time({job="nightshift", status="failed"}%s))' % RANGE, LOKI, legend="{{reason}}")],
             12, 5, 12, 8, "No call failed.", color=RED),
        series(7, "Latency per call, median, by model", "Median time of an answered call, per model.",
               [target("quantile_over_time(0.5, %s %s [$__interval]) by (model)" % (NS.replace('}', ', status="ok"}'), unwrap("latency_ms")), LOKI,
                       legend="{{model}}", instant=False)],
               0, 13, 12, 8, NS_EMPTY, unit="ms", style="line"),
        bars(8, "Tokens, by model", "Input and output tokens of the calls, per model.",
             [target("sum by (model) (sum_over_time(%s %s [$__range]))" % (NS, unwrap("input_tokens")), LOKI, legend="{{model}} in", ref="A"),
              target("sum by (model) (sum_over_time(%s %s [$__range]))" % (NS, unwrap("output_tokens")), LOKI, legend="{{model}} out", ref="B")],
             12, 13, 12, 8, NS_EMPTY),
    ], tr="now-6h")


def main():
    for build in (overview, one_review, evidence, models, engines, nightshift):
        write(build())


if __name__ == "__main__":
    main()
