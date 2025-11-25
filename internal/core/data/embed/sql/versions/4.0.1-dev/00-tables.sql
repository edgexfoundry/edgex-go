--
-- Copyright (C) 2025 IOTech Ltd
--
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE core_data.device_info ADD COLUMN IF NOT EXISTS mark_deleted BOOLEAN DEFAULT false;

-- create index on reading(event_id) to enhance the query performance
CREATE INDEX IF NOT EXISTS idx_reading_event_id ON core_data.reading(event_id);
