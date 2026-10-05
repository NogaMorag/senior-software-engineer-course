"""Webhook event models: the contract every line of the input file must satisfy."""

from datetime import UTC, datetime
from typing import Annotated, Literal, Self
from uuid import UUID

from pydantic import (
    AwareDatetime,
    BaseModel,
    ConfigDict,
    Field,
    TypeAdapter,
    field_validator,
    model_validator,
)

type Currency = Literal["ILS", "USD", "EUR"]
type ShipmentStatus = Literal["picked_up", "in_transit", "delivered"]
TenantId = Annotated[str, Field(pattern=r"^[a-z][a-z0-9-]{2,31}$")]


class Strict(BaseModel):
    """Shared configuration: unknown fields are errors, instances are immutable."""

    model_config = ConfigDict(extra="forbid", frozen=True)


class EventBase(Strict):
    id: UUID
    tenant: TenantId
    created_at: AwareDatetime

    @field_validator("created_at")
    @classmethod
    def to_utc(cls, value: datetime) -> datetime:
        return value.astimezone(UTC)


class InvoicePaidData(Strict):
    invoice_id: str = Field(min_length=5)
    amount_cents: int = Field(gt=0, strict=True)
    currency: Currency

    @field_validator("invoice_id")
    @classmethod
    def invoice_prefix(cls, value: str) -> str:
        if not value.startswith("inv_"):
            raise ValueError("invoice_id must start with 'inv_'")
        return value


class InvoicePaid(EventBase):
    type: Literal["invoice.paid"]
    data: InvoicePaidData


class ShipmentMovedData(Strict):
    shipment_id: str = Field(min_length=1)
    status: ShipmentStatus
    location: str | None = None

    @model_validator(mode="after")
    def location_when_in_transit(self) -> Self:
        if self.status == "in_transit" and not self.location:
            raise ValueError("location is required when status is 'in_transit'")
        return self


class ShipmentMoved(EventBase):
    type: Literal["shipment.moved"]
    data: ShipmentMovedData


type Event = Annotated[InvoicePaid | ShipmentMoved, Field(discriminator="type")]

# The discriminator values; pydantic puts the matching one first in an error's location.
EVENT_TYPES = frozenset({"invoice.paid", "shipment.moved"})

EVENT_ADAPTER: TypeAdapter[Event] = TypeAdapter(Event)
