#!/usr/bin/env python3
"""
Minimal zero-dependency YAML subset / JSON loader for Imprint Model `sources.yaml`
so `imprint-model` runs on stock Python 3 without requiring `pip install pyyaml`.
"""
import json
import re


def load_simple_yaml(path: str) -> dict:
    with open(path) as f:
        text = f.read()
    # Try JSON first (valid YAML 1.2 subset)
    try:
        return json.loads(text)
    except Exception:
        pass
    try:
        import yaml  # type: ignore
        return yaml.safe_load(text)
    except Exception:
        pass

    # Fallback parser for the flat/2-level structure used in sources.yaml
    result = {}
    section = None
    subsection = None
    for raw_line in text.splitlines():
        line = raw_line.split("#", 1)[0].rstrip()
        if not line.strip():
            continue
        indent = len(line) - len(line.lstrip(" "))
        stripped = line.strip()
        if indent == 0 and stripped.endswith(":"):
            section = stripped[:-1].strip()
            result[section] = {}
            subsection = None
        elif indent == 2 and stripped.startswith("- "):
            if isinstance(result.get(section), dict) and not result[section]:
                result[section] = []
            val = stripped[2:].strip().strip('"').strip("'")
            if isinstance(result.get(section), list):
                result[section].append(val)
        elif indent == 2 and ":" in stripped:
            k, v = stripped.split(":", 1)
            k = k.strip()
            v = v.strip().strip('"').strip("'")
            if v == "":
                result[section][k] = []
                subsection = k
            else:
                try:
                    v_parsed = float(v) if "." in v else int(v)
                except ValueError:
                    v_parsed = v
                result[section][k] = v_parsed
                subsection = k
        elif indent >= 4 and stripped.startswith("- "):
            item = stripped[2:].strip()
            if ":" in item and not item.startswith(('"', "'")):
                ik, iv = item.split(":", 1)
                result[section][subsection].append({ik.strip(): iv.strip().strip('"').strip("'")})
            else:
                result[section][subsection].append(item.strip('"').strip("'"))
        elif indent >= 4 and ":" in stripped and isinstance(result.get(section, {}).get(subsection), list):
            ik, iv = stripped.split(":", 1)
            result[section][subsection][-1][ik.strip()] = iv.strip().strip('"').strip("'")
    return result
