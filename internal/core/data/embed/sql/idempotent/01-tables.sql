--
-- Copyright (C) 2024 IOTech Ltd
--
-- SPDX-License-Identifier: Apache-2.0

-- core_data.event is used to store the event information
CREATE TABLE IF NOT EXISTS core_data.event (
    id UUID PRIMARY KEY,
    devicename TEXT,
    profilename TEXT,
    sourcename TEXT,
    origin BIGINT,
    tags JSONB
);

-- core_data.reading is used to store the reading information
CREATE TABLE IF NOT EXISTS core_data.reading (
    id UUID,
    event_id UUID,
    devicename TEXT,
    profilename TEXT,
    resourcename TEXT,
    origin BIGINT,
    valuetype TEXT DEFAULT '',
    units TEXT DEFAULT '',
    tags JSONB,
    value TEXT,
    mediatype TEXT,
    binaryvalue BYTEA,
    objectvalue JSONB,
    CONSTRAINT fk_event
        FOREIGN KEY(event_id)
          REFERENCES core_data.event(id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_reading_id_origin
    ON core_data.reading(id, origin) -- Using id and origin as index for TimescaleDB integration
