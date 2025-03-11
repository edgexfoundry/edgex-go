/*
	Copyright NetFoundry Inc.

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

package channel

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/michaelquigley/pfxlog"
	"github.com/pkg/errors"
	"io"
	"time"
)

/**
 * Message headers notes
 * 0-127 reserved for channel
 * 128-255 reserved for headers that need to be reflected back to sender on responses
 * 128 is used for a message UUID for tracing
 * 1000-1099 reserved for edge messages
 * 1100-1199 is reserved for control plane messages
 * 2000-2500 is reserved for xgress messages
 * 2000-2255 is reserved for xgress implementation headers
 */
const (
	ConnectionIdHeader              = 0
	ReplyForHeader                  = 1
	ResultSuccessHeader             = 2
	HelloRouterAdvertisementsHeader = 3
	HelloVersionHeader              = 4
	HeartbeatHeader                 = 5
	HeartbeatResponseHeader         = 6
	TypeHeader                      = 7
	IdHeader                        = 8

	// Headers in the range 128-255 inclusive will be reflected when creating replies
	ReflectedHeaderBitMask = 1 << 7
	MaxReflectedHeader     = (1 << 8) - 1
)

const magicLength = 4

type readFunction func(io.Reader) (*Message, error)
type marshalFunction func(m *Message) ([]byte, error)

type stringError string

func (s stringError) Error() string {
	return string(s)
}

const BadMagicNumberError = stringError("protocol error: invalid header")

type MessageHeader struct {
	ContentType int32
	sequence    int32
	replyFor    *int32
	Headers     Headers
}

func (header *MessageHeader) Sequence() int32 {
	return header.sequence
}

func (header *MessageHeader) cacheReplyFor() {
	if header.replyFor == nil {
		replyFor, found := header.Headers[ReplyForHeader]
		if found {
			if len(replyFor) != 4 {
				pfxlog.Logger().Warnf("incorrect replyFor encoding. length should be 4 not %v", len(replyFor))
			} else {
				val := int32(binary.LittleEndian.Uint32(replyFor))
				header.replyFor = &val
			}
		}

		if replyFor == nil {
			val := int32(-1)
			header.replyFor = &val
		}
	}
}

func (header *MessageHeader) ReplyFor() int32 {
	header.cacheReplyFor()
	return *header.replyFor
}

func (header *MessageHeader) IsReply() bool {
	header.cacheReplyFor()
	return *header.replyFor != -1
}

func (header *MessageHeader) IsReplyingTo(sequence int32) bool {
	header.cacheReplyFor()
	return *header.replyFor == sequence
}

func (header *MessageHeader) PutUint64Header(key int32, value uint64) {
	encoded := make([]byte, 8)
	binary.LittleEndian.PutUint64(encoded, value)
	header.Headers[key] = encoded
}

func (header *MessageHeader) GetUint64Header(key int32) (uint64, bool) {
	return header.Headers.GetUint64Header(key)
}

func (header *MessageHeader) PutUint32Header(key int32, value uint32) {
	header.Headers.PutUint32Header(key, value)
}

func (header *MessageHeader) GetUint32Header(key int32) (uint32, bool) {
	return header.Headers.GetUint32Header(key)
}

func (header *MessageHeader) PutUint16Header(key int32, value uint16) {
	header.Headers.PutUint16Header(key, value)
}

func (header *MessageHeader) GetUint16Header(key int32) (uint16, bool) {
	return header.Headers.GetUint16Header(key)
}

func (header *MessageHeader) PutByteHeader(key int32, value byte) {
	header.Headers.PutByteHeader(key, value)
}

func (header *MessageHeader) PutStringHeader(key int32, value string) {
	header.Headers.PutStringHeader(key, value)
}

func (header *MessageHeader) GetByteHeader(key int32) (byte, bool) {
	return header.Headers.GetByteHeader(key)
}

