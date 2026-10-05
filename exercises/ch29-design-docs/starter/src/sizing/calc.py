"""The arithmetic of a quick estimate. Every function is a few lines on purpose: the value of
an estimate is in the assumptions, which stay visible, not in clever formulas."""

from dataclasses import dataclass

from sizing.model import Assumptions, Record

SECONDS_PER_DAY = 86_400  # about 10^5: a million a day is about 12 a second


@dataclass(frozen=True)
class Rate:
    average: float  # per second
    peak: float  # per second


def qps(a: Assumptions) -> Rate:
    """Requests per second, on average and at the peak."""
    raise NotImplementedError(
        "TODO: requests_per_day spread over SECONDS_PER_DAY; peak = average x peak_factor"
    )


def writes(a: Assumptions, r: Record) -> Rate:
    """How many of record r are written per second."""
    raise NotImplementedError("TODO: qps(a) times r.per_request, for the average and the peak")


def bytes_per_day(a: Assumptions, r: Record) -> float:
    """New raw bytes of record r per day, before indexes and replicas."""
    raise NotImplementedError("TODO: requests per day x per_request x bytes")


def retained_bytes(a: Assumptions) -> float:
    """Bytes on disk in steady state: each record kept for its retention, plus indexes, times
    the number of copies."""
    raise NotImplementedError(
        "TODO: sum of bytes_per_day x retention_days, then indexes, then copies"
    )


def bandwidth(a: Assumptions) -> Rate:
    """Network bytes per second over all transfers."""
    raise NotImplementedError("TODO: bytes per request over all transfers, times qps(a)")


_BINARY = ("B", "KiB", "MiB", "GiB", "TiB", "PiB")


def human_bytes(n: float) -> str:
    """Storage in powers of two: 1 KiB is 2^10 bytes, 1 MiB is 2^20, 1 GiB is 2^30."""
    raise NotImplementedError("TODO: divide by 1024 until below 1024; see human_bits_per_second")


_DECIMAL = ("bit/s", "kbit/s", "Mbit/s", "Gbit/s", "Tbit/s")


def human_bits_per_second(bytes_per_second: float) -> str:
    """Network speed in bits and powers of ten, the way links are sold: 1 Gbit/s is 10^9 bit/s."""
    n = bytes_per_second * 8
    for unit in _DECIMAL:
        if abs(n) < 1000 or unit == _DECIMAL[-1]:
            return f"{n:.1f} {unit}"
        n /= 1000
    raise AssertionError("unreachable")
