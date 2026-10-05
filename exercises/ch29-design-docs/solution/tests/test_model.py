from collections.abc import Callable
from pathlib import Path

import pytest
from conftest import VALID

from sizing.model import AssumptionError, load


def test_load_reads_every_section(write_toml: Callable[[str], Path]) -> None:
    a = load(write_toml(VALID))
    assert a.name == "test"
    assert a.traffic.requests_per_day == 8_640_000
    assert [r.name for r in a.records] == ["events", "deliveries"]
    assert [t.name for t in a.transfers] == ["in", "out"]
    assert a.storage.replicas == 2


@pytest.mark.parametrize(
    ("old", "new", "message"),
    [
        ("peak_factor = 3.0", "", "traffic: missing peak_factor"),
        ("peak_factor = 3.0", "peak_factor = 0.5", "traffic.peak_factor: 0.5 is below 1.0"),
        ("bytes = 200", "bytes = -200", "records[1].bytes: -200 is below 0.0"),
        ("bytes = 200", 'bytes = "200"', "records[1].bytes: expected a number"),
        ("replicas = 2", "replicas = 1.5", "storage.replicas: expected a whole number"),
        ('name = "test"', "", "top level: missing name"),
        ("[storage]", "[disk]", "missing table [storage]"),
    ],
)
def test_load_rejects_bad_values(
    write_toml: Callable[[str], Path], old: str, new: str, message: str
) -> None:
    with pytest.raises(AssumptionError, match=message.replace("[", r"\[").replace(".", r"\.")):
        load(write_toml(VALID.replace(old, new, 1)))


def test_load_reports_invalid_toml(write_toml: Callable[[str], Path]) -> None:
    with pytest.raises(AssumptionError, match="not valid TOML"):
        load(write_toml("name = "))
