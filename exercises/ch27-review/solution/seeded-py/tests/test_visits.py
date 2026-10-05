import pytest

from tally import parse_line, tally, top


@pytest.mark.parametrize(
    ("line", "expected"),
    [
        ("Noga,he", ("noga", "he")),
        ("  dana , EN ", ("dana", "en")),
        ("", None),
        ("# header", None),
        ("too,many,fields", None),
        ("nofields", None),
    ],
)
def test_parse_line(line: str, expected: tuple[str, str] | None) -> None:
    assert parse_line(line) == expected


def test_tally_counts() -> None:
    assert tally(["noga,he", "dana,en", "noga,he", "# x"]) == {"noga": 2, "dana": 1}


def test_tally_calls_are_independent() -> None:
    first = tally(["noga,he"])
    second = tally(["dana,en"])
    assert first == {"noga": 1}
    assert second == {"dana": 1}


def test_tally_adds_to_given_counts() -> None:
    counts = {"noga": 5}
    assert tally(["noga,he"], counts) == {"noga": 6}


def test_top() -> None:
    assert top({"noga": 2, "dana": 2, "lior": 1}) == "dana"
    assert top({}) is None
