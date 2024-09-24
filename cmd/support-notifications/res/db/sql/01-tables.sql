--
-- Copyright (C) 2024 IOTech Ltd
--
-- SPDX-License-Identifier: Apache-2.0

-- support_notifications.notification is used to store the notification information
CREATE TABLE IF NOT EXISTS support_notifications.notification (
    id UUID PRIMARY KEY,
    content JSONB NOT NULL -- Note that this content is not the same as the content in Notification model
);

-- support_notifications.subscription is used to store the subscription information
CREATE TABLE IF NOT EXISTS support_notifications.subscription (
    id UUID PRIMARY KEY,
    content JSONB NOT NULL
);

-- support_notifications.transmission is used to store the transmission information
CREATE TABLE IF NOT EXISTS support_notifications.transmission (
    id UUID PRIMARY KEY,
    notification_id UUID NOT NULL,
    content JSONB NOT NULL,
    CONSTRAINT fk_notification
        FOREIGN KEY(notification_id)
        REFERENCES support_notifications.notification(id)
        ON DELETE CASCADE
);
