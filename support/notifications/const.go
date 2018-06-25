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

type ConfigurationStruct struct {
	ApplicationName               string
	ConsulProfilesActive          string
	HeartBeatTime                 int
	HeartBeatMsg                  string
	AppOpenMsg                    string
	FormatSpecifier               string
	ServicePort                   int
	ServiceTimeout                int
	ServiceAddress                string
	ServiceName                   string
	ConsulHost                    string
	ConsulCheckAddress            string
	ConsulPort                    int
	CheckInterval                 string
	EnableRemoteLogging           bool
	LoggingFile                   string
	LoggingRemoteURL              string
	MongoDBUserName               string
	MongoDBPassword               string
	MongoDatabaseName             string
	MongoDBHost                   string
	MongoDBPort                   int
	MongoDBConnectTimeout         int
	MongoDBMaxWaitTime            int
	MongoDBKeepAlive              bool
	ReadMaxLimit                  int
	ResendLimit                   int
	CleanupDefaultAge             int64
	SchedulerNormalDuration       string
	SchedulerNormalResendDuration string
	SchedulerCriticalResendDelay  int
	SMTPPort                      string
	SMTPHost                      string
	SMTPSender                    string
	SMTPPassword                  string
	SMTPSubject                   string
}

var configuration = ConfigurationStruct{} //  Needs to be initialized before used

var (
	SUPPORTNOTIFICATIONSSERVICENAME = "support-notifications"
	ESCALATIONSUBSCRIPTIONSLUG      = "ESCALATION"
	ESCALATIONPREFIX                = "escalated-"
	ESCALATEDCONTENTNOTICE          = "This notificaiton is escalated by the transmission"
)
