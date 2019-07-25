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
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	mail "net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

func sendViaChannel(n models.Notification, c models.Channel, receiver string) {
	LoggingClient.Debug("Sending notification: " + n.Slug + ", via channel: " + c.String())
	var tr models.TransmissionRecord
	if c.Type == models.ChannelType(models.Email) {
		tr = sendMail(n.Content, c.MailAddresses)
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
		tr = sendMail(t.Notification.Content, t.Channel.MailAddresses)
	} else {
		tr = restSend(t.Notification.Content, t.Channel.Url)
	}
	t.ResendCount = t.ResendCount + 1
	t.Status = tr.Status
	t.Records = append(t.Records, tr)
	err := dbClient.UpdateTransmission(t)
	if err == nil {
		handleFailedTransmission(t)
	}
}

func getTransmissionRecord(msg string, st models.TransmissionStatus) models.TransmissionRecord {
	tr := models.TransmissionRecord{}
	tr.Sent = db.MakeTimestamp()
	tr.Status = st
	tr.Response = msg
	return tr
}

func persistTransmission(tr models.TransmissionRecord, n models.Notification, c models.Channel, rec string) (models.Transmission, error) {
	trx := models.Transmission{Notification: n, Receiver: rec, Channel: c, ResendCount: 0, Status: tr.Status}
	trx.Records = []models.TransmissionRecord{tr}
	id, err := dbClient.AddTransmission(trx)
	if err != nil {
		LoggingClient.Error("Transmission cannot be persisted: " + trx.String())
		return trx, err
	}

	//We need to fetch this transmission for later use in retries, otherwise timestamp information will be lost.
	trx, err = dbClient.GetTransmissionById(id)
	if err != nil {
		LoggingClient.Error("error fetching newly saved transmission: " + id)
		return models.Transmission{}, err
	}
	return trx, nil
}

func sendMail(message string, addressees []string) models.TransmissionRecord {
	smtp := Configuration.Smtp
	tr := getTransmissionRecord("SMTP server received", models.Sent)
	buf := bytes.NewBufferString("Subject: " + smtp.Subject + "\r\n")
	// required CRLF at ends of lines and CRLF between header and body for SMTP RFC 822 style email
	buf.WriteString("\r\n")
	buf.WriteString(message)
	err := smtpSend(addressees, []byte(buf.String()), smtp)
	if err != nil {
		LoggingClient.Error("Problems sending message to: " + strings.Join(addressees, ",") + ", issue: " + err.Error())
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
		LoggingClient.Error("Problems sending message to: " + url)
		LoggingClient.Error("Error indication was:  " + err.Error())
		tr.Status = models.Failed
		tr.Response = err.Error()
		return tr
	}
	tr.Response = "Got response status code: " + rs.Status
	return tr
}

func handleFailedTransmission(t models.Transmission) {
	n := t.Notification
	if t.ResendCount >= Configuration.Writable.ResendLimit {
		LoggingClient.Error("Too many transmission resend attempts!  Giving up on transmission: " + t.ID + ", for notification: " + n.Slug)
	}
	if t.Status == models.Failed && n.Status != models.Escalated {
		LoggingClient.Debug("Handling failed transmission for: " + t.ID + " for notification: " + t.Notification.Slug + ", resends so far: " + strconv.Itoa(t.ResendCount))
		if n.Severity == models.Critical {
			if t.ResendCount < Configuration.Writable.ResendLimit {
				time.AfterFunc(time.Second*5, func() { criticalSeverityResend(t) })
			} else {
				escalate(t)
				t.Status = models.Trxescalated
				dbClient.UpdateTransmission(t)
			}
		}
	}
}

func deduceAuth(s SmtpInfo) (mail.Auth, error) {
	if s.Username == "" && s.Password == "" {
		return nil, nil
	}
	if s.Username != "" && s.Password == "" {
		return nil, errors.New("Notifications: Expecting password")
	}
	if s.Username == "" && s.Password != "" {
		return nil, errors.New("Notifications: Expecting username")
	}
	return mail.PlainAuth("", s.Username, s.Password, s.Host), nil
}

// The function smtpSend replicates the functionality provided by the SendMail function
// from smtp package. A rivision of standard function was needed because smtp.SendMail
// does not allow for set-reset of InsecureSkipVerify flag of tls.Config structure. This
// flag is needed to be manipulated for allowing the self-signed certificates.
//
// As it is replicating the functionality from smtp.SendMail, it borrows heavily from the
// original function in its design and implementation. This version adds new functionality
// for handling the SmtpInfo configuration and authentication management, along with the
// requirement of ability to set-reset the InsecureSkipVerify flag.
//
// This is using a lot of unexported methods and types from smtp package through exported
// interfaces, which makes it a little bit trickier to modify. Since, the intention for
// this function is to use it as a support function for handling the low level SMTP
// protocol mechanism, it is not exported.
//
// The addresses in the to parameter are the SMTP RCPT addresses.
//
// The msg parameter should be an RFC 822-style email with headers
// first, a blank line, and then the message body. The lines of msg
// should be CRLF terminated. The msg headers should usually include
// fields such as "From", "To", "Subject", and "Cc".  Sending "Bcc"
// messages is accomplished by including an email address in the to
// parameter but not including it in the msg headers.
//
// The smtpSend provide no support for DKIM signing, MIME
// attachments (see the mime/multipart package), or other mail
// functionality.

func smtpSend(to []string, msg []byte, s SmtpInfo) error {
	addr := s.Host + ":" + strconv.Itoa(s.Port)
	auth, err := deduceAuth(s)
	if err != nil {
		return err
	}
	c, err := mail.Dial(addr)
	defer c.Close()
	if err != nil {
		return errors.New("Notifications: Error dialing address")
	}
	serverName, _, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	if err = c.Hello(addr); err != nil {
		return err
	}
	if ok, _ := c.Extension("STARTTLS"); ok {
		config := &tls.Config{ServerName: serverName}
		config.InsecureSkipVerify = s.EnableSelfSignedCert
		if err = c.StartTLS(config); err != nil {
			return err
		}
	}
	if auth != nil {
		if ok, _ := c.Extension("AUTH"); !ok {
			return errors.New("Notifications: server doesn't support AUTH")
		}
		err = c.Auth(auth)
		if err != nil {
			return err
		}
	}
	if err = c.Mail(s.Sender); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}
