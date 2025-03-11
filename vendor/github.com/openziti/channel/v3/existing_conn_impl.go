package channel

import (
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/openziti/identity"
	"net"
	"sync"
	"time"
)

type existingConnImpl struct {
	peer         net.Conn
	id           *identity.TokenId
	connectionId string
	headers      map[int32][]byte
	closeLock    sync.Mutex
	closed       bool
	readF        readFunction
	marshalF     marshalFunction
}

func (impl *existingConnImpl) GetLocalAddr() net.Addr {
	return impl.peer.LocalAddr()
}

func (impl *existingConnImpl) GetRemoteAddr() net.Addr {
	return impl.peer.RemoteAddr()
}

func (impl *existingConnImpl) SetWriteTimeout(duration time.Duration) error {
	return impl.peer.SetWriteDeadline(time.Now().Add(duration))
}

func (impl *existingConnImpl) SetWriteDeadline(deadline time.Time) error {
	return impl.peer.SetWriteDeadline(deadline)
}

func (impl *existingConnImpl) rxHello() (*Message, error) {
	msg, readF, marshallF, err := readHello(impl.peer)
	impl.readF = readF
	impl.marshalF = marshallF
	return msg, err
}

func (impl *existingConnImpl) Rx() (*Message, error) {
	if impl.closed {
		return nil, errors.New("underlay closed")
	}
	return impl.readF(impl.peer)
}

func (impl *existingConnImpl) Tx(m *Message) error {
	if impl.closed {
		return errors.New("underlay closed")
	}

	data, err := impl.marshalF(m)
	if err != nil {
		return err
	}

	_, err = impl.peer.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (impl *existingConnImpl) Id() string {
	return impl.id.Token
}

func (impl *existingConnImpl) Headers() map[int32][]byte {
	return impl.headers
}

func (impl *existingConnImpl) LogicalName() string {
	return "existing"
}

func (impl *existingConnImpl) ConnectionId() string {
	return impl.connectionId
}

func (impl *existingConnImpl) Certificates() []*x509.Certificate {
	return nil
}

func (impl *existingConnImpl) Label() string {
	return fmt.Sprintf("u{%s}->i{%s}", impl.LogicalName(), impl.ConnectionId())
}

func (impl *existingConnImpl) Close() error {
	impl.closeLock.Lock()
	defer impl.closeLock.Unlock()

	if !impl.closed {
		impl.closed = true
		return impl.peer.Close()
	}
	return nil
}

func (impl *existingConnImpl) IsClosed() bool {
	return impl.closed
}

func newExistingImpl(peer net.Conn, version uint32) *existingConnImpl {
	readF := ReadV2
	marshalF := MarshalV2

	if version == 2 {
		readF = ReadV2
		marshalF = MarshalV2
	} else if version == 3 { // currently only used for testing fallback to a common protocol version
		readF = ReadV2
		marshalF = marshalV3
	}

	return &existingConnImpl{
		peer:     peer,
		readF:    readF,
		marshalF: marshalF,
	}
}
