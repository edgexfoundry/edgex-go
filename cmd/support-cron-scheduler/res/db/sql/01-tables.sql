--
-- Copyright (C) 2024 IOTech Ltd
--
-- SPDX-License-Identifier: Apache-2.0

-- scheduler.schedule_job is used to store the schedule job information
CREATE TABLE IF NOT EXISTS scheduler.schedule_job (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    content JSONB NOT NULL,
    created timestamp NOT NULL DEFAULT now(),
    modified timestamp NOT NULL DEFAULT now()
);

-- scheduler.schedule_action_record is used to store the schedule action record
CREATE TABLE IF NOT EXISTS scheduler.schedule_action_record (
    id UUID PRIMARY KEY,
    action_id UUID NOT NULL,
    job_name TEXT NOT NULL,
    action JSONB NOT NULL,
    status TEXT NOT NULL,
    scheduled_at timestamp NOT NULL,
    created timestamp NOT NULL DEFAULT now()
);
