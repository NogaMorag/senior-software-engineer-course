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
    average = a.traffic.requests_per_day / SECONDS_PER_DAY
    return Rate(average=average, peak=average * a.traffic.peak_factor)


def writes(a: Assumptions, r: Record) -> Rate:
    """How many of record r are written per second."""
    q = qps(a)
    return Rate(average=q.average * r.per_request, peak=q.peak * r.per_request)


def bytes_per_day(a: Assumptions, r: Record) -> float:
    """New raw bytes of record r per day, before indexes and replicas."""
    return a.traffic.requests_per_day * r.per_request * r.bytes


def retained_bytes(a: Assumptions) -> float:
    """Bytes on disk in steady state: each record kept for its retention, plus indexes, times
    the number of copies."""
    raw = sum(bytes_per_day(a, r) * r.retention_days for r in a.records)
    return raw * (1 + a.storage.index_overhead) * a.storage.replicas


def bandwidth(a: Assumptions) -> Rate:
    """Network bytes per second over all transfers."""
    per_request = sum(t.bytes * t.per_request for t in a.transfers)
    q = qps(a)
    return Rate(average=q.average * per_request, peak=q.peak * per_request)


_BINARY = ("B", "KiB", "MiB", "GiB", "TiB", "PiB")


def human_bytes(n: float) -> str:
    """Storage in powers of two: 1 KiB is 2^10 bytes, 1 MiB is 2^20, 1 GiB is 2^30."""
    for unit in _BINARY:
        if abs(n) < 1024 or unit == _BINARY[-1]:
            return f"{n:.0f} {unit}" if unit == "B" else f"{n:.1f} {unit}"
        n /= 1024
    raise AssertionError("unreachable")


_DECIMAL = ("bit/s", "kbit/s", "Mbit/s", "Gbit/s", "Tbit/s")


def human_bits_per_second(bytes_per_second: float) -> str:
    """Network speed in bits and powers of ten, the way links are sold: 1 Gbit/s is 10^9 bit/s."""
    n = bytes_per_second * 8
    for unit in _DECIMAL:
        if abs(n) < 1000 or unit == _DECIMAL[-1]:
            return f"{n:.1f} {unit}"
        n /= 1000
    raise AssertionError("unreachable")
