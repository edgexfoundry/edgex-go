/*******************************************************************************
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

import (
	"strconv"

	"github.com/edgexfoundry/edgex-go/support/notifications/models"
)

func startNormalResend() error {
	loggingClient.Info("Normal severity resend scheduler is triggered.")
	trxs, err := dbc.TransmissionsByStatus(resendLimit, models.TransmissionStatus(models.Failed))
	if err != nil {
		loggingClient.Error("Normal resend failed: unable to get FAILED transmissions")
	}
	for _, t := range trxs {
		resend(t)
	}
	loggingClient.Debug("Processed " + strconv.Itoa(len(trxs)) + " resend transmissions")
	loggingClient.Info("Normal severity resend scheduler has completed.")
	return nil
}
