#!/usr/bin/env python3
"""
Imprint Model Pipeline — Step 1 & 2: Corpus Hygiene, Cross-Store Deduplication,
Authorship Purification, and 80/20 Holdout Split.

Usage:
  python3 pipeline/build_corpus.py --instance <instance_dir>
"""
import argparse
import collections
import glob
import hashlib
import json
import os
import re
import sqlite3
from _yaml_loader import load_simple_yaml

DEFAULT_REJECT_PATTERNS = [
    ("worker_tag", re.compile(r"^\s*<worker>")),
    ("teammate_msg", re.compile(r"^\s*<teammate-message\b")),
    ("local_command", re.compile(r"^\s*<(?:local-command-caveat|command-name|command-message)>")),
    ("task_notification", re.compile(r"^\s*<task-notification>")),
    ("system_reminder", re.compile(r"^\s*<system-reminder>")),
    ("subagent_header", re.compile(r"^\s*\[Subagent Context\]")),
    ("slash_command", re.compile(r"^\s*/(?:login|logout|clear|compact|mcp|model|help|status|resume|init|cost|doctor|review|pr_comments)\b")),
    ("ultraplan_scaffold", re.compile(r"^\s*Use ultraplan\b", re.I)),
]

DEFAULT_INLINE_STRIP_PATTERNS = [
    re.compile(r"<system-reminder>.*?</system-reminder>", re.S),
    re.compile(r"<local-command-caveat>.*?</local-command-caveat>", re.S),
]


def extract_raw_uuid(session_id: str) -> str:
    return session_id.split(":", 1)[1] if ":" in session_id else session_id


def clean_user_text(raw: str, reject_patterns, inline_strip_patterns):
    if not raw or not raw.strip():
        return None, "empty"
    for name, pat in reject_patterns:
        if pat.search(raw):
            return None, name
    cleaned = raw
    for pat in inline_strip_patterns:
        cleaned = pat.sub("", cleaned)
    cleaned = cleaned.strip()
    if not cleaned:
        return None, "empty_after_strip"
    lines = [ln.strip() for ln in cleaned.splitlines() if ln.strip()]
    if len(lines) >= 6:
        stack_lines = sum(
            1 for ln in lines
            if ln.startswith(("at ", "File \"", "Traceback (most recent", "Caused by:", "E   ", "STDERR:"))
            or re.match(r"^\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}", ln)
        )
        if stack_lines / len(lines) > 0.75:
            return None, "pure_log_or_stacktrace"
    return cleaned, None


def is_holdout(raw_uuid: str, ratio: float = 0.20) -> bool:
    h = int(hashlib.sha256(raw_uuid.encode("utf-8")).hexdigest()[:8], 16)
    return (h % 1000) < int(ratio * 1000)


def resolve_local_dir(instance_dir: str) -> str:
    local_dir = os.path.join(instance_dir, ".local")
    return local_dir if os.path.isdir(local_dir) else instance_dir


