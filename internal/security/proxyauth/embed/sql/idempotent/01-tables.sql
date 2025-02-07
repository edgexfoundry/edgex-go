--
-- Copyright (C) 2025 IOTech Ltd
--
-- SPDX-License-Identifier: Apache-2.0

-- security_proxy_auth.key_store is used to store the key file
CREATE TABLE IF NOT EXISTS security_proxy_auth.key_store (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    content TEXT NOT NULL,
    created timestamp NOT NULL DEFAULT (now() AT TIME ZONE 'utc'),
    modified timestamp NOT NULL DEFAULT (now() AT TIME ZONE 'utc')
);
