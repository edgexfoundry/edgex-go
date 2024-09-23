--
-- Copyright (C) 2024 IOTech Ltd
--
-- SPDX-License-Identifier: Apache-2.0

-- core_metadata.device_service is used to store the device_service information
CREATE TABLE IF NOT EXISTS core_metadata.device_service (
    id UUID PRIMARY KEY,
    content JSONB NOT NULL
);

-- core_metadata.device_profile is used to store the device_profile information
CREATE TABLE IF NOT EXISTS core_metadata.device_profile (
    id UUID PRIMARY KEY,
    content JSONB NOT NULL
);