func (header *MessageHeader) PutBoolHeader(key int32, value bool) {
	header.Headers.PutBoolHeader(key, value)
}

func (header *MessageHeader) GetBoolHeader(key int32) (bool, bool) {
	return header.Headers.GetBoolHeader(key)
}

func (header *MessageHeader) GetStringHeader(key int32) (string, bool) {
	return header.Headers.GetStringHeader(key)
}

func (header *MessageHeader) PutStringSliceHeader(key int32, s []string) {
	header.Headers.PutStringSliceHeader(key, s)
}

func (header *MessageHeader) GetStringSliceHeader(key int32) ([]string, bool, error) {
	return header.Headers.GetStringSliceHeader(key)
}

func (header *MessageHeader) PutU32ToBytesMapHeader(key int32, m map[uint32][]byte) {
	header.Headers.PutU32ToBytesMapHeader(key, m)
}

func (header *MessageHeader) GetU32ToBytesMapHeader(key int32) (map[uint32][]byte, bool, error) {
	return header.Headers.GetU32ToBytesMapHeader(key)
}

func (header *MessageHeader) PutStringToStringMapHeader(key int32, m map[string]string) {
	header.Headers.PutStringToStringMapHeader(key, m)
}

func (header *MessageHeader) GetStringToStringMapHeader(key int32) (map[string]string, bool, error) {
	return header.Headers.GetStringToStringMapHeader(key)
}

type Headers map[int32][]byte

func (self Headers) PutUint64Header(key int32, value uint64) {
	encoded := make([]byte, 8)
	binary.LittleEndian.PutUint64(encoded, value)
	self[key] = encoded
}

func (self Headers) GetUint64Header(key int32) (uint64, bool) {
	encoded, ok := self[key]
	if !ok || len(encoded) != 8 {
		return 0, ok
	}
	result := binary.LittleEndian.Uint64(encoded)
	return result, true
}

func (self Headers) PutUint32Header(key int32, value uint32) {
	encoded := make([]byte, 4)
	binary.LittleEndian.PutUint32(encoded, value)
	self[key] = encoded
}

func (self Headers) GetUint32Header(key int32) (uint32, bool) {
	encoded, ok := self[key]
	if !ok || len(encoded) != 4 {
		return 0, false
	}
	result := binary.LittleEndian.Uint32(encoded)
	return result, true
}

func (self Headers) PutUint16Header(key int32, value uint16) {
	encoded := make([]byte, 2)
	binary.LittleEndian.PutUint16(encoded, value)
	self[key] = encoded
}

func (self Headers) GetUint16Header(key int32) (uint16, bool) {
	encoded, ok := self[key]
	if !ok || len(encoded) != 2 {
		return 0, false
	}
	result := binary.LittleEndian.Uint16(encoded)
	return result, true
}

func (self Headers) PutByteHeader(key int32, value byte) {
	self[key] = []byte{value}
}

func (self Headers) PutStringHeader(key int32, value string) {
	self[key] = []byte(value)
}

func (self Headers) GetByteHeader(key int32) (byte, bool) {
	encoded, ok := self[key]
	if !ok || len(encoded) < 1 {
		return 0, ok
	}
	return encoded[0], true
}

func (self Headers) PutBoolHeader(key int32, value bool) {
	byteVal := byte(0)
	if value {
		byteVal = 1
	}
	self[key] = []byte{byteVal}
}

func (self Headers) GetBoolHeader(key int32) (bool, bool) {
	encoded, ok := self[key]
	if !ok {
		return false, ok
	}
	result := len(encoded) > 0 && encoded[0] == 1
	return result, true
}

func (self Headers) GetStringHeader(key int32) (string, bool) {
	encoded, ok := self[key]
	return string(encoded), ok
}

func (self Headers) PutStringSliceHeader(key int32, s []string) {
	self[key] = EncodeStringSlice(s)
}

