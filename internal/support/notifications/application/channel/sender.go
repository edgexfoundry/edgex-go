//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package channel

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"
	notificationContainer "github.com/edgexfoundry/edgex-go/internal/support/notifications/container"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
)

// DingdingSender is the implementation of the interfaces.ChannelSender, which is used to send the notifications via REST
// add  by edgeGo
type DingdingSender struct {
	dic *di.Container
}

// AliSmsSender is the implementation of the interfaces.ChannelSender, which is used to send the notifications via REST
// add  by edgeGo
type AliSmsSender struct {
	dic *di.Container
}

// NewSmsSender creates the SmsSender instance
// add  by edgeGo
func NewSmsSender(dic *di.Container) Sender {
	return &AliSmsSender{dic: dic}
}

// NewDingdingSender creates the DingdingSender instance
// add  by edgeGo
func NewDingdingSender(dic *di.Container) Sender {
	return &DingdingSender{dic: dic}
}

// NewRESTSender creates the RESTSender instance
func NewRESTSender(dic *di.Container) Sender {
	return &RESTSender{dic: dic}
}

// Send sms request to the aliyun clound sms service address,add by edgeGo
func (sender *AliSmsSender) Send(notification models.Notification, address models.Address) (res string, err errors.EdgeX) {
	//lc := container.LoggingClientFrom(sender.dic.Get)
	smsInfo := notificationContainer.ConfigurationFrom(sender.dic.Get).SmsInfo

	smsAddress, ok := address.(models.SmsAddress)
	if !ok {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to cast Address to SmsAddress", nil)
	}

	client, e := dysmsapi.NewClientWithAccessKey(smsInfo.Regin, smsInfo.AccessKey, smsInfo.Secret)
	if e != nil {
		return e.Error(), errors.NewCommonEdgeXWrapper(e)
	}

	phoneNumberJson := "["
	for i := 0; i < len(smsAddress.Recipients); i++ {
		if i != len(smsAddress.Recipients)-1 {
			phoneNumberJson += "\"" + smsAddress.Recipients[i] + "\","
		} else {
			phoneNumberJson += "\"" + smsAddress.Recipients[i] + "\"]"
		}
	}

	request := dysmsapi.CreateSendBatchSmsRequest()
	request.Scheme = "https"
	request.PhoneNumberJson = phoneNumberJson
	request.SignNameJson = "EdgeGo"
	request.TemplateCode = "SMS_0000"
	request.TemplateParamJson = ""

	response, e := client.SendBatchSms(request)
	if e != nil {
		return e.Error(), errors.NewCommonEdgeXWrapper(e)
	}

	return response.String(), nil
}

//add by edgeGo
func sign(t int64, secret string) string {
	secStr := fmt.Sprintf("%d\n%s", t, secret)
	hmac256 := hmac.New(sha256.New, []byte(secret))
	hmac256.Write([]byte(secStr))
	result := hmac256.Sum(nil)
	return base64.StdEncoding.EncodeToString(result)
}

// Send dingding request to the webhook url ,add by edgeGo
func (sender *DingdingSender) Send(notification models.Notification, address models.Address) (res string, err errors.EdgeX) {
	dingInfo := notificationContainer.ConfigurationFrom(sender.dic.Get).DingdingInfo

	dingAddress, ok := address.(models.DingdingAddress)
	if !ok {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to cast Address to DingdingAddress", nil)
	}

	timestamp := time.Now().UnixNano() / 1e6
	signStr := sign(timestamp, dingAddress.Secret)
	dingUrl := fmt.Sprintf("%s?access_token=%s&timestamp=%d&sign=%s", dingInfo.Webhook, dingAddress.AccessToken, timestamp, signStr)

	resp, e := http.Post(dingUrl, "application/json", strings.NewReader(string(notification.Content)))
	if e != nil {
		return e.Error(), errors.NewCommonEdgeXWrapper(e)
	}
	defer resp.Body.Close()

	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return e.Error(), errors.NewCommonEdgeXWrapper(e)
	}

	return string(body), nil
}


// Sender abstracts the notification sending via specified channel
type Sender interface {
	Send(notification models.Notification, address models.Address) (res string, err errors.EdgeX)
}

// RESTSender is the implementation of the interfaces.ChannelSender, which is used to send the notifications via REST
type RESTSender struct {
	dic *di.Container
}


// NewRESTSender creates the RESTSender instance
func NewRESTSender(dic *di.Container) Sender {
	return &RESTSender{dic: dic}
}

// Send sends the REST request to the specified address
func (sender *RESTSender) Send(notification models.Notification, address models.Address) (res string, err errors.EdgeX) {
	lc := container.LoggingClientFrom(sender.dic.Get)

	restAddress, ok := address.(models.RESTAddress)
	if !ok {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to cast Address to RESTAddress", nil)
	}
	return utils.SendRequestWithRESTAddress(lc, notification.Content, notification.ContentType, restAddress)
}

// EmailSender is the implementation of the interfaces.ChannelSender, which is used to send the notifications via email
type EmailSender struct {
	dic *di.Container
}

// NewEmailSender creates the EmailSender instance
func NewEmailSender(dic *di.Container) Sender {
	return &EmailSender{dic: dic}
}

// Send sends the email to the specified address
func (sender *EmailSender) Send(notification models.Notification, address models.Address) (res string, err errors.EdgeX) {
	smtpInfo := notificationContainer.ConfigurationFrom(sender.dic.Get).Smtp

	emailAddress, ok := address.(models.EmailAddress)
	if !ok {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to cast Address to EmailAddress", nil)
	}

	msg := buildSmtpMessage(notification.Sender, smtpInfo.Subject, emailAddress.Recipients, notification.ContentType, notification.Content)
	auth, err := deduceAuth(sender.dic, smtpInfo)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}
	err = sendEmail(smtpInfo, auth, emailAddress.Recipients, msg)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}
	return "", nil
}
