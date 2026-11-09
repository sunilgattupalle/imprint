#!/usr/bin/env python3
"""
Imprint Model Pipeline — Step 5: Held-Out Evaluation (`evaluate_holdout.py`).

Evaluates an instance's portable imprint claims (`claims/imprint_claims.json`)
against the untouched 20% holdout split (`corpus/holdout_turns.jsonl`), and
separately tracks matches against demoted L2 (workspace) and L3 (agent baseline)
rules so the evaluation cleanly separates personal imprint coverage from
workspace/baseline noise.
"""
import argparse
import json
import os
import re

SIGNAL_RE = re.compile(
    r"("
    r"\bdon't just\b|\bdon't make it\b|\bdon't overwrite\b|\bdo not use\b|\bstop updating\b|\bwhat is wrong\b|\bwhat the hell\b|"
    r"\bconcise\b|\bprecise\b|\bno ai fluff\b|\bover-?complicat|\bover-?correct|\btoo concise\b|\btoo big\b|"
    r"\bmental model\b|\bprimitive\b|\bconstruct\b|\bfirst principles\b|\bsource of truth\b|\brubric\b|\bground truth\b|\bindependent doc\b|"
    r"\bhonest\b|\bstress test\b|\bwhat do you think\b|\bask opus\b|\bopus 5\b|\bam i conflating\b|\breview this thoroughly\b|"
    r"\bwrite -?v\d\b|\bevolution till v\d\b|\bseparate slide\b|"
    r"\buse svg\b|\bfix all the ascii\b|\buse ascii diagrams\b|\bvisually\b|\bvisual report\b|\btoo many colors\b|\btoo much text\b|"
    r"\bautonomy\b|\buse my identity\b|\bworkload identity\b|\btrust .*-issued identity\b|\bdispatches an event\b|\bcome back to resume\b|"
    r"\buse my writing style\b|\brephrase the whole thing\b|\bopening line\b"
    r")",
    re.I,
)


def resolve_local_dir(instance_dir: str) -> str:
    local_dir = os.path.join(instance_dir, ".local")
    return local_dir if os.path.isdir(local_dir) else instance_dir


def evaluate(instance_dir: str):
    instance_dir = os.path.abspath(os.path.expanduser(instance_dir))
    local_dir = resolve_local_dir(instance_dir)
    instance_name = os.path.basename(instance_dir)
    holdout_path = os.path.join(local_dir, "corpus/holdout_turns.jsonl")
    claims_path = os.path.join(local_dir, "evidence_ledger.json")
    demoted_path = os.path.join(local_dir, "demoted_non_imprint.json")
    eval_dir = os.path.join(local_dir, "eval")
    os.makedirs(eval_dir, exist_ok=True)

    with open(claims_path) as f:
        claims = json.load(f)
    with open(demoted_path) as f:
        demoted = json.load(f)

    seen = set()
    signal_turns = []
    for line in open(holdout_path):
        r = json.loads(line)
        txt = r["user_text"].strip()
        if len(txt) < 20 or len(txt) > 1200:
            continue
        if txt.startswith("{") or '"commits": [' in txt:
            continue
        if not SIGNAL_RE.search(txt):
            continue
        key = (r["session_id"], txt)
        if key in seen:
            continue
        seen.add(key)
        signal_turns.append(r)

    claim_regexes = [
        (c["id"], c["facet"], c["title"], re.compile(c["eval_matcher"], re.I))
        for c in claims
        if c.get("eval_matcher")
    ]
    demoted_regexes = [
        (d["id"], d["layer"], d["title"], re.compile(d["eval_matcher"], re.I))
        for d in demoted
        if d.get("eval_matcher")
    ]

    l1_hits = {c["id"]: [] for c in claims}
    l2_l3_hits = {d["id"]: [] for d in demoted}
    covered_l1 = 0
    covered_any = 0
    uncovered = []

    for r in signal_turns:
        txt = r["user_text"]
        matched_l1 = []
        for cid, facet, title, pat in claim_regexes:
            if pat.search(txt):
                matched_l1.append(cid)
                l1_hits[cid].append(r)
        matched_dem = []
        for did, layer, title, pat in demoted_regexes:
            if pat.search(txt):
                matched_dem.append(did)
                l2_l3_hits[did].append(r)

        if matched_l1:
            covered_l1 += 1
        if matched_l1 or matched_dem:
            covered_any += 1
        else:
            uncovered.append(r)

    total = len(signal_turns)
    l1_rate = covered_l1 / total if total else 0.0
    any_rate = covered_any / total if total else 0.0

    results = {
        "instance": instance_name,
        "holdout_signal_turns": total,
        "covered_by_L1_portable_imprint": covered_l1,
        "L1_coverage_rate": round(l1_rate, 3),
        "covered_by_L1_or_L2_L3": covered_any,
        "total_coverage_rate": round(any_rate, 3),
        "hits_per_L1_claim": {k: len(v) for k, v in l1_hits.items()},
        "hits_per_demoted_rule": {k: len(v) for k, v in l2_l3_hits.items()},
    }
    with open(os.path.join(eval_dir, "holdout_results.json"), "w") as f:
        json.dump(results, f, indent=2)

    md = [
        f"# Held-Out Evaluation Report (`{instance_name}/eval`)",
        "",
        f"- **High-signal directive/corrective turns in holdout (excluding JSON pastes):** {total}",
        f"- **Covered by L1 Portable Imprint (`imprint_claims.json`):** {covered_l1} / {total} (**{100*l1_rate:.1f}%**)",
        f"- **Covered including separated L2 Workspace & L3 Agent Baseline (`demoted_non_imprint.json`):** {covered_any} / {total} (**{100*any_rate:.1f}%**)",
        "",
        "## L1 Portable Imprint Claims — Holdout Verification",
        "",
        "| Claim ID | Facet | Portable Claim | Holdout Hits | Sample Held-Out Evidence |",
        "|---|---|---|---:|---|",
    ]
    for c in claims:
        cid = c["id"]
        hits = l1_hits[cid]
        sample = (
            f"`{hits[0]['timestamp'][:10]}` ({hits[0]['agent']}): *\"{hits[0]['user_text'][:105].replace(chr(10), ' ')}...\"*"
            if hits
            else "—"
        )
        md.append(f"| `{cid}` | `{c['facet']}` | {c['title']} | {len(hits)} | {sample} |")

    md.extend([
        "",
        "## Separated L2 Workspace & L3 Agent Baseline Rules",
        "",
        "| ID | Layer | Rule Summary | Why Excluded from L1 Imprint | Holdout Hits |",
        "|---|---|---|---|---:|",
    ])
    for d in demoted:
        did = d["id"]
        md.append(f"| `{did}` | `{d['layer']}` | {d['title']} | {d['reason_excluded_from_imprint']} | {len(l2_l3_hits[did])} |")

    with open(os.path.join(eval_dir, "holdout_report.md"), "w") as f:
        f.write("\n".join(md) + "\n")

    print(f"Holdout L1 portable coverage: {covered_l1}/{total} ({100*l1_rate:.1f}%)")
    print(f"Holdout L1+L2+L3 coverage:    {covered_any}/{total} ({100*any_rate:.1f}%)")


if __name__ == "__main__":
    ap = argparse.ArgumentParser()
    ap.add_argument("--instance", required=True)
    args = ap.parse_args()
    evaluate(args.instance)
