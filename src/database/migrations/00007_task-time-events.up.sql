-- TASK EVENTS MIGRATION
-- Date: 2026-01-17
-- Purpose: Add unified event system for tasks (time tracking, comments, etc.)

-- Table: task_events
-- Unified event store for all task-related events
-- event_type: 'time_start', 'time_stop', 'comment' (future), etc.
-- event_data: JSONB for flexible payload (e.g., comment text, metadata)
CREATE TABLE IF NOT EXISTS task_events (
  id          BIGSERIAL PRIMARY KEY,
  task_id     BIGINT NOT NULL,
  user_id     BIGINT NOT NULL,
  event_type  TEXT NOT NULL,
  event_data  JSONB DEFAULT '{}',
  event_time  TIMESTAMP NOT NULL DEFAULT NOW(),
  created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (task_id) REFERENCES tasks (id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users (id)
);

-- Index for efficient lookups by task
CREATE INDEX idx_task_events_task_id ON task_events (task_id);

-- Index for efficient lookups by user
CREATE INDEX idx_task_events_user_id ON task_events (user_id);

-- Index for filtering by event type
CREATE INDEX idx_task_events_event_type ON task_events (event_type);

-- Index for ordering by event time
CREATE INDEX idx_task_events_event_time ON task_events (event_time DESC);

-- Composite index for common query pattern: get events for a task by type
CREATE INDEX idx_task_events_task_type ON task_events (task_id, event_type);

-- Stored procedure: Insert a task event
CREATE OR REPLACE FUNCTION sp_insert_task_event(
  p_task_id BIGINT,
  p_user_id BIGINT,
  p_event_type TEXT,
  p_event_data JSONB DEFAULT '{}'
)
RETURNS BIGINT AS $$
DECLARE
  new_id BIGINT;
BEGIN
  INSERT INTO task_events (task_id, user_id, event_type, event_data, event_time, created_at)
  VALUES (p_task_id, p_user_id, p_event_type, p_event_data, NOW(), NOW())
  RETURNING id INTO new_id;
  RETURN new_id;
END;
$$ LANGUAGE plpgsql;

-- Stored procedure: Calculate total time spent on a task (in seconds)
-- Pairs time_start/time_stop events and sums the durations
CREATE OR REPLACE FUNCTION sp_get_task_time_spent(p_task_id BIGINT)
RETURNS BIGINT AS $$
DECLARE
  total_seconds BIGINT := 0;
  start_time TIMESTAMP;
  stop_time TIMESTAMP;
  event_row RECORD;
  in_progress BOOLEAN := FALSE;
BEGIN
  FOR event_row IN
    SELECT event_type, event_time
    FROM task_events
    WHERE task_id = p_task_id
      AND event_type IN ('time_start', 'time_stop')
    ORDER BY event_time ASC
  LOOP
    IF event_row.event_type = 'time_start' THEN
      start_time := event_row.event_time;
      in_progress := TRUE;
    ELSIF event_row.event_type = 'time_stop' AND in_progress THEN
      stop_time := event_row.event_time;
      total_seconds := total_seconds + EXTRACT(EPOCH FROM (stop_time - start_time))::BIGINT;
      in_progress := FALSE;
    END IF;
  END LOOP;
  
  -- If currently in progress, add time from last start to now
  IF in_progress THEN
    total_seconds := total_seconds + EXTRACT(EPOCH FROM (NOW() - start_time))::BIGINT;
  END IF;
  
  RETURN total_seconds;
END;
$$ LANGUAGE plpgsql;

-- Stored procedure: Check if a task is currently being tracked (has an unpaired time_start event)
CREATE OR REPLACE FUNCTION sp_is_task_in_progress(p_task_id BIGINT)
RETURNS BOOLEAN AS $$
DECLARE
  latest_event TEXT;
BEGIN
  SELECT event_type INTO latest_event
  FROM task_events
  WHERE task_id = p_task_id
    AND event_type IN ('time_start', 'time_stop')
  ORDER BY event_time DESC
  LIMIT 1;
  
  RETURN COALESCE(latest_event = 'time_start', FALSE);
END;
$$ LANGUAGE plpgsql;
