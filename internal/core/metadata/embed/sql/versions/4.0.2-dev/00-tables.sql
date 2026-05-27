--
-- Copyright (C) 2026 IOTech Ltd
--
-- SPDX-License-Identifier: Apache-2.0

-- idx_device_content_gin is a GIN index on the device content column to accelerate JSONB containment queries with '@>' operators,
-- such as lookups by ProfileName and ServiceName
CREATE INDEX IF NOT EXISTS idx_device_content_gin ON core_metadata.device USING GIN (content jsonb_path_ops);
