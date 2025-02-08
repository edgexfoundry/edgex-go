--
-- Copyright (C) 2024 IOTech Ltd
--
-- SPDX-License-Identifier: Apache-2.0

-- core_keeper.config is used to store the config information
CREATE TABLE IF NOT EXISTS core_keeper.config (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    created timestamp NOT NULL DEFAULT (now() AT TIME ZONE 'utc'),
    modified timestamp NOT NULL DEFAULT (now() AT TIME ZONE 'utc')
);

-- core_keeper.registry is used to store the registry information
CREATE TABLE IF NOT EXISTS core_keeper.registry (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    content jsonb NOT NULL
);
