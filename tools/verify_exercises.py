"""Run every exercise's checks and print one line per module. Read-only apart from build caches.

    uv run --no-project python tools/verify_exercises.py [ch03 ch04 ...]

For each exercises/chNN-*/: canonical tests in tests/ must be byte-identical to their copies in starter/ and solution/.
Go modules (a go.mod in solution/ or starter/, or nested below): solution gets gofmt -l, go vet, staticcheck and
go test -race -count=1; starter gets go build and go vet. Python uv projects (pyproject.toml): solution gets
uv sync --frozen, ruff check, ruff format --check, mypy, pytest; starter gets uv sync --frozen and a ruff check.
Integration tests skip unless DATABASE_URL / REDIS_ADDR / KAFKA_BROKERS are set (bring the exercise's compose stack
up first to run them). Added 2026-10-05.
"""

from __future__ import annotations

import filecmp
import os
import re
import shutil
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
EXERCISES = ROOT / "exercises"
HOME = Path.home()
ENV = dict(os.environ, PATH=f"{HOME}/.local/go/bin:{HOME}/.local/bin:{os.environ['PATH']}", GOFLAGS="-mod=readonly")
SKIP_DIRS = {".venv", "node_modules", ".terraform", "testdata"}


def run(cmd: list[str], cwd: Path) -> tuple[bool, str]:
    proc = subprocess.run(cmd, cwd=cwd, env=ENV, capture_output=True, text=True)
    out = (proc.stdout + proc.stderr).strip()
    return proc.returncode == 0, out


def modules(base: Path, marker: str) -> list[Path]:
    found = []
    for path in base.rglob(marker):
        if not any(part in SKIP_DIRS for part in path.relative_to(base).parts):
            found.append(path.parent)
    return sorted(found)


def go_counts(output: str) -> str:
    passed = len(re.findall(r"^\s*--- PASS", output, re.M))
    skipped = len(re.findall(r"^\s*--- SKIP", output, re.M))
    failed = len(re.findall(r"^\s*--- FAIL", output, re.M))
    return f"{passed} passed, {skipped} skipped, {failed} failed (incl. subtests)"


def check_go_solution(mod: Path) -> list[str]:
    problems = []
    ok, out = run(["gofmt", "-l", "."], mod)
    if out:
        problems.append(f"gofmt: {out.splitlines()[:3]}")
    for cmd in (["go", "vet", "./..."], ["staticcheck", "./..."]):
        ok, out = run(cmd, mod)
        if not ok:
            problems.append(f"{cmd[0]}: {out[-400:]}")
    ok, out = run(["go", "test", "-race", "-count=1", "-v", "./..."], mod)
    summary = go_counts(out)
    if not ok:
        problems.append(f"go test FAILED: {out[-800:]}")
    return problems + [summary]


def check_go_starter(mod: Path) -> list[str]:
    problems = []
    for cmd in (["go", "build", "-o", os.devnull, "./..."], ["go", "vet", "./..."]):
        ok, out = run(cmd, mod)
        if not ok and "no packages to build" not in out:
            problems.append(f"starter {cmd[1]}: {out[-400:]}")
    return problems


def check_py(mod: Path, solution: bool) -> list[str]:
    problems = []
    ok, out = run(["uv", "sync", "--frozen", "--quiet"], mod)
    if not ok:
        return [f"uv sync: {out[-400:]}"]
    ok, out = run(["uv", "run", "ruff", "check", "."], mod)
    if not ok:
        problems.append(f"ruff: {out[-300:]}")
    if solution:
        ok, out = run(["uv", "run", "ruff", "format", "--check", "."], mod)
        if not ok:
            problems.append(f"ruff format: {out[-300:]}")
        target = "src" if (mod / "src").exists() else "."
        ok, out = run(["uv", "run", "mypy", target], mod)
        if not ok:
            problems.append(f"mypy: {out[-300:]}")
        ok, out = run(["uv", "run", "pytest", "-q", "-p", "no:cacheprovider"], mod)
        last = out.splitlines()[-1] if out else ""
        if not ok:
            problems.append(f"pytest FAILED: {out[-600:]}")
        problems.append(last)
    shutil.rmtree(mod / ".venv", ignore_errors=True)
    return problems


def check_test_copies(folder: Path) -> list[str]:
    problems = []
    tests = folder / "tests"
    if not tests.exists():
        return problems
    for src in tests.rglob("*"):
        if src.is_dir() or "__pycache__" in src.parts:
            continue
        rel = src.relative_to(tests)
        for side in ("starter", "solution"):
            base = folder / side
            if not base.exists():
                continue
            candidates = [base / rel, base / "tests" / rel]
            if not any(c.exists() and filecmp.cmp(src, c, shallow=False) for c in candidates):
                problems.append(f"tests/{rel} not identical in {side}/")
    return problems


def main(argv: list[str]) -> int:
    folders = sorted(EXERCISES.glob("ch*/"))
    if argv:
        folders = [f for f in folders if any(f.name.startswith(a) for a in argv)]
    failures = 0
    for folder in folders:
        lines = check_test_copies(folder)
        for side in ("solution", "starter"):
            base = folder / side
            if not base.exists():
                continue
            for mod in modules(base, "go.mod"):
                res = check_go_solution(mod) if side == "solution" else check_go_starter(mod)
                lines += [f"{mod.relative_to(EXERCISES)}: {r}" for r in res]
            for mod in modules(base, "pyproject.toml"):
                res = check_py(mod, side == "solution")
                lines += [f"{mod.relative_to(EXERCISES)}: {r}" for r in res]
        bad = [l for l in lines if any(k in l for k in ("FAILED", "gofmt:", "vet:", "staticcheck:", "build:", "ruff", "mypy:", "uv sync", "not identical"))]
        failures += len(bad)
        print(("FAIL " if bad else "ok   ") + folder.name)
        for line in lines:
            print("     ", line)
    print("OK" if not failures else f"{failures} problem(s)")
    return 1 if failures else 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
