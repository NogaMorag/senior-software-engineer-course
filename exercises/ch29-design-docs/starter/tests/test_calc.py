from collections.abc import Callable
from pathlib import Path

import pytest
from conftest import VALID

from sizing.calc import (
    SECONDS_PER_DAY,
    bandwidth,
    bytes_per_day,
    human_bits_per_second,
    human_bytes,
    qps,
    retained_bytes,
    writes,
)
from sizing.model import Assumptions, load


@pytest.fixture
def a(write_toml: Callable[[str], Path]) -> Assumptions:
    return load(write_toml(VALID))


def test_a_day_is_about_ten_to_the_fifth_seconds() -> None:
    assert SECONDS_PER_DAY == 24 * 60 * 60
    assert 10**5 * 0.8 < SECONDS_PER_DAY < 10**5


def test_qps_average_and_peak(a: Assumptions) -> None:
    q = qps(a)
    assert q.average == pytest.approx(100.0)
    assert q.peak == pytest.approx(300.0)


def test_writes_follow_fan_out(a: Assumptions) -> None:
    w = writes(a, a.records[1])
    assert w.average == pytest.approx(200.0)
    assert w.peak == pytest.approx(600.0)


def test_storage_per_day_and_retained(a: Assumptions) -> None:
    assert bytes_per_day(a, a.records[0]) == pytest.approx(8_640_000 * 1_000)
    assert bytes_per_day(a, a.records[1]) == pytest.approx(8_640_000 * 2 * 200)
    raw = 8_640_000 * 1_000 * 10 + 8_640_000 * 400 * 5
    assert retained_bytes(a) == pytest.approx(raw * 1.5 * 2)


def test_bandwidth_sums_transfers(a: Assumptions) -> None:
    b = bandwidth(a)
    assert b.average == pytest.approx(100 * (1_000 + 2 * 500))
    assert b.peak == pytest.approx(300 * 2_000)


@pytest.mark.parametrize(
    ("n", "text"),
    [
        (0, "0 B"),
        (1023, "1023 B"),
        (1024, "1.0 KiB"),
        (1536, "1.5 KiB"),
        (2**20, "1.0 MiB"),
        (2**30 * 14, "14.0 GiB"),
        (2**40 * 1.9, "1.9 TiB"),
        (2**60, "1024.0 PiB"),
    ],
)
def test_human_bytes_uses_powers_of_two(n: float, text: str) -> None:
    assert human_bytes(n) == text


@pytest.mark.parametrize(
    ("bytes_per_second", "text"),
    [
        (100, "800.0 bit/s"),
        (125_000, "1.0 Mbit/s"),
        (125_000_000, "1.0 Gbit/s"),
    ],
)
def test_bits_per_second_use_powers_of_ten(bytes_per_second: float, text: str) -> None:
    assert human_bits_per_second(bytes_per_second) == text
