--
-- Copyright (C) 2024 IOTech Ltd
--
-- SPDX-License-Identifier: Apache-2.0

-- core-data.event is used to store the event information
CREATE TABLE IF NOT EXISTS "core-data".event (
    id UUID PRIMARY KEY,
    content JSONB NOT NULL,
    created timestamptz NOT NULL DEFAULT now()
);
