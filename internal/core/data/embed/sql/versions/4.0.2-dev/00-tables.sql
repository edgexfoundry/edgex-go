--
-- Copyright (C) 2025 IOTech Ltd
--
-- SPDX-License-Identifier: Apache-2.0

-- create index on reading(device_info_id) to enhance the performance of queries that join reading with device_info on device_info_id
CREATE INDEX IF NOT EXISTS idx_reading_device_info_id ON core_data.reading(device_info_id);

-- create index on event(device_info_id) to enhance the performance of queries that join event with device_info on device_info_id
CREATE INDEX IF NOT EXISTS idx_event_device_info_id ON core_data.event(device_info_id);
