/*******************************************************************************
 * Copyright 2017 Dell Inc.
 * Copyright 2018 Dell Technologies Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 *******************************************************************************/

package notifications

var (
	/* ----------------------- CONSTANTS ----------------------------*/
	ESCALATIONSUBSCRIPTIONSLUG = "ESCALATION"
	ESCALATIONPREFIX           = "escalated-"
	ESCALATEDCONTENTNOTICE     = "This notification is escalated by the transmission"

	/* ---------------- URL PARAM NAMES -----------------------*/
	START        = "start"
	END          = "end"
	LIMIT        = "limit"
	NOTIFICATION = "notification"
	SUBSCRIPTION = "subscription"
	TRANSMISSION = "transmission"
	CLEANUP      = "cleanup"
	SLUG         = "slug"
	LABELS       = "labels"
	CATEGORIES   = "categories"
	ID           = "id"
	SENDER       = "sender"
	RECEIVER     = "receiver"
	AGE          = "age"
	NEW          = "new"
	ESCALATED    = "escalated"
	ACKNOWLEDGED = "acknowledged"
	FAILED       = "failed"
	SENT         = "sent"
)