def run(instance_dir: str):
    instance_dir = os.path.abspath(os.path.expanduser(instance_dir))
    local_dir = resolve_local_dir(instance_dir)
    cfg_path = os.path.join(local_dir, "sources.yaml")
    cfg = load_simple_yaml(cfg_path)

    hygiene_cfg = cfg.get("hygiene", {})
    reject_patterns = list(DEFAULT_REJECT_PATTERNS)
    for item in hygiene_cfg.get("extra_reject_patterns", []):
        reject_patterns.append((item["name"], re.compile(item["regex"])))

    inline_strip_patterns = list(DEFAULT_INLINE_STRIP_PATTERNS)
    for rx in hygiene_cfg.get("extra_inline_strip_regexes", []):
        inline_strip_patterns.append(re.compile(rx, re.S))

    db_path = os.path.expanduser(cfg["session_archive"]["sqlite_path"])
    out_dir = os.path.join(local_dir, "corpus")
    os.makedirs(out_dir, exist_ok=True)

    conn = sqlite3.connect(f"file:{db_path}?mode=ro", uri=True)
    conn.row_factory = sqlite3.Row
    cur = conn.cursor()

    label_overrides = {}
    for row in cur.execute("SELECT session_id, label FROM session_labels WHERE label LIKE 'source=%'"):
        label_overrides[row[0]] = row[1].split("=", 1)[1]

    sessions = [
        dict(r) for r in cur.execute(
            "SELECT id, project, machine, agent, first_message, started_at, ended_at, "
            "message_count, user_message_count, parent_session_id, is_automated, cwd, git_branch "
            "FROM sessions WHERE COALESCE(is_automated, 0) = 0"
        )
    ]

    by_uuid = collections.defaultdict(list)
    for s in sessions:
        s["raw_uuid"] = extract_raw_uuid(s["id"])
        s["effective_agent"] = label_overrides.get(s["id"], s["agent"])
        by_uuid[s["raw_uuid"]].append(s)

    canonical_sessions = []
    dedup_dropped = 0
    cursor_meta_merged = 0
    for uuid, group in by_uuid.items():
        if len(group) == 1:
            canonical_sessions.append(group[0])
            continue
        agents = {g["effective_agent"]: g for g in group}
        if "cursor-ide" in agents and "cursor" in agents:
            primary = agents["cursor-ide"]
            secondary = agents["cursor"]
            if (not primary["project"] or primary["project"] == "unassigned") and secondary["project"]:
                primary["project"] = secondary["project"]
                cursor_meta_merged += 1
            if not primary["cwd"] and secondary["cwd"]:
                primary["cwd"] = secondary["cwd"]
            if not primary["git_branch"] and secondary["git_branch"]:
                primary["git_branch"] = secondary["git_branch"]
            canonical_sessions.append(primary)
            dedup_dropped += len(group) - 1
        else:
            group.sort(key=lambda x: (x["message_count"] or 0), reverse=True)
            canonical_sessions.append(group[0])
            dedup_dropped += len(group) - 1

    stats = {
        "raw_sessions_in_db": len(sessions),
        "canonical_sessions_after_dedup": len(canonical_sessions),
        "duplicate_sessions_dropped": dedup_dropped,
        "cursor_sessions_enriched_with_project": cursor_meta_merged,
        "rejected_turns_by_reason": collections.Counter(),
        "kept_turns_by_agent": collections.Counter(),
        "sessions_by_agent": collections.Counter(),
        "train_sessions": 0,
        "holdout_sessions": 0,
        "train_turns": 0,
        "holdout_turns": 0,
    }

    train_f = open(os.path.join(out_dir, "train_turns.jsonl"), "w")
    holdout_f = open(os.path.join(out_dir, "holdout_turns.jsonl"), "w")
    holdout_ratio = hygiene_cfg.get("holdout_ratio", 0.20)

    for s in canonical_sessions:
        sid = s["id"]
        agent = s["effective_agent"]
        holdout = is_holdout(s["raw_uuid"], holdout_ratio)
        msgs = [
            dict(r) for r in cur.execute(
                "SELECT ordinal, role, content, timestamp, has_tool_use, content_length "
                "FROM messages WHERE session_id = ? ORDER BY ordinal ASC",
                (sid,),
            )
        ]
        kept_in_session = 0
        for i, m in enumerate(msgs):
            if m["role"] != "user":
                continue
            cleaned, reject_reason = clean_user_text(m["content"], reject_patterns, inline_strip_patterns)
            if reject_reason:
                stats["rejected_turns_by_reason"][reject_reason] += 1
                continue
            prev_assistant = None
            for j in range(i - 1, -1, -1):
                if msgs[j]["role"] == "assistant" and msgs[j]["content"] and msgs[j]["content"].strip():
                    prev_assistant = msgs[j]["content"].strip()[:600]
                    break
            next_assistant = None
            for j in range(i + 1, len(msgs)):
                if msgs[j]["role"] == "assistant" and msgs[j]["content"] and msgs[j]["content"].strip():
                    next_assistant = msgs[j]["content"].strip()[:600]
                    break
            record = {
                "session_id": sid,
                "raw_uuid": s["raw_uuid"],
                "agent": agent,
                "project": s["project"] or "unassigned",
                "cwd": s["cwd"] or "",
                "git_branch": s["git_branch"] or "",
                "is_subagent": bool(s["parent_session_id"]),
                "ordinal": m["ordinal"],
                "timestamp": m["timestamp"] or s["started_at"] or "",
                "user_text": cleaned,
                "user_chars": len(cleaned),
                "prev_assistant_snippet": prev_assistant,
                "next_assistant_snippet": next_assistant,
                "split": "holdout" if holdout else "train",
            }
            line = json.dumps(record, ensure_ascii=False) + "\n"
            if holdout:
                holdout_f.write(line)
                stats["holdout_turns"] += 1
            else:
                train_f.write(line)
                stats["train_turns"] += 1
            stats["kept_turns_by_agent"][agent] += 1
            kept_in_session += 1

        if kept_in_session > 0:
            stats["sessions_by_agent"][agent] += 1
            if holdout:
                stats["holdout_sessions"] += 1
            else:
                stats["train_sessions"] += 1

    default_export_project = cfg.get("chatgpt_exports", {}).get("default_project", "exported_chats")
    for g in cfg.get("chatgpt_exports", {}).get("globs", []):
        for p in sorted(glob.glob(os.path.expanduser(g))):
            with open(p) as f:
                data = json.load(f)
            sid = f"chatgpt:{os.path.splitext(os.path.basename(p))[0]}"
            holdout = is_holdout(sid, holdout_ratio)
            msgs = data.get("messages", [])
            kept = 0
            for idx, m in enumerate(msgs):
                if m.get("role") != "user":
                    continue
                cleaned, reject_reason = clean_user_text(m.get("text", ""), reject_patterns, inline_strip_patterns)
                if reject_reason:
                    stats["rejected_turns_by_reason"][reject_reason] += 1
                    continue
                prev_a = msgs[idx - 1]["text"][:600] if idx > 0 and msgs[idx - 1].get("role") == "assistant" else None
                next_a = msgs[idx + 1]["text"][:600] if idx + 1 < len(msgs) and msgs[idx + 1].get("role") == "assistant" else None
                record = {
                    "session_id": sid,
                    "raw_uuid": sid,
                    "agent": "chatgpt",
                    "project": default_export_project,
                    "cwd": os.path.dirname(p),
                    "git_branch": "",
                    "is_subagent": False,
                    "ordinal": idx,
                    "timestamp": data.get("updated_at", ""),
                    "user_text": cleaned,
                    "user_chars": len(cleaned),
                    "prev_assistant_snippet": prev_a,
                    "next_assistant_snippet": next_a,
                    "split": "holdout" if holdout else "train",
                }
                line = json.dumps(record, ensure_ascii=False) + "\n"
                if holdout:
                    holdout_f.write(line)
                    stats["holdout_turns"] += 1
                else:
                    train_f.write(line)
                    stats["train_turns"] += 1
                stats["kept_turns_by_agent"]["chatgpt"] += 1
                kept += 1
            if kept > 0:
                stats["sessions_by_agent"]["chatgpt"] += 1
                if holdout:
                    stats["holdout_sessions"] += 1
                else:
                    stats["train_sessions"] += 1

    train_f.close()
    holdout_f.close()
    stats["rejected_turns_by_reason"] = dict(stats["rejected_turns_by_reason"])
    stats["kept_turns_by_agent"] = dict(stats["kept_turns_by_agent"])
    stats["sessions_by_agent"] = dict(stats["sessions_by_agent"])
    with open(os.path.join(out_dir, "stats.json"), "w") as f:
        json.dump(stats, f, indent=2)
    print(json.dumps(stats, indent=2))


if __name__ == "__main__":
    ap = argparse.ArgumentParser()
    ap.add_argument("--instance", required=True, help="Path to person imprint instance directory")
    args = ap.parse_args()
    run(args.instance)
