#!/usr/bin/env python3
"""
Compares Kanidm OpenAPI spec against api-coverage.yaml to report implementation coverage.
Outputs Markdown suitable for GitHub Actions Summary.
"""

import json
import os
import sys
import urllib.request
from pathlib import Path

import yaml

ROOT = Path(__file__).parent.parent
SPEC_FILE = ROOT / "internal/spec/kanidm-openapi.json"
COVERAGE_FILE = ROOT / "api-coverage.yaml"

SKIP_PREFIXES = ("/robots.txt", "/ui/", "/scim/", "/v1/debug/", "/v1/jwk/",
                 "/v1/logout", "/v1/auth", "/v1/reauth", "/v1/credential/",
                 "/v1/self", "/v1/schema", "/v1/raw/", "/v1/recycle_bin",
                 "/v1/sync_account", "/v1/system", "/v1/domain", "/v1/status")

def load_spec():
    with open(SPEC_FILE) as f:
        return json.load(f)

def load_coverage():
    with open(COVERAGE_FILE) as f:
        data = yaml.safe_load(f)
    return {(e["method"].upper(), e["path"]) for e in data["implemented"]}

def is_relevant(path):
    return not any(path.startswith(p) for p in SKIP_PREFIXES)

def main():
    spec = load_spec()
    implemented = load_coverage()
    kanidm_version = spec["info"]["version"]

    all_endpoints = []
    for path, methods in sorted(spec["paths"].items()):
        if not is_relevant(path):
            continue
        for method in methods:
            all_endpoints.append((method.upper(), path))

    total = len(all_endpoints)
    covered = sum(1 for ep in all_endpoints if ep in implemented)
    pct = covered / total * 100 if total else 0

    # stdout output (GitHub Summary)
    print(f"## API Coverage: {covered}/{total} endpoints ({pct:.1f}%)")
    print(f"> Kanidm v{kanidm_version} · OpenTofu Provider")
    print()

    categories = {}
    for method, path in all_endpoints:
        cat = path.split("/")[2] if path.startswith("/v1/") else "other"
        categories.setdefault(cat, []).append((method, path))

    for cat in sorted(categories):
        eps = categories[cat]
        cat_covered = sum(1 for ep in eps if ep in implemented)
        cat_total = len(eps)
        cat_pct = cat_covered / cat_total * 100 if cat_total else 0
        print()
        print(f"### {cat} ({cat_covered}/{cat_total} · {cat_pct:.0f}%)")
        print()
        print("| | Method | Endpoint |")
        print("|---|---|---|")
        for method, path in sorted(eps, key=lambda x: x[1]):
            status = "✅" if (method, path) in implemented else "❌"
            print(f"| {status} | `{method}` | `{path}` |")

    update_gist(pct)

def update_gist(pct: float) -> None:
    gist_id = os.environ.get("GIST_ID")
    token = os.environ.get("GIST_SECRET")
    if not gist_id or not token:
        return

    color = "brightgreen" if pct >= 80 else "green" if pct >= 60 else "yellow" if pct >= 40 else "orange"
    badge = {
        "schemaVersion": 1,
        "label": "API Coverage",
        "message": f"{pct:.1f}%",
        "color": color,
    }
    data = json.dumps({"files": {"badge.json": {"content": json.dumps(badge)}}}).encode()
    req = urllib.request.Request(
        f"https://api.github.com/gists/{gist_id}",
        data=data,
        method="PATCH",
        headers={"Authorization": f"Bearer {token}", "Content-Type": "application/json"},
    )
    urllib.request.urlopen(req)
    print(f"\n> Badge updated: {pct:.1f}%", file=sys.stderr)


if __name__ == "__main__":
    main()
