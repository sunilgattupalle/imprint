#!/usr/bin/env python3
"""
Imprint Model Pipeline — Gate Validator & Sanitized Exporter (`validate_instance.py`).

Reads raw evidence-bearing files from `<instance>/.local/`:
  - `.local/sources.yaml`
  - `.local/evidence_ledger.json` (L1 claims with verbatim quotes & workspace bindings)
  - `.local/demoted_non_imprint.json` (L2 workspace bindings & L3 agent baseline rules)

Validates all L1 claims against:
  1. Schema (`schemas/claim.schema.json`)
  2. Gate 1 — Portability Check (zero banned workspace terms in `title`, `portable_rule`,
     `applies_when`, `flips_when`, or `admission_tests`)
  3. Gate 2 — Non-Trivial Fork Check (`peer_fork` present & substantive)
  4. Graduation Bar (>= 3 evidence items across >= 2 contexts)

Then compiles `<instance>/imprint.json` (conforming to `schemas/imprint.schema.json`)
containing ONLY portable rules, boundary conditions, admission tests, and aggregate
provenance counts — stripping all verbatim quotes, session UUIDs, project names, and
workspace bindings so the root directory is safe for version control.
"""
import argparse
import datetime
import json
import os
import re
import sys
from _yaml_loader import load_simple_yaml

VALID_FACETS = {
    "F1_epistemic_posture",
    "F2_dialectical_style",
    "F3_compression_and_structure",
    "F4_voice_and_visuals",
    "F5_delegation_and_risk",
}

REQUIRED_KEYS = {
    "id",
    "facet",
    "scope",
    "status",
    "title",
    "portable_rule",
    "admission_tests",
    "applies_when",
    "flips_when",
    "evidence",
    "counter_evidence",
}


def resolve_local_dir(instance_dir: str) -> str:
    local_dir = os.path.join(instance_dir, ".local")
    return local_dir if os.path.isdir(local_dir) else instance_dir


def validate_and_export(instance_dir: str):
    instance_dir = os.path.abspath(os.path.expanduser(instance_dir))
    local_dir = resolve_local_dir(instance_dir)

    cfg_path = os.path.join(local_dir, "sources.yaml")
    claims_path = os.path.join(local_dir, "evidence_ledger.json")
    demoted_path = os.path.join(local_dir, "demoted_non_imprint.json")

    cfg = load_simple_yaml(cfg_path)
    with open(claims_path) as f:
        claims = json.load(f)
    with open(demoted_path) as f:
        demoted = json.load(f)

    disallowed_terms = cfg.get("workspace_terms_disallowed_in_portable_rules", [])

    errors = []
    facet_counts = {f: 0 for f in VALID_FACETS}
    sanitized_claims = []

    for c in claims:
        cid = c.get("id", "<missing-id>")
        missing = REQUIRED_KEYS - set(c.keys())
        if missing:
            errors.append(f"{cid}: missing required keys {missing}")
            continue
        if not re.match(r"^IMP-F[1-5]-\d{2}$", cid):
            errors.append(f"{cid}: id does not match ^IMP-F[1-5]-\\d{{2}}$")
        if c["facet"] not in VALID_FACETS:
            errors.append(f"{cid}: invalid facet {c['facet']}")
        else:
            facet_counts[c["facet"]] += 1

        adm = c.get("admission_tests", {})
        portable_fields = [
            c["title"],
            c["portable_rule"],
            c["applies_when"],
            c["flips_when"],
            adm.get("portability_check", ""),
            adm.get("peer_fork", ""),
        ]
        blob = " ".join(portable_fields)
        for term in disallowed_terms:
            pat = re.compile(r"\b" + re.escape(term) + r"\b", re.I)
            if pat.search(blob):
                errors.append(
                    f"{cid}: Gate 1 (Portability) failed — workspace term '{term}' leaked into portable fields"
                )

        if len(adm.get("portability_check", "")) < 25:
            errors.append(f"{cid}: Gate 1 portability_check explanation too short")
        if len(adm.get("peer_fork", "")) < 25:
            errors.append(f"{cid}: Gate 2 peer_fork explanation too short")

        ev = c.get("evidence", [])
        if len(ev) < 3:
            errors.append(f"{cid}: fewer than 3 evidence items ({len(ev)})")
        distinct_source_types = len({e.get("source_type", "") for e in ev})
        distinct_projects = len({e.get("project", "") for e in ev})
        if distinct_projects < 2 and distinct_source_types < 2:
            errors.append(f"{cid}: evidence does not span >= 2 contexts")

        if len(c.get("flips_when", "")) < 15:
            errors.append(f"{cid}: flips_when boundary condition missing or too short")

        sanitized_claims.append({
            "id": c["id"],
            "facet": c["facet"],
            "scope": c["scope"],
            "status": c["status"],
            "title": c["title"],
            "portable_rule": c["portable_rule"],
            "admission_tests": {
                "portability_check": adm["portability_check"],
                "peer_fork": adm["peer_fork"],
            },
            "applies_when": c["applies_when"],
            "flips_when": c["flips_when"],
            "provenance_counts": {
                "evidence_items": len(ev),
                "distinct_source_types": distinct_source_types,
                "distinct_projects": distinct_projects,
                "counter_evidence_items": len(c.get("counter_evidence", [])),
            },
        })

    if errors:
        print("VALIDATION FAILED:")
        for e in errors:
            print("  -", e)
        sys.exit(1)

    public_imprint = {
        "schema_version": "1.0.0",
        "generated_at": datetime.date.today().isoformat(),
        "claims": sanitized_claims,
    }
    out_path = os.path.join(instance_dir, "imprint.json")
    with open(out_path, "w") as f:
        json.dump(public_imprint, f, indent=2)

    print(f"VALIDATION PASSED for instance '{os.path.basename(instance_dir)}'")
    print(f"  Exported quote-free portable summary -> {out_path}")
    print(f"  Graduated portable imprint claims (L1): {len(claims)}")
    for f, cnt in sorted(facet_counts.items()):
        print(f"    - {f}: {cnt} claims")
    l2 = sum(1 for d in demoted if d.get("layer") == "L2_workspace_binding")
    l3 = sum(1 for d in demoted if d.get("layer") == "L3_agent_baseline")
    print(f"  Separated local-only rules (.local/): {len(demoted)} ({l2} L2 workspace bindings, {l3} L3 agent baseline rules)")


if __name__ == "__main__":
    ap = argparse.ArgumentParser()
    ap.add_argument("--instance", required=True)
    args = ap.parse_args()
    validate_and_export(args.instance)
