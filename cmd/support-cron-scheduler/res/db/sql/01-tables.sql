--
-- Copyright (C) 2024 IOTech Ltd
--
-- SPDX-License-Identifier: Apache-2.0

-- cron_scheduler.schedule_job is used to store the schedule job information
CREATE TABLE IF NOT EXISTS cron_scheduler.schedule_job (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    job JSONB NOT NULL,
    created timestamptz NOT NULL DEFAULT now(),
    modified timestamptz NOT NULL DEFAULT now()
);

-- cron_scheduler.schedule_action_record is used to store the schedule action record
-- Note: All the records belong to the same job should have the same created time.
CREATE TABLE IF NOT EXISTS cron_scheduler.schedule_action_record (
    id UUID PRIMARY KEY,
    job_name TEXT NOT NULL,
    action JSONB NOT NULL,
    status TEXT NOT NULL,
    scheduled_at timestamptz NOT NULL,
    created timestamptz NOT NULL
);
