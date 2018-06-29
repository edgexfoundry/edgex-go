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
	"bytes"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/edgexfoundry/edgex-go/support/notifications/models"
)

func sendViaChannel(n models.Notification, c models.Channel, receiver string) {
	loggingClient.Debug("Sending notification: " + n.Slug + ", via channel: " + c.String())
	var tr models.TransmissionRecord
	if c.Type == models.ChannelType(models.Email) {
		tr = smtpSend(n.Content, c.MailAddresses)
	} else {
		tr = restSend(n.Content, c.Url)
	}
	t, err := persistTransmission(tr, n, c, receiver)
	if err == nil {
		handleFailedTransmission(t)
	}
}

func resendViaChannel(t models.Transmission) {
	var tr models.TransmissionRecord
	if t.Channel.Type == models.ChannelType(models.Email) {
		tr = smtpSend(t.Notification.Content, t.Channel.MailAddresses)
	} else {
		tr = restSend(t.Notification.Content, t.Channel.Url)
	}
	t.ResendCount = t.ResendCount + 1
	t.Status = tr.Status
	t.Records = append(t.Records, tr)
	err := dbc.UpdateTransmission(t)
	if err == nil {
		handleFailedTransmission(t)
	}
}

func getTransmissionRecord(msg string, st models.TransmissionStatus) models.TransmissionRecord {
	tr := models.TransmissionRecord{}
	tr.Sent = time.Now().UnixNano() / int64(time.Millisecond)
	tr.Status = st
	tr.Response = msg
	return tr
}

func persistTransmission(tr models.TransmissionRecord, n models.Notification, c models.Channel, rec string) (models.Transmission, error) {
	trx := models.Transmission{Notification: n, Receiver: rec, Channel: c, ResendCount: 0, Status: tr.Status}
	trx.Records = []models.TransmissionRecord{tr}
	_, err := dbc.AddTransmission(&trx)
	if err != nil {
		loggingClient.Error("Transmission cannot be persisted: " + trx.String())
		return trx, err
	}
	return trx, nil
}

func smtpSend(message string, addressees []string) models.TransmissionRecord {
	tr := getTransmissionRecord("SMTP server received", models.Sent)
	buf := bytes.NewBufferString("Subject: " + smtpSubject + "\r\n")
	// required CRLF at ends of lines and CRLF between header and body for SMTP RFC 822 style email
	buf.WriteString("\r\n")
	buf.WriteString(message)
	err := smtp.SendMail(smtpHost+":"+smtpPort,
		smtp.PlainAuth("", smtpSender, smtpPassword, smtpHost),
		smtpSender, addressees, []byte(buf.String()))
	if err != nil {
		loggingClient.Error("Problems sending message to: " + strings.Join(addressees, ",") + ", issue: " + err.Error())
		tr.Status = models.Failed
		tr.Response = err.Error()
		return tr
	}
	return tr

}

func restSend(message string, url string) models.TransmissionRecord {
	tr := getTransmissionRecord("", models.Sent)
	rs, err := http.Post(url, "text/plain", bytes.NewBuffer([]byte(message)))
	if err != nil {
		loggingClient.Error("Problems sending message to: " + url)
		loggingClient.Error("Error indication was:  " + err.Error())
		tr.Status = models.Failed
		tr.Response = err.Error()
		return tr
	}
	tr.Response = "Got response status code: " + rs.Status
	return tr
}

func handleFailedTransmission(t models.Transmission) {
	n := t.Notification
	if t.ResendCount >= resendLimit {
		loggingClient.Error("Too many transmission resend attempts!  Giving up on transmission: " + t.ID.String() + ", for notification: " + n.Slug)
	}
	if t.Status == models.Failed && n.Status != models.Escalated {
		loggingClient.Debug("Handling failed transmission for: " + t.ID.String() + " for notification: " + t.Notification.Slug + ", resends so far: " + strconv.Itoa(t.ResendCount))
		if n.Severity == models.Critical {
			if t.ResendCount < resendLimit {
				time.AfterFunc(time.Second*5, func() { criticalSeverityResend(t) })
			} else {
				escalate(t)
				t.Status = models.Escalated
				dbc.UpdateTransmission(t)
			}
		}
	}
}
