#!/usr/bin/env python3
"""Prepare an AI review bundle from the current git working tree.

This script is intentionally API-key-free: it creates files you can paste into
another model/chat for review.

Outputs a folder under: <git-root>/.ai-review/<timestamp>_<branch>/
- prompt.md
- status.txt
- diff_unstaged.patch
- diff_staged.patch
- files_changed.txt

Usage:
  python tools/ai_review/prepare_review.py
  python tools/ai_review/prepare_review.py --out .ai-review/custom
"""

from __future__ import annotations

import argparse
import datetime as _dt
import os
import pathlib
import subprocess
import sys
from typing import Iterable


def _run(cmd: list[str], cwd: pathlib.Path) -> str:
    try:
        res = subprocess.run(
            cmd,
            cwd=str(cwd),
            check=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
        )
    except subprocess.CalledProcessError as e:
        raise RuntimeError(
            f"Command failed: {' '.join(cmd)}\n\nSTDOUT:\n{e.stdout}\n\nSTDERR:\n{e.stderr}"
        ) from e
    return res.stdout


def _find_git_root(start: pathlib.Path) -> pathlib.Path | None:
    current = start.resolve()
    for parent in [current, *current.parents]:
        if (parent / ".git").exists():
            return parent
    return None


def _safe_name(value: str) -> str:
    return "".join(ch if ch.isalnum() or ch in ("-", "_", ".") else "_" for ch in value).strip("_")


def _write(path: pathlib.Path, content: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content, encoding="utf-8")


def _read_optional(path: pathlib.Path) -> str:
    try:
        return path.read_text(encoding="utf-8")
    except FileNotFoundError:
        return ""


def _lines(items: Iterable[str]) -> str:
    return "\n".join(items) + ("\n" if items else "")


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--out", type=str, default="", help="Output directory (default: auto under .ai-review)")
    args = ap.parse_args()

    cwd = pathlib.Path.cwd()
    git_root = _find_git_root(cwd)
    if git_root is None:
        # Fallback: try from this script's directory (useful when invoked from non-repo folders)
        script_dir = pathlib.Path(__file__).resolve().parent
        git_root = _find_git_root(script_dir)
    if git_root is None:
        # Final fallback: rely on git's discovery (may still fail)
        git_root = pathlib.Path(_run(["git", "rev-parse", "--show-toplevel"], cwd=cwd).strip())
    branch = _run(["git", "rev-parse", "--abbrev-ref", "HEAD"], cwd=git_root).strip()
    ts = _dt.datetime.now().strftime("%Y%m%d_%H%M%S")

    if args.out:
        out_dir = (git_root / args.out).resolve()
    else:
        out_dir = (git_root / ".ai-review" / f"{ts}_{_safe_name(branch)}").resolve()

    status = _run(["git", "status", "--porcelain=v1", "--branch"], cwd=git_root)
    diff_unstaged = _run(["git", "diff"], cwd=git_root)
    diff_staged = _run(["git", "diff", "--staged"], cwd=git_root)

    changed_files = []
    for line in status.splitlines():
        if not line or line.startswith("##"):
            continue
        # porcelain: XY <path>
        parts = line.split(maxsplit=1)
        if len(parts) == 2:
            changed_files.append(parts[1])

    # Add a bit of project context if present (keep short; reviewer can open files as needed)
    ponsu_dir = git_root / "ponsu"
    design_path = ponsu_dir / "docs" / "design" / "MVP設計書.md"
    tasks_path = ponsu_dir / "docs" / "tasks.yaml"

    design_hint = ""
    if design_path.exists():
        # Only include the first ~120 lines to avoid gigantic prompts.
        design_text = design_path.read_text(encoding="utf-8").splitlines()
        design_hint = _lines(design_text[:120])

    tasks_hint = ""
    if tasks_path.exists():
        tasks_text = tasks_path.read_text(encoding="utf-8").splitlines()
        tasks_hint = _lines(tasks_text[:180])

    prompt = f"""# Review Request (External Model)

You are a senior engineer performing a strict code review.

## Context
- Repo: {git_root.name}
- Branch: {branch}
- Date: {ts}

## What I need from you
1. Identify correctness bugs and edge cases.
2. Identify security/privacy issues (authz leaks, injection, unsafe file handling, secrets).
3. Identify architectural issues (clean architecture boundaries, event sourcing invariants).
4. Identify test gaps and suggest minimal tests.
5. Provide concrete, actionable change requests.

## Constraints
- Keep MVP scope in mind. Avoid gold-plating.
- Prefer minimal changes that fix root causes.

## Outputs provided
- status.txt
- diff_unstaged.patch
- diff_staged.patch
- files_changed.txt

## Project notes (excerpt)
### MVP設計書.md (excerpt)
{design_hint if design_hint else "(not included)"}

### tasks.yaml (excerpt)
{tasks_hint if tasks_hint else "(not included)"}

## Review checklist (use this as structure)
- Correctness
- Security
- Maintainability
- Performance
- Tests
- Documentation

## Diff to review
Paste and review BOTH diffs if present.
"""

    _write(out_dir / "status.txt", status)
    _write(out_dir / "diff_unstaged.patch", diff_unstaged)
    _write(out_dir / "diff_staged.patch", diff_staged)
    _write(out_dir / "files_changed.txt", _lines(changed_files))
    _write(out_dir / "prompt.md", prompt)

    print(str(out_dir))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
