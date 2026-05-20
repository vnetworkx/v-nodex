from __future__ import annotations

import json
from dataclasses import dataclass
from typing import Any, Callable
from urllib import error, parse, request

from .exceptions import VNodeXError, VNodeXHTTPError
from .types import EventRecord, EventRequest, HealthResponse, PeerInfo, SnapshotInfo, StateView


JsonValue = Any
JsonObject = dict[str, JsonValue]


@dataclass(slots=True)
class _Response:
    status: int
    body: str
    data: JsonObject | list[JsonValue] | None


class VNodeXClient:
    """
    Minimal Python client for v-nodex HTTP APIs.

    This client intentionally stays generic because the exact node API contract
    may evolve. The methods below map to common node operations.
    """

    def __init__(
        self,
        base_url: str,
        *,
        timeout: float = 30.0,
        token: str | None = None,
        opener: Callable[..., Any] | None = None,
    ) -> None:
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout
        self.token = token
        self._opener = opener or request.urlopen

    def health(self) -> HealthResponse:
        data = self._request_json("GET", "/health")
        return HealthResponse.from_dict(self._ensure_object(data))

    def info(self) -> JsonObject:
        data = self._request_json("GET", "/info")
        return self._ensure_object(data)

    def submit_event(self, event: EventRequest | JsonObject) -> EventRecord:
        payload = event.to_dict() if isinstance(event, EventRequest) else event
        data = self._request_json("POST", "/events", payload)
        return EventRecord.from_dict(self._ensure_object(data))

    def get_event(self, event_hash: str) -> EventRecord:
        data = self._request_json("GET", f"/events/{parse.quote(event_hash)}")
        return EventRecord.from_dict(self._ensure_object(data))

    def list_events(
        self,
        *,
        entity_id: str | None = None,
        region_id: str | None = None,
        limit: int | None = None,
    ) -> list[EventRecord]:
        query: dict[str, str] = {}
        if entity_id is not None:
            query["entity_id"] = entity_id
        if region_id is not None:
            query["region_id"] = region_id
        if limit is not None:
            query["limit"] = str(limit)

        path = "/events"
        if query:
            path = f"{path}?{parse.urlencode(query)}"

        data = self._request_json("GET", path)
        items = self._ensure_list(data)
        return [EventRecord.from_dict(self._ensure_object(item)) for item in items]

    def state(self, *, entity_id: str | None = None, region_id: str | None = None) -> StateView:
        query: dict[str, str] = {}
        if entity_id is not None:
            query["entity_id"] = entity_id
        if region_id is not None:
            query["region_id"] = region_id

        path = "/state"
        if query:
            path = f"{path}?{parse.urlencode(query)}"

        data = self._request_json("GET", path)
        return StateView.from_dict(self._ensure_object(data))

    def peers(self) -> list[PeerInfo]:
        data = self._request_json("GET", "/peers")
        items = self._ensure_list(data)
        return [PeerInfo.from_dict(self._ensure_object(item)) for item in items]

    def snapshots(self) -> list[SnapshotInfo]:
        data = self._request_json("GET", "/snapshots")
        items = self._ensure_list(data)
        return [SnapshotInfo.from_dict(self._ensure_object(item)) for item in items]

    def replay(self, *, from_snapshot: str | None = None) -> JsonObject:
        payload: JsonObject = {}
        if from_snapshot is not None:
            payload["from_snapshot"] = from_snapshot
        data = self._request_json("POST", "/replay", payload)
        return self._ensure_object(data)

    def _request_json(self, method: str, path: str, payload: JsonObject | None = None) -> JsonValue:
        response = self._request(method, path, payload)
        return response.data

    def _request(self, method: str, path: str, payload: JsonObject | None = None) -> _Response:
        url = f"{self.base_url}{path}"
        headers = {
            "Accept": "application/json",
            "User-Agent": "vnodex-python/0.1.0",
        }
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"

        body: bytes | None = None
        if payload is not None:
            body = json.dumps(payload, separators=(",", ":"), sort_keys=True).encode("utf-8")
            headers["Content-Type"] = "application/json"

        req = request.Request(url=url, method=method.upper(), headers=headers, data=body)

        try:
            with self._opener(req, timeout=self.timeout) as resp:
                raw = resp.read().decode("utf-8")
                data = json.loads(raw) if raw else None
                return _Response(status=getattr(resp, "status", 200), body=raw, data=data)
        except error.HTTPError as exc:
            body_text = exc.read().decode("utf-8", errors="replace") if exc.fp else None
            raise VNodeXHTTPError(exc.code, exc.reason, body_text) from exc
        except error.URLError as exc:
            raise VNodeXError(f"Failed to reach v-nodex at {url}: {exc.reason}") from exc
        except json.JSONDecodeError as exc:
            raise VNodeXError(f"Invalid JSON returned by v-nodex at {url}") from exc

    @staticmethod
    def _ensure_object(value: JsonValue) -> JsonObject:
        if not isinstance(value, dict):
            raise VNodeXError(f"Expected JSON object, got {type(value).__name__}")
        return value

    @staticmethod
    def _ensure_list(value: JsonValue) -> list[JsonValue]:
        if not isinstance(value, list):
            raise VNodeXError(f"Expected JSON list, got {type(value).__name__}")
        return value