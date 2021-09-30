//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package channel

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	mail "net/smtp"
	"strconv"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal/support/notifications/config"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

const (
	// secretKeyUsername is the key to read the username from the secret path
	secretKeyUsername = "username"
	// secretKeyPassword is the key to read the password from the secret path
	secretKeyPassword = "password"
)

func buildSmtpMessage(sender string, subject string, toAddresses []string, contentType string, message string) []byte {
	smtpNewline := "\r\n"

	// required CRLF at ends of lines and CRLF between header and body for SMTP RFC 822 style email
	buf := bytes.NewBufferString("Subject: " + subject + smtpNewline)

	buf.WriteString("From: " + sender + smtpNewline)

	buf.WriteString("To: " + strings.Join(toAddresses, ",") + smtpNewline)

	// only add MIME header if notification content type was set
	if contentType != "" {
		buf.WriteString(fmt.Sprintf("MIME-version: 1.0;\r\nContent-Type: %s; charset=\"UTF-8\";\r\n", contentType))
	}

	buf.WriteString(smtpNewline)

	//maximum line size is 1000
	//split on newline first then break further as needed
	for _, line := range strings.Split(message, smtpNewline) {
		ln := 998
		idx := 0
		for len(line) > idx+ln {
			buf.WriteString(line[idx:idx+ln] + smtpNewline)
			idx += ln
		}
		buf.WriteString(line[idx:] + smtpNewline)
	}

	return buf.Bytes()
}

func deduceAuth(dic *di.Container, s config.SmtpInfo) (mail.Auth, errors.EdgeX) {
	lc := container.LoggingClientFrom(dic.Get)
	secretProvider := container.SecretProviderFrom(dic.Get)
	if secretProvider == nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "secret provider is missing. Make sure it is specified to be used in bootstrap.Run()", nil)
	}
	secrets, err := secretProvider.GetSecret(s.SecretPath, secretKeyUsername, secretKeyPassword)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(err), "fail to retrieve the secrets from the secret store", err)
	}
	username, exists := secrets[secretKeyUsername]
	if !exists || username == "" {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "username doesn't exist for SMTP auth", nil)
	}
	password, exists := secrets[secretKeyPassword]
	if !exists || password == "" {
		lc.Debugf("user didn't provide the password, send the email without auth")
		return nil, nil
	}
	return mail.PlainAuth("", username, password, s.Host), nil
}

// sendEmail replicates the functionality provided by the SendMail function from smtp package.
// A revision of standard function was needed because smtp.SendMail
// does not allow for set-reset of InsecureSkipVerify flag of tls.Config structure.
// This flag is needed to be manipulated for allowing the self-signed certificates.
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
func sendEmail(s config.SmtpInfo, auth mail.Auth, to []string, msg []byte) errors.EdgeX {
	addr := s.Host + ":" + strconv.Itoa(s.Port)
	c, err := mail.Dial(addr)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("fail to connected the SMTP server with address %s", addr), err)
	}
	defer c.Close()
	serverName, _, err := net.SplitHostPort(addr)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	if err = c.Hello(addr); err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	if ok, _ := c.Extension("STARTTLS"); ok {
		config := &tls.Config{MinVersion: tls.VersionTLS12, ServerName: serverName}
		config.InsecureSkipVerify = s.EnableSelfSignedCert
		if err = c.StartTLS(config); err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}
	if auth != nil {
		if ok, _ := c.Extension("AUTH"); !ok {
			return errors.NewCommonEdgeX(errors.KindServerError, "SMTP server doesn't support AUTH", nil)
		}
		err = c.Auth(auth)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}
	if err = c.Mail(s.Sender); err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}
	w, err := c.Data()
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	_, err = w.Write(msg)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	err = w.Close()
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	err = c.Quit()
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}
