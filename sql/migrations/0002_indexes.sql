CREATE INDEX IF NOT EXISTS idx_v_events_time_created ON v_events(time_created);
CREATE INDEX IF NOT EXISTS idx_v_snapshots_space_scope ON v_snapshots(space_id, scope, created_at DESC);
