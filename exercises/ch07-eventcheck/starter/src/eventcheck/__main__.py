"""Allows `python -m eventcheck`, which is also what `python -m cProfile -m eventcheck` runs."""

from eventcheck.cli import main

raise SystemExit(main())
