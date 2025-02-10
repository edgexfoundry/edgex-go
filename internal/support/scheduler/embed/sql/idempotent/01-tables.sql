--
-- Copyright (C) 2024 IOTech Ltd
--
-- SPDX-License-Identifier: Apache-2.0

-- support_scheduler.job is used to store the schedule job information
CREATE TABLE IF NOT EXISTS support_scheduler.job (
    id UUID PRIMARY KEY,
    content JSONB NOT NULL
);

-- support_scheduler.record is used to store the schedule action record
CREATE TABLE IF NOT EXISTS support_scheduler.record (
    id UUID PRIMARY KEY,
    action_id UUID NOT NULL,
    job_name TEXT NOT NULL,
    action JSONB NOT NULL,
    status TEXT NOT NULL,
    scheduled_at timestamp NOT NULL,
    created timestamp NOT NULL DEFAULT (now() AT TIME ZONE 'utc')
);