func (self Headers) GetStringSliceHeader(key int32) ([]string, bool, error) {
	encoded, ok := self[key]
	if !ok {
		return nil, false, nil
	}

	v, err := DecodeStringSlice(encoded)
	return v, true, err
}

func (self Headers) PutU32ToBytesMapHeader(key int32, m map[uint32][]byte) {
	self[key] = EncodeU32ToBytesMap(m)
}

func (self Headers) GetU32ToBytesMapHeader(key int32) (map[uint32][]byte, bool, error) {
	encoded, ok := self[key]
	if !ok {
		return nil, false, nil
	}

	v, err := DecodeU32ToBytesMap(encoded)
	return v, true, err
}

func (self Headers) PutStringToStringMapHeader(key int32, m map[string]string) {
	self[key] = EncodeStringToStringMap(m)
}

func (self Headers) GetStringToStringMapHeader(key int32) (map[string]string, bool, error) {
	encoded, ok := self[key]
	if !ok {
		return nil, false, nil
	}

	v, err := DecodeStringToStringMap(encoded)
	return v, true, err
}

func NewMessage(contentType int32, body []byte) *Message {
	return &Message{
		MessageHeader: MessageHeader{
			ContentType: contentType,
			sequence:    -1,
			Headers:     make(map[int32][]byte),
		},
		Body: body,
	}
}

type Message struct {
	MessageHeader
	Body []byte
}

func (m *Message) SetSequence(seq int32) {
	m.sequence = seq
}

func (m *Message) Send(ch Channel) error {
	return ch.Send(m)
}

func (m *Message) ToSendable() Sendable {
	return m
}

func (m *Message) SendListener() SendListener {
	return BaseSendListener{}
}

func (m *Message) ReplyReceiver() ReplyReceiver {
	return nil
}

func (m *Message) Msg() *Message {
	return m
}

func (m *Message) WithPriority(p Priority) Envelope {
	return &priorityEnvelopeImpl{msg: m, p: p}
}

func (m *Message) WithTimeout(duration time.Duration) TimeoutEnvelope {
	ctx, cancelF := context.WithTimeout(context.Background(), duration)
	return &envelopeImpl{
		msg:     m,
		p:       Standard,
		context: ctx,
		cancelF: cancelF,
	}
}

func (m *Message) Context() context.Context {
	return context.Background()
}

func (m *Message) Priority() Priority {
	return Standard
}

func (m *Message) ReplyTo(o *Message) Envelope {
	replyFor := o.sequence
	m.replyFor = &replyFor
	for key, value := range o.Headers {
		if key&ReflectedHeaderBitMask != 0 && key <= MaxReflectedHeader {
			m.Headers[key] = value
		}
	}
	return m
}

func (m *Message) String() string {
	if m.IsReply() {
		return fmt.Sprintf("//ct:[%4d]/sq:[%4d]/rf:[%4d]/l:[%4d]", m.ContentType, m.sequence, m.replyFor, len(m.Body))
	} else {
		return fmt.Sprintf("//ct:[%4d]/sq:[%4d]/rf:[    ]/l:[%4d]", m.ContentType, m.sequence, len(m.Body))
	}
}

var magicUnknownVersion = []byte{0x03, 0x06, 0x09, 0x0a}

const versionLen = 4

/*
 * Channel V2 Wire Format
 *
 *  [ message section ]
 * <marker:[]byte{0x03,0x06,0x09,0x0c}>				0  1  2  3
 * <content-type:int32>                             4  5  6  7
 * <sequence:int32>                                 8  9 10  11
 * <headers-length:int32>							12 13 14 15
 * <body-length:int32>								16 17 18 19
 *
 *  [ data section ]
 * <headers>										20 -> (20 + headers-length)
 * <body>											(20 + headers-length) -> (20 + headers-length + body-length)
 */
var magicV2 = []byte{0x03, 0x06, 0x09, 0x0c}

const dataSectionV2 = 20

var magicV3 = []byte{0x03, 0x06, 0x09, 0x0d}

