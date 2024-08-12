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
