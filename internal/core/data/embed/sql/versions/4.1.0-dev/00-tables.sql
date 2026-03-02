--
-- Copyright (C) 2025 IOTech Ltd
--
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE core_data.device_info ADD COLUMN IF NOT EXISTS mark_deleted BOOLEAN DEFAULT false;

-- create index on reading(event_id) to enhance the query performance
CREATE INDEX IF NOT EXISTS idx_reading_event_id ON core_data.reading(event_id);

ALTER TABLE core_data.reading ADD COLUMN IF NOT EXISTS numeric_value NUMERIC;

-- create index on reading(device_info_id) to enhance the performance of queries that join reading with device_info on device_info_id
CREATE INDEX IF NOT EXISTS idx_reading_device_info_id ON core_data.reading(device_info_id);

-- create index on event(device_info_id) to enhance the performance of queries that join event with device_info on device_info_id
CREATE INDEX IF NOT EXISTS idx_event_device_info_id ON core_data.event(device_info_id);

-- -- create index on reading and device_info to enhance the performance of queries that join reading with device_info and specified device and resource
CREATE INDEX IF NOT EXISTS idx_reading_device_info_origin ON core_data.reading(device_info_id, origin DESC);
CREATE INDEX IF NOT EXISTS idx_device_info_device_resource ON core_data.device_info(mark_deleted, devicename, resourcename);