type UnsupportedVersionError struct {
	supportedVersions []uint32
}

func (u UnsupportedVersionError) Error() string {
	return "server did not support requested channel version"
}

func readHello(peer io.Reader) (*Message, readFunction, marshalFunction, error) {
	version := make([]byte, versionLen)
	read, err := io.ReadFull(peer, version)

	defaultReadF := ReadV2
	defaultMarshalF := MarshalV2

	if err != nil {
		return nil, defaultReadF, defaultMarshalF, err
	}

	if read != versionLen {
		return nil, defaultReadF, defaultMarshalF, errors.New("short read")
	}

	if bytes.Equal(version, magicV2) {
		msg, err := readHelloV2(peer)
		return msg, ReadV2, MarshalV2, err
	}

	return nil, defaultReadF, defaultMarshalF, BadMagicNumberError
}

func readHelloV2(peer io.Reader) (*Message, error) {
	messageSection := make([]byte, dataSectionV2)
	copy(messageSection, magicV2)
	read, err := io.ReadFull(peer, messageSection[versionLen:])
	if err != nil {
		return nil, err
	}
	if read != dataSectionV2-versionLen {
		return nil, errors.New("short read")
	}
	headersLength := readUint32(messageSection[12:16])
	bodyLength := readUint32(messageSection[16:20])
	if headersLength > 4192 || bodyLength > 4192 {
		return nil, fmt.Errorf("hello message too big. header len: %v, body len: %v", headersLength, bodyLength)
	}

	return unmarshalV2(peer, messageSection, headersLength, bodyLength)
}

// ReadV2 reads a V2 message from the given reader and returns the unmarshalled message
func ReadV2(peer io.Reader) (*Message, error) {
	messageSection := make([]byte, dataSectionV2)
	read, err := io.ReadFull(peer, messageSection)
	if err != nil {
		return nil, err
	}

	if read < magicLength {
		return nil, errors.New("short read")
	}

	if !bytes.Equal(messageSection[0:magicLength], magicV2) {
		log := pfxlog.Logger()
		log.Debugf("received message version bytes: %v", messageSection[:4])
		if bytes.Equal(messageSection[0:magicLength], magicUnknownVersion) {
			log.Debug("message appears to be unknown version response")
			return nil, readUnknownVersionResponse(messageSection[4:], peer)
		}
		return nil, BadMagicNumberError
	}

	headersLength := readUint32(messageSection[12:16])
	bodyLength := readUint32(messageSection[16:20])

	return unmarshalV2(peer, messageSection, headersLength, bodyLength)
}

// unmarshalV2 converts a block of V2 wire format data into a *Message.
func unmarshalV2(peer io.Reader, messageSectionData []byte, headersLength, bodyLength uint32) (*Message, error) {
	dataSectionData := make([]byte, headersLength+bodyLength)
	read, err := io.ReadFull(peer, dataSectionData)
	if err != nil {
		return nil, err
	}

	if read != int(headersLength+bodyLength) {
		return nil, errors.New("short read")
	}

	if len(messageSectionData) < dataSectionV2 {
		return nil, errors.New("short data stream")
	}

	if !bytes.Equal(messageSectionData[0:magicLength], magicV2) {
		return nil, errors.New("magic mismatch")
	}

	var headers map[int32][]byte
	if headersLength > 0 {
		headers, err = unmarshalHeaders(dataSectionData[:headersLength])
	} else {
		headers = make(map[int32][]byte)
	}
	if err != nil {
		return nil, err
	}
	m := &Message{
		MessageHeader: MessageHeader{
			ContentType: readInt32(messageSectionData[4:8]),
			sequence:    readInt32(messageSectionData[8:12]),
			Headers:     headers,
		},
		Body: dataSectionData[headersLength:],
	}
	return m, nil
}

/*
 * Channel V1 Headers Wire Format
 *
 * <key:int32> 			0  1  2  3
 * <length:int32>		4  5  6  7
 * <data>				8 -> (8 + length)
 */

