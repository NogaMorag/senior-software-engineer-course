"""The event contract, case by case."""

from datetime import UTC
from typing import Any

import pytest
from pydantic import ValidationError

from eventcheck.models import EVENT_ADAPTER, InvoicePaid, ShipmentMoved


def test_invoice_parses_into_its_own_model(invoice: dict[str, Any]) -> None:
    event = EVENT_ADAPTER.validate_python(invoice)
    assert isinstance(event, InvoicePaid)
    assert event.data.amount_cents == 1999


def test_created_at_is_normalized_to_utc(invoice: dict[str, Any]) -> None:
    event = EVENT_ADAPTER.validate_python(invoice)
    assert event.created_at.tzinfo is UTC
    assert event.created_at.hour == 9  # 12:00 at +03:00


def test_shipment_parses_into_its_own_model(shipment: dict[str, Any]) -> None:
    assert isinstance(EVENT_ADAPTER.validate_python(shipment), ShipmentMoved)


def test_models_are_frozen(invoice: dict[str, Any]) -> None:
    event = EVENT_ADAPTER.validate_python(invoice)
    with pytest.raises(ValidationError):
        event.tenant = "other"  # type: ignore[misc]


@pytest.mark.parametrize(
    ("path", "value", "location"),
    [
        (("data", "amount_cents"), 0, ("invoice.paid", "data", "amount_cents")),
        (("data", "amount_cents"), "1999", ("invoice.paid", "data", "amount_cents")),
        (("data", "currency"), "GBP", ("invoice.paid", "data", "currency")),
        (("data", "invoice_id"), "bill_000001", ("invoice.paid", "data", "invoice_id")),
        (("tenant",), "Acme Corp", ("invoice.paid", "tenant")),
        (("created_at",), "2026-10-01T12:00:00", ("invoice.paid", "created_at")),
        (("id",), "not-a-uuid", ("invoice.paid", "id")),
        (("data", "extra"), 1, ("invoice.paid", "data", "extra")),
    ],
    ids=[
        "zero-amount",
        "amount-as-string",
        "unknown-currency",
        "invoice-prefix",
        "tenant-pattern",
        "naive-datetime",
        "bad-uuid",
        "extra-field",
    ],
)
def test_invalid_invoice_fields(
    invoice: dict[str, Any], path: tuple[str, ...], value: object, location: tuple[str, ...]
) -> None:
    target = invoice
    for key in path[:-1]:
        target = target[key]
    target[path[-1]] = value
    with pytest.raises(ValidationError) as caught:
        EVENT_ADAPTER.validate_python(invoice)
    assert [error["loc"] for error in caught.value.errors()] == [location]


def test_in_transit_requires_location(shipment: dict[str, Any]) -> None:
    del shipment["data"]["location"]
    with pytest.raises(ValidationError, match="location is required"):
        EVENT_ADAPTER.validate_python(shipment)


@pytest.mark.parametrize("status", ["picked_up", "delivered"])
def test_location_is_optional_otherwise(shipment: dict[str, Any], status: str) -> None:
    shipment["data"] = {"shipment_id": "shp_000002", "status": status}
    assert isinstance(EVENT_ADAPTER.validate_python(shipment), ShipmentMoved)


def test_unknown_type_is_rejected(invoice: dict[str, Any]) -> None:
    invoice["type"] = "invoice.refunded"
    with pytest.raises(ValidationError) as caught:
        EVENT_ADAPTER.validate_python(invoice)
    assert caught.value.errors()[0]["type"] == "union_tag_invalid"
