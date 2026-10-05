"""The assumptions file: a TOML document, read with the standard library and validated."""

import tomllib
from dataclasses import dataclass
from pathlib import Path
from typing import Any


class AssumptionError(ValueError):
    """The assumptions file is missing a value or holds one that makes no sense."""


@dataclass(frozen=True)
class Traffic:
    requests_per_day: float  # the unit of work the design is about (an ingested event, a page view)
    peak_factor: float  # peak rate divided by the average rate


@dataclass(frozen=True)
class Record:
    """Something stored for every request: a row, an object, a log line."""

    name: str
    bytes: float  # size of one, including per-row overhead you know about
    per_request: float  # how many are written per request (fan-out, retries)
    retention_days: float


@dataclass(frozen=True)
class Transfer:
    """Something sent over the network for every request."""

    name: str
    bytes: float
    per_request: float


@dataclass(frozen=True)
class Storage:
    index_overhead: float  # extra space for indexes, as a fraction of the data (0.5 is 50%)
    replicas: int  # copies of every byte: the primary plus each replica


@dataclass(frozen=True)
class Assumptions:
    name: str
    traffic: Traffic
    records: tuple[Record, ...]
    transfers: tuple[Transfer, ...]
    storage: Storage


def _number(table: dict[str, Any], key: str, where: str, *, minimum: float = 0.0) -> float:
    if key not in table:
        raise AssumptionError(f"{where}: missing {key}")
    value = table[key]
    if isinstance(value, bool) or not isinstance(value, int | float):
        raise AssumptionError(f"{where}.{key}: expected a number, got {value!r}")
    if value < minimum:
        raise AssumptionError(f"{where}.{key}: {value} is below {minimum}")
    return float(value)


def _whole(value: float, where: str) -> int:
    if not value.is_integer():
        raise AssumptionError(f"{where}: expected a whole number, got {value}")
    return int(value)


def _table(doc: dict[str, Any], key: str) -> dict[str, Any]:
    value = doc.get(key)
    if not isinstance(value, dict):
        raise AssumptionError(f"missing table [{key}]")
    return value


def _array(doc: dict[str, Any], key: str) -> list[dict[str, Any]]:
    value = doc.get(key, [])
    if not isinstance(value, list) or not all(isinstance(v, dict) for v in value):
        raise AssumptionError(f"{key}: expected an array of tables, [[{key}]]")
    return value


def _name(table: dict[str, Any], where: str) -> str:
    name = table.get("name")
    if not isinstance(name, str) or not name:
        raise AssumptionError(f"{where}: missing name")
    return name


def parse(doc: dict[str, Any]) -> Assumptions:
    """Validate a parsed TOML document and turn it into Assumptions."""
    traffic = _table(doc, "traffic")
    storage = _table(doc, "storage")
    records = tuple(
        Record(
            name=_name(r, f"records[{i}]"),
            bytes=_number(r, "bytes", f"records[{i}]"),
            per_request=_number(r, "per_request", f"records[{i}]"),
            retention_days=_number(r, "retention_days", f"records[{i}]"),
        )
        for i, r in enumerate(_array(doc, "records"))
    )
    transfers = tuple(
        Transfer(
            name=_name(t, f"transfers[{i}]"),
            bytes=_number(t, "bytes", f"transfers[{i}]"),
            per_request=_number(t, "per_request", f"transfers[{i}]"),
        )
        for i, t in enumerate(_array(doc, "transfers"))
    )
    return Assumptions(
        name=_name(doc, "top level"),
        traffic=Traffic(
            requests_per_day=_number(traffic, "requests_per_day", "traffic"),
            peak_factor=_number(traffic, "peak_factor", "traffic", minimum=1.0),
        ),
        records=records,
        transfers=transfers,
        storage=Storage(
            index_overhead=_number(storage, "index_overhead", "storage"),
            replicas=_whole(
                _number(storage, "replicas", "storage", minimum=1.0), "storage.replicas"
            ),
        ),
    )


def load(path: Path) -> Assumptions:
    """Read and validate an assumptions file."""
    try:
        doc = tomllib.loads(path.read_text(encoding="utf-8"))
    except tomllib.TOMLDecodeError as e:
        raise AssumptionError(f"{path}: not valid TOML: {e}") from e
    return parse(doc)