func unmarshalHeaders(headerData []byte) (map[int32][]byte, error) {
	out := make(map[int32][]byte)
	if len(headerData) > 0 && len(headerData) < 8 {
		return nil, errors.New("truncated header data")
	}
	i := 0
	for i < len(headerData) {
		if (i + 8) > len(headerData) {
			return nil, fmt.Errorf("short header meta-data (%d >= %d)", i+8, len(headerData))
		}

		key := readInt32(headerData[i : i+4])
		length := readUint32(headerData[i+4 : i+8])
		if (i + 8 + int(length)) > len(headerData) {
			return nil, fmt.Errorf("short header data (%d >= %d)", i+8+int(length), len(headerData))
		}
		data := headerData[i+8 : i+8+int(length)]
		out[key] = data
		i += 8 + int(length)
	}
	return out, nil
}

// MarshalV2 converts a *Message into a block of V2 wire format data.
func MarshalV2(m *Message) ([]byte, error) {
	return marshalWithVersion(m, magicV2)
}

// marshalTest converts a *Message into a block of V3 wire format data.
// this is only here for testing, so we can test selection of an earlier
// supported version
func marshalV3(m *Message) ([]byte, error) {
	return marshalWithVersion(m, magicV3)
}

// marshalWithVersion converts a *Message into a block of V2 wire format data.
func marshalWithVersion(m *Message, version []byte) ([]byte, error) {
	data := new(bytes.Buffer)
	data.Write(version)
	if err := binary.Write(data, binary.LittleEndian, m.ContentType); err != nil { // content-type
		return nil, err
	}
	if err := binary.Write(data, binary.LittleEndian, m.sequence); err != nil { // sequence
		return nil, err
	}
	if m.replyFor != nil {
		m.PutUint32Header(ReplyForHeader, uint32(*m.replyFor))
	}
	headersData, err := marshalHeaders(m.Headers)
	if err != nil {
		return nil, err
	}

	if err := binary.Write(data, binary.LittleEndian, int32(len(headersData))); err != nil { // header-length
		return nil, err
	}
	if err := binary.Write(data, binary.LittleEndian, int32(len(m.Body))); err != nil { // body-length
		return nil, err
	}
	n, err := data.Write(headersData)
	if err != nil {
		return nil, err
	}
	if n != len(headersData) {
		return nil, errors.New("short headers write")
	}
	n, err = data.Write(m.Body)
	if err != nil {
		return nil, err
	}
	if n != len(m.Body) {
		return nil, errors.New("short body write")
	}
	return data.Bytes(), nil
}

func marshalHeaders(headers map[int32][]byte) ([]byte, error) {
	data := new(bytes.Buffer)
	for k, v := range headers {
		if err := binary.Write(data, binary.LittleEndian, k); err != nil {
			return nil, err
		}
		if err := binary.Write(data, binary.LittleEndian, int32(len(v))); err != nil {
			return nil, err
		}
		n, err := data.Write(v)
		if err != nil {
			return nil, err
		}
		if n != len(v) {
			return nil, errors.New("short header write")
		}
	}
	return data.Bytes(), nil
}

// readInt32 pulls a 4-byte int32 out of a byte array (or slice).
func readInt32(data []byte) int32 {
	return int32(binary.LittleEndian.Uint32(data))
}

func readUint32(data []byte) uint32 {
	return binary.LittleEndian.Uint32(data)
}

func WriteUnknownVersionResponse(writer io.Writer) {
	data := new(bytes.Buffer)
	data.Write(magicUnknownVersion)

	for _, val := range []uint32{2, 1, 2} { // 2 versions being sent, version 1 and version 2
		if err := binary.Write(data, binary.LittleEndian, val); err != nil {
			pfxlog.Logger().WithError(err).Warnf("Unable to write value to bytes.Buffer")
			return
		}
	}

	written, err := writer.Write(data.Bytes())
	if err != nil {
		pfxlog.Logger().WithError(err).Warnf("Unable to write unknown message version response")
	} else if written != data.Len() {
		pfxlog.Logger().Warnf("Short write while writing unknown message version response")
	}
}

