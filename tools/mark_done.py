"""Mark chapters (by number) or appendices (by letter file) as done in book/README.md.

    uv run --no-project python tools/mark_done.py 3 4 5 a-glossary
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

README = Path(__file__).resolve().parent.parent / "book" / "README.md"


def main(argv: list[str]) -> None:
    text = README.read_text(encoding="utf-8")
    for arg in argv:
        key = f"`{int(arg):02d}-" if arg.isdigit() else f"`appendices/{arg}"
        text, n = re.subn(rf"^(\| [^|]+ \| \[{re.escape(key)}[^|]*\| )(?:⏳|🛠)( \|)$", r"\1✅\2", text, flags=re.M)
        print(arg, "marked" if n else "not found or already done")
    README.write_text(text, encoding="utf-8")


if __name__ == "__main__":
    main(sys.argv[1:])
