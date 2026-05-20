CREATE TABLE IF NOT EXISTS v_events (
    event_id TEXT PRIMARY KEY,
    parent_hashes JSONB NOT NULL DEFAULT '[]'::jsonb,
    region_id TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    operation TEXT NOT NULL,
    params JSONB NOT NULL DEFAULT '{}'::jsonb,
    actor_pk TEXT NOT NULL,
    signature TEXT,
    event_hash TEXT NOT NULL UNIQUE,
    payload_hash TEXT NOT NULL,
    state_root TEXT NOT NULL,
    region_root TEXT NOT NULL,
    logical_clock BIGINT NOT NULL,
    accepted BOOLEAN NOT NULL DEFAULT TRUE,
    certified BOOLEAN NOT NULL DEFAULT FALSE,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    raw_event JSONB NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_v_events_created_at ON v_events(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_v_events_region_id ON v_events(region_id);
CREATE INDEX IF NOT EXISTS idx_v_events_entity_id ON v_events(entity_id);
CREATE INDEX IF NOT EXISTS idx_v_events_state_root ON v_events(state_root);

CREATE TABLE IF NOT EXISTS v_node_state (
    node_id TEXT PRIMARY KEY,
    state_root TEXT NOT NULL,
    latest_clock BIGINT NOT NULL,
    event_count BIGINT NOT NULL,
    derived_state JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS v_peers (
    peer_id TEXT PRIMARY KEY,
    base_url TEXT,
    multiaddr TEXT,
    healthy BOOLEAN NOT NULL DEFAULT FALSE,
    score INTEGER NOT NULL DEFAULT 0,
    last_seen TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS v_snapshots (
    snapshot_id TEXT PRIMARY KEY,
    state_root TEXT NOT NULL,
    history_root TEXT NOT NULL,
    head_hashes JSONB NOT NULL DEFAULT '[]'::jsonb,
    region_root TEXT NOT NULL,
    logical_clock BIGINT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