func readUnknownVersionResponse(initial []byte, reader io.Reader) error {
	log := pfxlog.Logger()
	if len(initial) < 4 {
		log.Debug("didn't receive enough bytes for an unknown version response")
		return errors.New("channel synchronization")
	}
	versionCount := binary.LittleEndian.Uint32(initial)
	buf := initial[4:]
	size := versionCount * 4

	if uint32(len(buf)) < size {
		leftover := buf
		buf := make([]byte, size)
		copy(buf, leftover)
		restBuf := buf[len(leftover):]
		if _, err := io.ReadFull(reader, restBuf); err != nil {
			log.Debugf("unable to read all %v bytes for unknown version response", len(restBuf))
			return errors.New("channel synchronization")
		}
	}

	var supported []uint32

	for len(buf) > 0 {
		version := binary.LittleEndian.Uint32(buf)
		supported = append(supported, version)
		buf = buf[4:]
	}

	return UnsupportedVersionError{supportedVersions: supported}
}

func GetRetryVersion(err error) (uint32, bool) {
	return getRetryVersionFor(err, 2, 2)
}

func getRetryVersionFor(err error, defaultVersion uint32, localVersions ...uint32) (uint32, bool) {
	versionErr, ok := err.(UnsupportedVersionError)
	log := pfxlog.Logger()
	if ok && len(versionErr.supportedVersions) > 0 {
		log.Info("received unsupported version response from server")
		for _, version := range localVersions {
			for _, remoteVersion := range versionErr.supportedVersions {
				if remoteVersion == version {
					log.Infof("using highest supported version %v", version)
					return version, true
				}
			}
		}
	}

	log.Infof("defaulting to version %v", defaultVersion)
	return defaultVersion, false
}

func DecodeString(t string, b []byte) ([]byte, string, error) {
	if len(b) < 4 {
		return nil, "", fmt.Errorf("invalid string in %s, not enough bytes for string length", t)
	}
	sl := binary.LittleEndian.Uint32(b[0:4])
	if sl > 8192 {
		return nil, "", fmt.Errorf("strings in %s may have max length 8192", t)
	}
	b = b[4:]
	if sl > uint32(len(b)) {
		return nil, "", fmt.Errorf("invalid string length in %s, longer than remaining data", t)
	}
	s := string(b[:sl])
	b = b[sl:]
	return b, s, nil
}

func EncodeStringSlice(strSlice []string) []byte {
	if len(strSlice) == 0 {
		return nil
	}
	l := 4 + (4 * len(strSlice))
	for _, s := range strSlice {
		l += len(s)
	}
	result := make([]byte, l)

	binary.LittleEndian.PutUint32(result, uint32(len(strSlice)))
	idx := 4
	for _, s := range strSlice {
		binary.LittleEndian.PutUint32(result[idx:], uint32(len(s)))
		copy(result[idx+4:], s)
		idx += 4 + len(s)
	}

	if idx != l {
		panic(fmt.Errorf("invalid length was %d, expected %d", idx, l))
	}

	return result
}

func DecodeStringSlice(b []byte) ([]string, error) {
	if len(b) == 0 {
		return nil, nil
	}

	if len(b) < 4 {
		return nil, errors.New("invalid string slice, not enough bytes for slice length")
	}

	var result []string
	l := binary.LittleEndian.Uint32(b[0:4])

	if l > 128 {
		return nil, errors.New("string slice may have at most 128 entries")
	}
	b = b[4:]
	var i uint32
	var err error
	for i = 0; i < l; i++ {
		var s string
		b, s, err = DecodeString("string slice", b)
		if err != nil {
			return nil, err
		}
		result = append(result, s)
	}

	return result, nil
}

