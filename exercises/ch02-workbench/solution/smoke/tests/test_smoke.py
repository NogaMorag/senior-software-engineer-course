"""Smoke tests: the right Python, inside the project environment, importing the project."""

import sys

from smoke import describe


def test_python_is_3_14() -> None:
    assert sys.version_info[:2] == (3, 14)


def test_runs_inside_the_project_environment() -> None:
    # `uv run` puts the project's .venv first; a stray system Python fails here.
    assert sys.prefix != sys.base_prefix


def test_describe_names_version_and_system() -> None:
    text = describe()
    assert text.startswith("Python 3.14.")
    assert " on " in text
