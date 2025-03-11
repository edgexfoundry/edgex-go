/*
	Copyright 2019 NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package edge

import (
	"encoding/binary"
	"github.com/openziti/channel/v3"
	"github.com/openziti/foundation/v2/uuidz"
	"github.com/openziti/sdk-golang/pb/edge_client_pb"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	ContentTypeConnect            = int32(edge_client_pb.ContentType_ConnectType)
	ContentTypeStateConnected     = int32(edge_client_pb.ContentType_StateConnectedType)
	ContentTypeStateClosed        = int32(edge_client_pb.ContentType_StateClosedType)
	ContentTypeData               = int32(edge_client_pb.ContentType_DataType)
	ContentTypeDial               = int32(edge_client_pb.ContentType_DialType)
	ContentTypeDialSuccess        = int32(edge_client_pb.ContentType_DialSuccessType)
	ContentTypeDialFailed         = int32(edge_client_pb.ContentType_DialFailedType)
	ContentTypeBind               = int32(edge_client_pb.ContentType_BindType)
	ContentTypeUnbind             = int32(edge_client_pb.ContentType_UnbindType)
	ContentTypeStateSessionEnded  = int32(edge_client_pb.ContentType_StateSessionEndedType)
	ContentTypeProbe              = int32(edge_client_pb.ContentType_ProbeType)
	ContentTypeUpdateBind         = int32(edge_client_pb.ContentType_UpdateBindType)
	ContentTypeHealthEvent        = int32(edge_client_pb.ContentType_HealthEventType)
	ContentTypeTraceRoute         = int32(edge_client_pb.ContentType_TraceRouteType)
	ContentTypeTraceRouteResponse = int32(edge_client_pb.ContentType_TraceRouteResponseType)

	ContentTypeConnInspectRequest  = int32(edge_client_pb.ContentType_ConnInspectRequest)
	ContentTypeConnInspectResponse = int32(edge_client_pb.ContentType_ConnInspectResponse)
	ContentTypeBindSuccess         = int32(edge_client_pb.ContentType_BindSuccess)

	ContentTypeUpdateToken        = int32(edge_client_pb.ContentType_UpdateTokenType)
	ContentTypeUpdateTokenSuccess = int32(edge_client_pb.ContentType_UpdateTokenSuccessType)
	ContentTypeUpdateTokenFailure = int32(edge_client_pb.ContentType_UpdateTokenFailureType)

	ContentTypePostureResponse = int32(edge_client_pb.ContentType_PostureResponseType)
)

const (
	// UUIDHeader is put in the reflected range so replies will share the same UUID
	UUIDHeader = int32(edge_client_pb.HeaderId_UUID)

	ConnIdHeader                   = int32(edge_client_pb.HeaderId_ConnId)
	SeqHeader                      = int32(edge_client_pb.HeaderId_Seq)
	SessionTokenHeader             = int32(edge_client_pb.HeaderId_SessionToken)
	PublicKeyHeader                = int32(edge_client_pb.HeaderId_PublicKey)
	CostHeader                     = int32(edge_client_pb.HeaderId_Cost)
	PrecedenceHeader               = int32(edge_client_pb.HeaderId_Precedence)
	TerminatorIdentityHeader       = int32(edge_client_pb.HeaderId_TerminatorIdentity)
	TerminatorIdentitySecretHeader = int32(edge_client_pb.HeaderId_TerminatorIdentitySecret)
	CallerIdHeader                 = int32(edge_client_pb.HeaderId_CallerId)
	CryptoMethodHeader             = int32(edge_client_pb.HeaderId_CryptoMethod)
	FlagsHeader                    = int32(edge_client_pb.HeaderId_Flags)
	AppDataHeader                  = int32(edge_client_pb.HeaderId_AppData)
	RouterProvidedConnId           = int32(edge_client_pb.HeaderId_RouterProvidedConnId)
	HealthStatusHeader             = int32(edge_client_pb.HeaderId_HealthStatus)
	ErrorCodeHeader                = int32(edge_client_pb.HeaderId_ErrorCode)
	TimestampHeader                = int32(edge_client_pb.HeaderId_Timestamp)
	TraceHopCountHeader            = int32(edge_client_pb.HeaderId_TraceHopCount)
	TraceHopTypeHeader             = int32(edge_client_pb.HeaderId_TraceHopType)
	TraceHopIdHeader               = int32(edge_client_pb.HeaderId_TraceHopId)
	TraceSourceRequestIdHeader     = int32(edge_client_pb.HeaderId_TraceSourceRequestId)
	TraceError                     = int32(edge_client_pb.HeaderId_TraceError)
	ListenerId                     = int32(edge_client_pb.HeaderId_ListenerId)
	ConnTypeHeader                 = int32(edge_client_pb.HeaderId_ConnType)
	SupportsInspectHeader          = int32(edge_client_pb.HeaderId_SupportsInspect)
	SupportsBindSuccessHeader      = int32(edge_client_pb.HeaderId_SupportsBindSuccess)
	ConnectionMarkerHeader         = int32(edge_client_pb.HeaderId_ConnectionMarker)
	CircuitIdHeader                = int32(edge_client_pb.HeaderId_CircuitId)
	StickinessTokenHeader          = int32(edge_client_pb.HeaderId_StickinessToken)
)

const (
	ErrorCodeInternal                    = int32(edge_client_pb.Error_Internal)
	ErrorCodeInvalidApiSession           = int32(edge_client_pb.Error_InvalidApiSession)
	ErrorCodeInvalidSession              = int32(edge_client_pb.Error_InvalidSession)
	ErrorCodeWrongSessionType            = int32(edge_client_pb.Error_WrongSessionType)
	ErrorCodeInvalidEdgeRouterForSession = int32(edge_client_pb.Error_InvalidEdgeRouterForSession)
	ErrorCodeInvalidService              = int32(edge_client_pb.Error_InvalidService)
	ErrorCodeTunnelingNotEnabled         = int32(edge_client_pb.Error_TunnelingNotEnabled)
	ErrorCodeInvalidTerminator           = int32(edge_client_pb.Error_InvalidTerminator)
	ErrorCodeInvalidPrecedence           = int32(edge_client_pb.Error_InvalidPrecedence)
	ErrorCodeInvalidCost                 = int32(edge_client_pb.Error_InvalidCost)
	ErrorCodeEncryptionDataMissing       = int32(edge_client_pb.Error_EncryptionDataMissing)
)

const (
	PrecedenceDefault  = Precedence(edge_client_pb.PrecedenceValue_Default)
	PrecedenceRequired = Precedence(edge_client_pb.PrecedenceValue_Required)
	PrecedenceFailed   = Precedence(edge_client_pb.PrecedenceValue_Failed)
)

const (
	// CryptoMethodLibsodium are used to indicate the crypto engine in use
	CryptoMethodLibsodium CryptoMethod = 0 // default: crypto_kx_*, crypto_secretstream_*
	CryptoMethodSSL       CryptoMethod = 1 // OpenSSL(possibly with FIPS): ECDH, AES256-GCM
)

const (
	// FIN is an edge payload flag used to signal communication ends
	FIN = uint32(edge_client_pb.Flag_FIN)
	// TRACE_UUID indicates that peer will send data messages with specially constructed UUID headers
	TRACE_UUID = uint32(edge_client_pb.Flag_TRACE_UUID)
	// MULTIPART indicates that peer can accept multipart data messages
	MULTIPART = uint32(edge_client_pb.Flag_MULTIPART)
	// STREAM indicates connection with stream semantics
	// this allows consolidation of payloads to lower overhead
	STREAM = uint32(edge_client_pb.Flag_STREAM)
	// MULTIPART_MSG set on data message with multiple payloads
	MULTIPART_MSG = uint32(edge_client_pb.Flag_MULTIPART_MSG)
)

type CryptoMethod byte

type Precedence byte

var ContentTypeValue = map[string]int32{
	"EdgeConnectType":            ContentTypeConnect,
	"EdgeStateConnectedType":     ContentTypeStateConnected,
	"EdgeStateClosedType":        ContentTypeStateClosed,
	"EdgeDataType":               ContentTypeData,
	"EdgeDialType":               ContentTypeDial,
	"EdgeDialSuccessType":        ContentTypeDialSuccess,
	"EdgeDialFailedType":         ContentTypeDialFailed,
	"EdgeBindType":               ContentTypeBind,
	"EdgeUnbindType":             ContentTypeUnbind,
	"EdgeProbeType":              ContentTypeProbe,
	"EdgeUpdateTokenType":        ContentTypeUpdateToken,
	"EdgeUpdateTokenSuccessType": ContentTypeUpdateTokenSuccess,
	"EdgeUpdateTokenFailureType": ContentTypeUpdateTokenFailure,
}

var ContentTypeNames = map[int32]string{
	ContentTypeConnect:            "EdgeConnectType",
	ContentTypeStateConnected:     "EdgeStateConnectedType",
	ContentTypeStateClosed:        "EdgeStateClosedType",
	ContentTypeData:               "EdgeDataType",
	ContentTypeDial:               "EdgeDialType",
	ContentTypeDialSuccess:        "EdgeDialSuccessType",
	ContentTypeDialFailed:         "EdgeDialFailedType",
	ContentTypeBind:               "EdgeBindType",
	ContentTypeUnbind:             "EdgeUnbindType",
	ContentTypeProbe:              "EdgeProbeType",
	ContentTypeUpdateToken:        "EdgeUpdateTokenType",
	ContentTypeUpdateTokenSuccess: "EdgeUpdateTokenSuccessType",
	ContentTypeUpdateTokenFailure: "EdgeUpdateTokenFailureType",
}

type MsgEvent struct {
	ConnId  uint32
	Seq     uint32
	MsgUUID []byte
	Msg     *channel.Message
}

func newMsg(contentType int32, connId uint32, seq uint32, data []byte) *channel.Message {
	msg := channel.NewMessage(contentType, data)
	msg.PutUint32Header(ConnIdHeader, connId)
	msg.PutUint32Header(SeqHeader, seq)
	return msg
}

func NewDataMsg(connId uint32, seq uint32, data []byte) *channel.Message {
	return newMsg(ContentTypeData, connId, seq, data)
}

func NewProbeMsg() *channel.Message {
	return channel.NewMessage(ContentTypeProbe, nil)
}

func NewTraceRouteMsg(connId uint32, hops uint32, timestamp uint64) *channel.Message {
	msg := channel.NewMessage(ContentTypeTraceRoute, nil)
	msg.PutUint32Header(ConnIdHeader, connId)
	msg.PutUint32Header(TraceHopCountHeader, hops)
	msg.PutUint64Header(TimestampHeader, timestamp)
	return msg
}

func NewTraceRouteResponseMsg(connId uint32, hops uint32, timestamp uint64, hopType, hopId string) *channel.Message {
	msg := channel.NewMessage(ContentTypeTraceRouteResponse, nil)
	msg.PutUint32Header(ConnIdHeader, connId)
	msg.PutUint32Header(TraceHopCountHeader, hops)
	msg.PutUint64Header(TimestampHeader, timestamp)
	msg.Headers[TraceHopTypeHeader] = []byte(hopType)
	msg.Headers[TraceHopIdHeader] = []byte(hopId)

	return msg
}

func NewConnInspectResponse(connId uint32, connType ConnType, state string) *channel.Message {
	msg := channel.NewMessage(ContentTypeConnInspectResponse, []byte(state))
	msg.PutUint32Header(ConnIdHeader, connId)
	msg.PutByteHeader(ConnTypeHeader, byte(connType))
	return msg
}

func NewConnectMsg(connId uint32, token string, pubKey []byte, options *DialOptions) *channel.Message {
	msg := newMsg(ContentTypeConnect, connId, 0, []byte(token))
	if pubKey != nil {
		msg.Headers[PublicKeyHeader] = pubKey
		msg.PutByteHeader(CryptoMethodHeader, byte(CryptoMethodLibsodium))
	}

	if options.Identity != "" {
		msg.Headers[TerminatorIdentityHeader] = []byte(options.Identity)
	}
	if options.CallerId != "" {
		msg.Headers[CallerIdHeader] = []byte(options.CallerId)
	}
	if options.AppData != nil {
		msg.Headers[AppDataHeader] = options.AppData
	}
	if options.StickinessToken != nil {
		msg.Headers[StickinessTokenHeader] = options.StickinessToken
	}
	return msg
}

func NewStateConnectedMsg(connId uint32) *channel.Message {
	return newMsg(ContentTypeStateConnected, connId, 0, nil)
}

func NewStateClosedMsg(connId uint32, message string) *channel.Message {
	return newMsg(ContentTypeStateClosed, connId, 0, []byte(message))
}

func NewDialMsg(connId uint32, token string, callerId string) *channel.Message {
	msg := newMsg(ContentTypeDial, connId, 0, []byte(token))
	msg.Headers[CallerIdHeader] = []byte(callerId)
	return msg
}

func NewBindMsg(connId uint32, token string, pubKey []byte, options *ListenOptions) *channel.Message {
	msg := newMsg(ContentTypeBind, connId, 0, []byte(token))
	msg.PutBoolHeader(SupportsInspectHeader, true)
	msg.PutBoolHeader(SupportsBindSuccessHeader, true)

	if pubKey != nil {
		msg.Headers[PublicKeyHeader] = pubKey
		msg.PutByteHeader(CryptoMethodHeader, byte(CryptoMethodLibsodium))
	}

	if options.Cost > 0 {
		costBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(costBytes, options.Cost)
		msg.Headers[CostHeader] = costBytes
	}
	if options.Precedence != PrecedenceDefault {
		msg.PutByteHeader(PrecedenceHeader, byte(options.Precedence))
	}

	if options.Identity != "" {
		msg.PutStringHeader(TerminatorIdentityHeader, options.Identity)

		if options.IdentitySecret != "" {
			msg.PutStringHeader(TerminatorIdentitySecretHeader, options.IdentitySecret)
		}
	}

	if options.ListenerId != "" {
		msg.PutStringHeader(ListenerId, options.ListenerId)
	}

	msg.PutBoolHeader(RouterProvidedConnId, true)
	return msg
}

func NewUnbindMsg(connId uint32, token string) *channel.Message {
	return newMsg(ContentTypeUnbind, connId, 0, []byte(token))
}

func NewUpdateBindMsg(connId uint32, token string, cost *uint16, precedence *Precedence) *channel.Message {
	msg := newMsg(ContentTypeUpdateBind, connId, 0, []byte(token))
	if cost != nil {
		msg.PutUint16Header(CostHeader, *cost)
	}
	if precedence != nil {
		msg.Headers[PrecedenceHeader] = []byte{byte(*precedence)}
	}
	return msg
}

func NewHealthEventMsg(connId uint32, token string, pass bool) *channel.Message {
	msg := newMsg(ContentTypeHealthEvent, connId, 0, []byte(token))
	msg.PutBoolHeader(HealthStatusHeader, pass)
	return msg
}

func NewDialSuccessMsg(connId uint32, newConnId uint32) *channel.Message {
	newConnIdBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(newConnIdBytes, newConnId)
	msg := newMsg(ContentTypeDialSuccess, connId, 0, newConnIdBytes)
	return msg
}

func NewDialFailedMsg(connId uint32, message string) *channel.Message {
	return newMsg(ContentTypeDialFailed, connId, 0, []byte(message))
}

func NewStateSessionEndedMsg(reason string) *channel.Message {
	return newMsg(ContentTypeStateSessionEnded, 0, 0, []byte(reason))
}

// NewUpdateTokenMsg creates a message sent to edge routers to update the token that
// allows the client to stay connection. If the token is not update before the current
// one expires, the connection and all service connections through it will be terminated.
func NewUpdateTokenMsg(token []byte) *channel.Message {
	msg := channel.NewMessage(ContentTypeUpdateToken, token)
	return msg
}

// NewUpdateTokenFailedMsg is returned in response to a token update where the token failed
// validation.
func NewUpdateTokenFailedMsg(err error) *channel.Message {
	msg := channel.NewMessage(ContentTypeUpdateTokenFailure, []byte(err.Error()))
	return msg
}

// NewUpdateTokenSuccessMsg is returned in response to a toke update where the token
// was accepted.
func NewUpdateTokenSuccessMsg() *channel.Message {
	msg := channel.NewMessage(ContentTypeUpdateTokenSuccess, nil)
	return msg
}

type DialResult struct {
	ConnId    uint32
	NewConnId uint32
	Success   bool
	Message   string
}

func UnmarshalDialResult(msg *channel.Message) (*DialResult, error) {
	connId, found := msg.GetUint32Header(ConnIdHeader)
	if !found {
		return nil, errors.Errorf("received edge message with no connection id header")
	}

	if msg.ContentType == ContentTypeDialSuccess {
		if len(msg.Body) != 4 {
			return nil, errors.Errorf("dial success msg improperly formatted. body len: %v", len(msg.Body))
		}
		newConnId := binary.LittleEndian.Uint32(msg.Body)
		return &DialResult{
			ConnId:    connId,
			NewConnId: newConnId,
			Success:   true,
		}, nil
	}

	if msg.ContentType == ContentTypeDialFailed {
		return &DialResult{
			ConnId:  connId,
			Success: false,
			Message: string(msg.Body),
		}, nil
	}

	return nil, errors.Errorf("unexpected response. received %v instead of dial result message", msg.ContentType)
}

func GetLoggerFields(msg *channel.Message) logrus.Fields {
	var msgUUID string
	if id, found := msg.Headers[UUIDHeader]; found {
		msgUUID = uuidz.ToString(id)
	}

	connId, _ := msg.GetUint32Header(ConnIdHeader)
	seq, _ := msg.GetUint32Header(SeqHeader)

	fields := logrus.Fields{
		"connId":  connId,
		"type":    ContentTypeNames[msg.ContentType],
		"chSeq":   msg.Sequence(),
		"edgeSeq": seq,
	}

	if msgUUID != "" {
		fields["uuid"] = msgUUID
	}

	if circuitId, found := msg.GetStringHeader(CircuitIdHeader); found {
		fields["circuitId"] = circuitId
	}

	return fields
}

type ConnType byte

const (
	ConnTypeInvalid ConnType = 0
	ConnTypeDial    ConnType = 1
	ConnTypeBind    ConnType = 2
	ConnTypeUnknown ConnType = 3
)

type InspectResult struct {
	ConnId uint32
	Type   ConnType
	Detail string
}

func UnmarshalInspectResult(msg *channel.Message) (*InspectResult, error) {
	if msg.ContentType == ContentTypeConnInspectResponse {
		connId, _ := msg.GetUint32Header(ConnIdHeader)
		connType, found := msg.GetByteHeader(ConnTypeHeader)
		if !found {
			connType = byte(ConnTypeUnknown)
		}
		return &InspectResult{
			ConnId: connId,
			Type:   ConnType(connType),
			Detail: string(msg.Body),
		}, nil
	}

	return nil, errors.Errorf("unexpected response. received %v instead of inspect result message", msg.ContentType)
}
