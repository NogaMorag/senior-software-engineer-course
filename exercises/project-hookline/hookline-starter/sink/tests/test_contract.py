"""Receiver side of the webhook contract (milestone M9): what the schema allows, the sink accepts.

The schema and the examples belong to the producer, in ../contract next to the Go module. This test
reads them from there and never imports Go code: the files are the whole agreement.
"""

import json
from pathlib import Path

import httpx
import pytest
from jsonschema import Draft202012Validator

from helpers import Headers

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
def test_valid_example_is_accepted_by_schema(path: Path) -> None:
    validator.validate(json.loads(path.read_bytes()))


@pytest.mark.parametrize("path", INVALID, ids=lambda p: p.name)
def test_invalid_example_is_rejected_by_schema(path: Path) -> None:
    errors = list(validator.iter_errors(json.loads(path.read_bytes())))
    assert errors, f"{path.name} was accepted"


# Read once, at import: the async test below must not block the event loop on file reads.
VALID_BODIES = {path.name: path.read_bytes() for path in VALID}


@pytest.mark.anyio
@pytest.mark.parametrize("name", sorted(VALID_BODIES))
async def test_sink_accepts_every_valid_example(
    name: str, client: httpx.AsyncClient, signed: Headers
) -> None:
    body = VALID_BODIES[name]
    headers = signed(body, event_id=json.loads(body)["id"])
    response = await client.post("/hooks/orders", content=body, headers=headers)
    assert response.status_code == 200, response.text
