"""Build the Hookline starter from the reference implementation.

    uv run --no-project python tools/make_project_starter.py

Go: every function and method body outside `KEEP_WHOLE` becomes `panic("TODO: see chapter 38")` (via tools/gostub),
unused imports are pruned so the starter compiles, `_test.go` files are kept as the specification, and `main`
functions are kept. Python (the `sink` receiver): every function body becomes `raise NotImplementedError`.
Kept as is: go.mod/go.sum, migrations, OpenAPI spec, compose files, Helm chart, Terraform, workflows, k6 scripts,
dashboards, alert rules, docs and tests. Adapted from book 1 on 2026-10-05.
"""

from __future__ import annotations

import ast
import os
import shutil
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
SRC = ROOT / "exercises/project-hookline/hookline"
DST = ROOT / "exercises/project-hookline/hookline-starter"
# Scaffolding the reader receives whole: config, wiring, generated code and small shared types.
KEEP_WHOLE = {
    "internal/config/config.go",
    "internal/platform/env.go",
    "internal/platform/logging.go",
    "internal/platform/middleware.go",
    "internal/platform/server.go",
    "internal/platform/debug.go",
    "internal/platform/json.go",
    "internal/problem/problem.go",  # chapter 9, unchanged
    "internal/page/page.go",  # chapter 9, unchanged
    "internal/domain/types.go",
    "internal/domain/errors.go",
    "internal/domain/id.go",
    "internal/domain/apikey.go",
    "internal/domain/validate.go",
    "internal/api/server.go",  # routes and middleware order: the wiring
    "internal/db/db.go",  # chapter 11, unchanged
    "internal/testdb/testdb.go",  # test infrastructure
    "internal/store/storetest/storetest.go",  # the conformance suite is a specification
    "internal/wire/wire.go",  # connections shared by the binaries
    "internal/bus/bus.go",  # chapter 13 client construction, plus CreateTopic
    "internal/testenv/testenv.go",  # test infrastructure
    "internal/testenv/fixtures.go",  # test infrastructure
    "internal/telemetry/telemetry.go",
    "internal/telemetry/start.go",  # M8: the no-op switch around Setup
    "internal/platform/route.go",  # chapter 24, unchanged
    "internal/oteltest/oteltest.go",  # test infrastructure
    "internal/ssrf/ssrf.go",  # chapter 28, plus Transport; cmd/worker's main calls it from M5 on
    "internal/ssrf/hookline.go",  # M9: Hookline's policy and resolver; cmd/api's main calls them
    "internal/redact/redact.go",  # chapter 28, plus Handler; platform.NewLogger wraps every output in it
}
KEEP_PY = {"__init__.py", "settings.py"}
IGNORE = shutil.ignore_patterns(".venv", "__pycache__", ".pytest_cache", ".ruff_cache", ".mypy_cache", "bin", "out", ".terraform", "*.tfstate*", "coverage.out")


class Stubber(ast.NodeTransformer):
    def _stub(self, node: ast.FunctionDef | ast.AsyncFunctionDef):
        doc = ast.get_docstring(node)
        body: list[ast.stmt] = [node.body[0]] if doc else []
        body.append(ast.parse("raise NotImplementedError  # TODO: see chapter 38").body[0])
        node.body = body
        return node

    def visit_FunctionDef(self, node):  # noqa: N802 - ast visitor API
        return self._stub(node)

    def visit_AsyncFunctionDef(self, node):  # noqa: N802
        return self._stub(node)


def stub_python(text: str) -> str:
    tree = Stubber().visit(ast.parse(text))
    ast.fix_missing_locations(tree)
    return ast.unparse(tree) + "\n"


def main() -> None:
    if DST.exists():
        shutil.rmtree(DST)
    shutil.copytree(SRC, DST, ignore=IGNORE)
    go_files = [
        p for p in sorted(DST.rglob("*.go"))
        if not p.name.endswith("_test.go")
        and p.relative_to(DST).as_posix() not in KEEP_WHOLE
        and "/cmd/" not in f"/{p.relative_to(DST).as_posix()}"
        and not p.name.endswith(".pb.go")
        and "testutil" not in p.parts
    ]
    env = dict(os.environ, GOFLAGS="-mod=mod")
    subprocess.run(["go", "run", ".", *map(str, go_files)], cwd=ROOT / "tools/gostub", check=True, env=env)
    py_stubbed = 0
    for path in sorted((DST / "sink" / "src").rglob("*.py")) if (DST / "sink" / "src").exists() else []:
        if path.name in KEEP_PY:
            continue
        path.write_text(stub_python(path.read_text(encoding="utf-8")), encoding="utf-8")
        py_stubbed += 1
    (DST / "STARTER.md").write_text(
        "# Hookline: starter\n\nThe full project structure, tests, migrations, contracts, charts, workflows and scripts, with\n"
        "every Go function body replaced by `panic(\"TODO: see chapter 38\")` and every Python function body by\n"
        "`raise NotImplementedError`. Chapter 38 walks through the ten milestones; each names the packages to implement\n"
        "and the tests that must pass. The reference implementation is next to this directory (`../hookline`).\n\n"
        "```bash\ngo build ./...          # compiles from day one\ngo test ./...           # red at first; green package by package, M1 to M10\n```\n",
        encoding="utf-8",
    )
    print(f"starter written to {DST.relative_to(ROOT)} ({len(go_files)} Go files, {py_stubbed} Python modules stubbed)")


if __name__ == "__main__":
    main()
