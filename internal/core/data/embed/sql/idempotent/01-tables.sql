--
-- Copyright (C) 2024-2025 IOTech Ltd
--
-- SPDX-License-Identifier: Apache-2.0

-- core_data.device_info is used to store the device related information reusing in the event and reading
CREATE TABLE IF NOT EXISTS core_data.device_info (
    id SERIAL PRIMARY KEY,
    devicename TEXT,
    profilename TEXT,
    sourcename TEXT,
    tags JSONB,
    resourcename TEXT,
    valuetype TEXT DEFAULT '',
    units TEXT DEFAULT '',
    mediatype TEXT DEFAULT ''
);

-- core_data.event is used to store the event information
CREATE TABLE IF NOT EXISTS core_data.event (
    id UUID,
    origin BIGINT,
    device_info_id SERIAL
);

CREATE INDEX IF NOT EXISTS idx_event_origin
    ON core_data.event(origin);

-- core_data.reading is used to store the reading information
CREATE TABLE IF NOT EXISTS core_data.reading (
    event_id UUID,
    device_info_id SERIAL,
    origin BIGINT,
    value TEXT,
    binaryvalue BYTEA,
    objectvalue JSONB
);

CREATE INDEX IF NOT EXISTS idx_reading_origin
    ON core_data.reading(origin);
