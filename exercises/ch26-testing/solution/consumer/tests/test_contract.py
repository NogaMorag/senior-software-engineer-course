"""Consumer side of the contract: everything the schema allows, this consumer accepts.

The schema and the examples belong to the producer (the Go module next to this project). This
test reads them from there and never imports Go code: the files are the whole agreement.
"""

import json
from pathlib import Path

import pytest
from jsonschema import Draft202012Validator

from consumer.event import parse_event

CONTRACT = Path(__file__).resolve().parents[2] / "contract"
SCHEMA = json.loads((CONTRACT / "webhook-event.schema.json").read_text(encoding="utf-8"))
EXAMPLES = CONTRACT / "testdata" / "examples"
VALID = sorted((EXAMPLES / "valid").glob("*.json"))
INVALID = sorted((EXAMPLES / "invalid").glob("*.json"))

# Without format_checker, "format" is only an annotation and "evt_123" passes as a uuid.
validator = Draft202012Validator(SCHEMA, format_checker=Draft202012Validator.FORMAT_CHECKER)


def test_schema_is_valid_2020_12() -> None:
    Draft202012Validator.check_schema(SCHEMA)


def test_formats_are_really_checked() -> None:
    # date-time is checked only when rfc3339-validator is installed; fail loudly if it is not.
    assert "date-time" in Draft202012Validator.FORMAT_CHECKER.checkers
    assert "uuid" in Draft202012Validator.FORMAT_CHECKER.checkers


def test_there_are_examples() -> None:
    assert VALID
    assert INVALID


@pytest.mark.parametrize("path", VALID, ids=lambda p: p.name)
def test_valid_example_is_accepted_by_schema_and_consumer(path: Path) -> None:
    body = path.read_bytes()
    doc = json.loads(body)
    validator.validate(doc)
    event = parse_event(body)
    assert str(event.id) == doc["id"].lower()
    assert event.type == doc["type"]
    assert event.created_at.tzinfo is not None
    assert event.data == doc["data"]


@pytest.mark.parametrize("path", INVALID, ids=lambda p: p.name)
def test_invalid_example_is_rejected_by_schema(path: Path) -> None:
    errors = list(validator.iter_errors(json.loads(path.read_bytes())))
    assert errors, f"{path.name} was accepted"


def test_unknown_properties_are_ignored() -> None:
    body = (EXAMPLES / "valid" / "extra-property.json").read_bytes()
    assert "api_version" in json.loads(body)
    assert parse_event(body).type == "invoice.paid"
