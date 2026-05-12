CREATE TABLE IF NOT EXISTS v_events (
  event_hash TEXT PRIMARY KEY,
  parent_hashes JSONB NOT NULL DEFAULT '[]'::jsonb,
  space_id TEXT NOT NULL,
  region_id TEXT NOT NULL,
  entity_id TEXT NOT NULL,
  target_entity_id TEXT NOT NULL DEFAULT '',
  vector_type TEXT NOT NULL,
  operation TEXT NOT NULL,
  input_vector JSONB NOT NULL,
  output_vector JSONB NOT NULL,
  magnitude_before DOUBLE PRECISION NOT NULL DEFAULT 0,
  magnitude_after DOUBLE PRECISION NOT NULL DEFAULT 0,
  direction_before JSONB NOT NULL DEFAULT '[]'::jsonb,
  direction_after JSONB NOT NULL DEFAULT '[]'::jsonb,
  space_coordinates_before JSONB NOT NULL DEFAULT '[]'::jsonb,
  space_coordinates_after JSONB NOT NULL DEFAULT '[]'::jsonb,
  time_created TIMESTAMPTZ NOT NULL,
  logical_order BIGINT NOT NULL,
  signer_id TEXT NOT NULL,
  signer_public_key TEXT NOT NULL DEFAULT '',
  signature TEXT NOT NULL DEFAULT '',
  validation_proof TEXT NOT NULL DEFAULT '',
  event_body_hash TEXT NOT NULL,
  metadata JSONB NOT NULL DEFAULT '[]'::jsonb,
  certified BOOLEAN NOT NULL DEFAULT FALSE,
  accepted BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS v_state_heads (
  entity_id TEXT PRIMARY KEY,
  space_id TEXT NOT NULL,
  region_id TEXT NOT NULL,
  latest_event_hash TEXT NOT NULL,
  latest_logical_order BIGINT NOT NULL,
  vector_state JSONB NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS v_snapshots (
  snapshot_id TEXT PRIMARY KEY,
  space_id TEXT NOT NULL,
  scope TEXT NOT NULL,
  root_hash TEXT NOT NULL,
  event_hash TEXT NOT NULL,
  state_root JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  sealed BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS v_peers (
  peer_id TEXT PRIMARY KEY,
  address TEXT NOT NULL,
  region_scope TEXT NOT NULL DEFAULT '',
  last_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  score DOUBLE PRECISION NOT NULL DEFAULT 0,
  metadata JSONB NOT NULL DEFAULT '[]'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_v_events_space_region_order ON v_events(space_id, region_id, logical_order, event_hash);
CREATE INDEX IF NOT EXISTS idx_v_events_entity_order ON v_events(entity_id, logical_order, event_hash);
CREATE INDEX IF NOT EXISTS idx_v_events_parent_hashes_gin ON v_events USING GIN (parent_hashes);
CREATE INDEX IF NOT EXISTS idx_v_state_heads_space_region ON v_state_heads(space_id, region_id);
