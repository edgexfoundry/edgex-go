--
-- Copyright (C) 2025 IOTech Ltd
--
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE core_data.device_info ADD COLUMN IF NOT EXISTS mark_deleted BOOLEAN DEFAULT false;