func EncodeU32ToBytesMap(m map[uint32][]byte) []byte {
	if len(m) == 0 {
		return nil
	}
	l := 4 + (8 * len(m))
	for _, v := range m {
		l += len(v)
	}
	result := make([]byte, l)

	binary.LittleEndian.PutUint32(result, uint32(len(m)))
	idx := 4
	for k, v := range m {
		binary.LittleEndian.PutUint32(result[idx:], k)
		binary.LittleEndian.PutUint32(result[idx+4:], uint32(len(v)))
		copy(result[idx+8:], v)
		idx += 8 + len(v)
	}

	if idx != l {
		panic(fmt.Errorf("invalid length was %d, expected %d", idx, l))
	}

	return result
}

func DecodeU32ToBytesMap(b []byte) (map[uint32][]byte, error) {
	if len(b) == 0 {
		return nil, nil
	}

	if len(b) < 4 {
		return nil, errors.New("invalid uint32->bytes map, not enough bytes for map length")
	}

	result := map[uint32][]byte{}
	l := binary.LittleEndian.Uint32(b[0:4])

	if l > 128 {
		return nil, errors.New("uint32->bytes map may have at most 128 entries")
	}
	b = b[4:]
	var i uint32
	for i = 0; i < l; i++ {
		if len(b) < 4 {
			return nil, errors.New("invalid uint32->bytes map, not enough bytes for key")
		}
		key := binary.LittleEndian.Uint32(b[0:4])
		b = b[4:]
		if len(b) < 4 {
			return nil, errors.New("invalid uint32->bytes map, not enough bytes for entry length")
		}
		entryLen := binary.LittleEndian.Uint32(b[0:4])
		if entryLen > 8192 {
			return nil, errors.New("entries in uint32->bytes map may have max length 8192")
		}
		b = b[4:]
		if entryLen > uint32(len(b)) {
			return nil, errors.New("invalid uint32->bytes map entry length, longer than remaining data")
		}
		if entryLen == 0 {
			result[key] = nil
		} else {
			result[key] = b[:entryLen]
			b = b[entryLen:]
		}
	}

	return result, nil
}

func EncodeStringToStringMap(m map[string]string) []byte {
	if len(m) == 0 {
		return nil
	}
	l := 4 + (8 * len(m))
	for k, v := range m {
		l += len(k) + len(v)
	}
	result := make([]byte, l)

	binary.LittleEndian.PutUint32(result, uint32(len(m)))
	idx := 4
	for k, v := range m {
		binary.LittleEndian.PutUint32(result[idx:], uint32(len(k)))
		copy(result[idx+4:], k)
		idx += 4 + len(k)

		binary.LittleEndian.PutUint32(result[idx:], uint32(len(v)))
		copy(result[idx+4:], v)
		idx += 4 + len(v)
	}

	if idx != l {
		panic(fmt.Errorf("invalid length was %d, expected %d", idx, l))
	}

	return result
}

func DecodeStringToStringMap(b []byte) (map[string]string, error) {
	if len(b) == 0 {
		return nil, nil
	}

	if len(b) < 4 {
		return nil, errors.New("invalid string->string map, not enough bytes for map length")
	}

	result := map[string]string{}
	l := binary.LittleEndian.Uint32(b[0:4])

	if l > 128 {
		return nil, errors.New("string->string map may have at most 128 entries")
	}
	b = b[4:]
	var i uint32
	var err error
	for i = 0; i < l; i++ {
		var key string
		b, key, err = DecodeString("string->string map", b)
		if err != nil {
			return nil, err
		}

		var val string
		b, val, err = DecodeString("string->string map", b)
		if err != nil {
			return nil, err
		}
		result[key] = val
	}

	return result, nil
}

func MarshalV2WithRaw(m *Message) ([]byte, error) {
	if m.ContentType == ContentTypeRaw {
		return m.Body, nil
	}
	return MarshalV2(m)
}
