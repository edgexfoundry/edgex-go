--
-- Copyright (C) 2024 IOTech Ltd
--
-- SPDX-License-Identifier: Apache-2.0

-- core_keeper.config is used to store the config information
CREATE TABLE IF NOT EXISTS core_keeper.config (
    id UUID PRIMARY KEY,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    created timestamp NOT NULL DEFAULT now(),
    modified timestamp NOT NULL DEFAULT now()
);
