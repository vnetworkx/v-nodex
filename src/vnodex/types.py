from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any


JsonValue = Any
JsonObject = dict[str, JsonValue]


@dataclass(slots=True)
class EventRequest:
    operation: str
    entity_id: str
    payload: JsonObject = field(default_factory=dict)
    parent_hashes: list[str] = field(default_factory=list)
    region_id: str | None = None
    actor_pk: str | None = None
    logical_clock: int | None = None

    def to_dict(self) -> JsonObject:
        data: JsonObject = {
            "operation": self.operation,
            "entity_id": self.entity_id,
            "payload": self.payload,
            "parent_hashes": self.parent_hashes,
        }
        if self.region_id is not None:
            data["region_id"] = self.region_id
        if self.actor_pk is not None:
            data["actor_pk"] = self.actor_pk
        if self.logical_clock is not None:
            data["logical_clock"] = self.logical_clock
        return data


@dataclass(slots=True)
class EventRecord:
    raw: JsonObject

    @classmethod
    def from_dict(cls, data: JsonObject) -> "EventRecord":
        return cls(raw=data)

    def to_dict(self) -> JsonObject:
        return self.raw


@dataclass(slots=True)
class HealthResponse:
    status: str
    raw: JsonObject = field(default_factory=dict)

    @classmethod
    def from_dict(cls, data: JsonObject) -> "HealthResponse":
        status = str(data.get("status", "unknown"))
        return cls(status=status, raw=data)

    def to_dict(self) -> JsonObject:
        return self.raw


@dataclass(slots=True)
class StateView:
    raw: JsonObject

    @classmethod
    def from_dict(cls, data: JsonObject) -> "StateView":
        return cls(raw=data)

    def to_dict(self) -> JsonObject:
        return self.raw


@dataclass(slots=True)
class PeerInfo:
    raw: JsonObject

    @classmethod
    def from_dict(cls, data: JsonObject) -> "PeerInfo":
        return cls(raw=data)

    def to_dict(self) -> JsonObject:
        return self.raw


@dataclass(slots=True)
class SnapshotInfo:
    raw: JsonObject

    @classmethod
    def from_dict(cls, data: JsonObject) -> "SnapshotInfo":
        return cls(raw=data)

    def to_dict(self) -> JsonObject:
        return self.raw