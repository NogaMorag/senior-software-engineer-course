"""sizing FILE.toml: print the estimate for one assumptions file."""

import argparse
import sys
from pathlib import Path

from sizing.calc import (
    bandwidth,
    bytes_per_day,
    human_bits_per_second,
    human_bytes,
    qps,
    retained_bytes,
    writes,
)
from sizing.model import AssumptionError, Assumptions, load


def report(a: Assumptions) -> str:
    q = qps(a)
    lines = [
        f"== {a.name}",
        f"{'requests':<22} average {q.average:>10,.1f}/s   peak {q.peak:>10,.1f}/s",
    ]
    for r in a.records:
        w = writes(a, r)
        lines.append(
            f"{r.name:<22} average {w.average:>10,.1f}/s   peak {w.peak:>10,.1f}/s"
            f"   {human_bytes(bytes_per_day(a, r))}/day, kept {r.retention_days:g} days"
        )
    lines.append(
        f"{'storage retained':<22} {human_bytes(retained_bytes(a))}"
        f" (index overhead {a.storage.index_overhead:.0%}, copies {a.storage.replicas})"
    )
    b = bandwidth(a)
    lines.append(
        f"{'network':<22} average {human_bits_per_second(b.average)}"
        f"   peak {human_bits_per_second(b.peak)}"
    )
    return "\n".join(lines)


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(prog="sizing", description=__doc__)
    parser.add_argument("files", nargs="+", type=Path, help="assumptions files (TOML)")
    args = parser.parse_args(argv)
    for i, path in enumerate(args.files):
        try:
            a = load(path)
        except (OSError, AssumptionError) as e:
            print(f"sizing: {e}", file=sys.stderr)
            return 2
        print(("\n" if i else "") + report(a))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
