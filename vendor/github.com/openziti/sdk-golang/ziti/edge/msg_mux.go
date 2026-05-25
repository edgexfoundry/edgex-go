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
	"fmt"
	"math"
	"strings"
	"sync/atomic"
	"time"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v4"
	"github.com/openziti/sdk-golang/inspect"
	"github.com/openziti/sdk-golang/xgress"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// MsgSink represents a message handler that can receive and process messages
// for a specific connection in a multiplexed channel. Each MsgSink is associated
// with a unique connection ID and handles the message flow for that logical connection.
//
// MsgSink implementations are responsible for:
//   - Processing incoming messages from the multiplexer
//   - Providing a unique connection ID for routing
//   - Handling connection lifecycle events (setup, teardown)
//   - Managing connection-specific state and sequencing
//   - Storing and retrieving arbitrary context data
//
// Thread Safety: Implementations should be safe for concurrent use, as messages
// may be delivered from multiple goroutines.
type MsgSink[T any] interface {
	// Id returns the unique connection ID that this sink handles.
	// This ID is used by the ConnMux to route messages to the correct sink.
	// The ID must remain constant for the lifetime of the sink.
	//
	// Returns:
	//   uint32 - the connection ID for this message sink
	//
	// Example:
	//   connId := sink.Id()
	//   // Use connId for routing decisions
	Id() uint32

	// AcceptMessage processes an incoming message for this connection.
	// The message is guaranteed to be intended for this sink's connection ID.
	// The sink should handle the message appropriately based on its type and content.
	//
	// Parameters:
	//   msg - the incoming message to process
	//
	// Example:
	//   func (s *myMsgSink) AcceptMessage(msg *channel.Message) {
	//       switch msg.ContentType {
	//       case edge.ContentTypeData:
	//           s.handleData(msg)
	//       case edge.ContentTypeStateClosed:
	//           s.handleClose(msg)
	//       }
	//   }
	AcceptMessage(msg *channel.Message)

	// HandleMuxClose is called when the underlying multiplexer is closing.
	// This gives the sink an opportunity to perform cleanup operations
	// and notify any dependent components that the connection is being terminated.
	//
	// The sink should:
	//   - Release any held resources
	//   - Notify dependent components of the closure
	//   - Complete any pending operations gracefully
	//
	// Returns:
	//   error - any error encountered during cleanup, nil if successful
	//
	// Example:
	//   func (s *myMsgSink) HandleMuxClose() error {
	//       s.cleanup()
	//       return s.conn.Close()
	//   }
	HandleMuxClose() error

	// GetData retrieves arbitrary context data associated with this connection.
	// This allows implementing code to store and retrieve connection-specific
	// context, state, or metadata that may be needed across different operations.
	//
	// The data is connection-scoped and persists for the lifetime of the MsgSink.
	// Returns nil if no data has been set.
	//
	// Returns:
	//  T - the stored context data, or nil if none exists
	//
	// Example:
	//   userData := sink.GetData()
	//   if edgeCtx, ok := userData.(*EdgeConnContext); ok {
	//       // Use context
	//   }
	GetData() T

	// SetData stores arbitrary context data associated with this connection.
	// This allows implementing code to attach connection-specific context,
	// state, or metadata that can be retrieved later during the connection's lifetime.
	//
	// The data should be treated as connection-scoped and will be available
	// until the connection is closed or the data is overwritten.
	//
	// Parameters:
	//   data - arbitrary context data to associate with this connection
	//
	// Example:
	//   authCtx := &AuthContext{UserID: "user123", Permissions: perms}
	//   sink.SetData(authCtx)
	//
	//   // Later retrieve it
	//   if ctx := sink.GetData(); ctx != nil {
	//       edgeCtx := ctx.(*EdgeConnContext)
	//       // Use context
	//   }
	SetData(data T)
}

// ConnMux (Connection Multiplexer) manages multiple logical connections
// over a single channel. It provides message routing and connection lifecycle
// management based on connection IDs (connId).
//
// The multiplexer enables efficient use of transport resources by allowing many
// application-level connections to share a single underlying channel connection.
// Each logical connection is identified by a unique uint32 connection ID.
//
// Thread Safety: Implementations must be safe for concurrent use by multiple goroutines.
type ConnMux[T any] interface {
	// Add registers a message handler for a specific connection.
	// The sink's ID() method determines which connection ID it handles.
	// Returns an error if the connection ID is already registered or if
	// registration fails for any other reason.
	//
	// Example:
	//   conn := &edgeXgressConn{connId: 12345, ...}
	//   err := mux.Add(conn)
	Add(sink MsgSink[T]) error

	// Remove unregisters the specified message handler.
	// This removes the connection from the multiplexer's routing table.
	// The method is idempotent - removing a non-existent sink is not an error.
	//
	// Example:
	//   mux.Remove(conn)
	Remove(sink MsgSink[T])

	// RemoveByConnId removes a connection by its ID from the multiplexer.
	// This is a convenience method that removes the connection without
	// requiring a reference to the original MsgSink.
	// The method is idempotent - removing a non-existent connection ID is not an error.
	//
	// Parameters:
	//   connId - the connection ID to remove
	//
	// Example:
	//   mux.RemoveByConnId(12345)
	RemoveByConnId(connId uint32)

	// Close shuts down the multiplexer and all managed connections.
	// After calling Close, the multiplexer should not accept new connections
	// or route any more messages. All registered message sinks will be notified
	// of the closure.
	//
	// This method should be called when the underlying channel is being closed
	// to ensure proper cleanup of all multiplexed connections.
	Close()

	// GetActiveConnIds returns a slice of all active connection IDs.
	// This method provides visibility into which connections are currently
	// being managed by the multiplexer.
	//
	// Returns:
	//   []uint32 - slice of active connection IDs, may be empty if no connections exist
	//
	// Example:
	//   connIds := mux.GetActiveConnIds()
	//   fmt.Printf("Active connections: %v\n", connIds)
	GetActiveConnIds() []uint32

	// HasConn checks if a specific connection ID is currently active.
	// This is useful for validation before attempting operations on a connection.
	//
	// Parameters:
	//   connId - the connection ID to check
	//
	// Returns:
	//   bool - true if the connection exists, false otherwise
	//
	// Example:
	//   if mux.HasConn(12345) {
	//       // connection exists, safe to send messages
	//   }
	HasConn(connId uint32) bool

	// GetConnCount returns the number of active connections.
	// This provides a quick way to check multiplexer load without
	// allocating a slice of connection IDs.
	//
	// Returns:
	//   int - number of active connections
	//
	// Example:
	//   count := mux.GetConnCount()
	//   if count > maxConnections {
	//       // handle overload condition
	//   }
	GetConnCount() int

	// GetSinks returns a snapshot of all currently active message sinks managed by this multiplexer.
	// This method provides access to the complete set of active connections and their associated
	// message handlers, indexed by their connection IDs.
	//
	// The returned map is a snapshot taken at the time of the call and will not reflect
	// subsequent additions or removals of connections. This method is useful for:
	//   - Administrative operations that need to inspect all active connections
	//   - Debugging and monitoring tools
	//   - Broadcasting operations across all connections
	//   - Connection lifecycle management and cleanup
	//
	// Returns:
	//   map[uint32]MsgSink[T] - a map of connection IDs to their corresponding message sinks
	//
	// Thread Safety: This method is safe for concurrent use.
	//
	// Example:
	//   sinks := mux.GetSinks()
	//   for connId, sink := range sinks {
	//       fmt.Printf("Connection %d: %v\n", connId, sink.GetData())
	//   }
	GetSinks() map[uint32]MsgSink[T]

	// GetNextId generates the next available connection ID for creating new connections.
	// This method ensures that connection IDs are unique within the multiplexer's scope
	// and handles ID allocation automatically. The implementation may use sequential
	// numbering, random generation, or other strategies to avoid collisions.
	//
	// The returned ID is guaranteed to be unique among currently active connections
	// and can be safely used for creating new MsgSink instances.
	//
	// Returns:
	//   uint32 - a unique connection ID that can be used for a new connection
	//
	// Example:
	//   connId := mux.GetNextId()
	//   conn := &myConnection{id: connId}
	//   mux.Add(conn)
	GetNextId() uint32

	// TypedReceiveHandler is an embedded interface
	//
	// TypedReceiveHandler provides typed message handling capabilities for the multiplexer.
	// This allows the ConnMux to be registered as a message handler with the underlying
	// channel infrastructure, enabling automatic message routing based on message types.
	//
	// The embedded interface typically includes methods like:
	//   - ContentType() int32 - returns the message content type this handler processes
	//   - HandleReceive(msg *channel.Message, ch channel.Channel) - processes typed messages
	//
	// This integration allows the ConnMux to participate in the channel's message
	// dispatch system while maintaining its connection multiplexing functionality.
	channel.TypedReceiveHandler

	// CloseHandler is an embedded interface
	//
	// CloseHandler provides automatic cleanup capabilities when the underlying channel closes.
	// This ensures that all multiplexed connections are properly notified and cleaned up
	// when the physical channel connection is terminated.
	//
	// The embedded interface typically includes:
	//   - HandleClose(ch channel.Channel) - called when the channel is closing
	//
	// When the channel closes, the ConnMux should:
	//   - Notify all registered MsgSinks via their HandleMuxClose() method
	//   - Clean up internal routing tables and connection state
	//   - Release any held resources
	//
	// This automatic integration ensures graceful shutdown of all multiplexed
	// connections without requiring manual cleanup by the application.
	channel.CloseHandler
}

func NewChannelConnMapMux[T any](inspectF func() *inspect.ContextInspectResult) ConnMux[T] {
	result := &ConnMuxImpl[T]{
		maxId: (math.MaxUint32 / 2) - 1,
		sinks: cmap.NewWithCustomShardingFunction[uint32, MsgSink[T]](func(key uint32) uint32 {
			return key
		}),
		contextInspector: inspectF,
	}
	return result
}

type ConnMuxImpl[T any] struct {
	closed           atomic.Bool
	sinks            cmap.ConcurrentMap[uint32, MsgSink[T]]
	nextId           uint32
	minId            uint32
	maxId            uint32
	contextInspector func() *inspect.ContextInspectResult
}

func (mux *ConnMuxImpl[T]) GetActiveConnIds() []uint32 {
	return mux.sinks.Keys()
}

func (mux *ConnMuxImpl[T]) HasConn(connId uint32) bool {
	_, found := mux.sinks.Get(connId)
	return found
}

func (mux *ConnMuxImpl[T]) GetConnCount() int {
	return mux.sinks.Count()
}

func (mux *ConnMuxImpl[T]) GetSinks() map[uint32]MsgSink[T] {
	return mux.sinks.Items()
}

func (mux *ConnMuxImpl[T]) GetNextId() uint32 {
	nextId := atomic.AddUint32(&mux.nextId, 1)
	for {
		if _, found := mux.sinks.Get(nextId); found {
			// if it's in use, try next one
			nextId = atomic.AddUint32(&mux.nextId, 1)
		} else if nextId < mux.minId || nextId >= mux.maxId {
			// it's not in use, but not in the valid range, so reset to beginning of range
			atomic.StoreUint32(&mux.nextId, mux.minId)
			nextId = atomic.AddUint32(&mux.nextId, 1)
		} else {
			// If it's not in use, and in the valid range, return it
			return nextId
		}
	}
}

func (mux *ConnMuxImpl[T]) ContentType() int32 {
	return ContentTypeData
}

func (mux *ConnMuxImpl[T]) HandleReceive(msg *channel.Message, ch channel.Channel) {
	connId, found := msg.GetUint32Header(ConnIdHeader)
	if !found {
		if msg.ContentType == ContentTypeInspectRequest {
			go mux.HandleInspect(msg, ch)
			return
		}
		pfxlog.Logger().Errorf("received edge message with no connId header. content type: %v", msg.ContentType)
		return
	}

	if sink, found := mux.sinks.Get(connId); found {
		sink.AcceptMessage(msg)
	} else if msg.ContentType == ContentTypeConnInspectRequest {
		go mux.HandleNotFoundConnInspect(connId, msg, ch)
	} else if msg.ContentType == ContentTypeXgPayload {
		mux.handlePayloadWithNoSink(msg, ch)
	} else if msg.ContentType == ContentTypeStateClosed {
		// ignore, as conn is already closed
	} else {
		pfxlog.Logger().WithField("connId", connId).WithField("contentType", msg.ContentType).
			Info("unable to dispatch msg received for unknown edge conn id")
	}
}

func (mux *ConnMuxImpl[T]) handlePayloadWithNoSink(msg *channel.Message, ch channel.Channel) {
	connId, _ := msg.GetUint32Header(ConnIdHeader)
	payload, err := xgress.UnmarshallPayload(msg)
	if err == nil {
		if (payload.IsCircuitEndFlagSet() || payload.IsFlagEOFSet()) && len(payload.Data) == 0 {
			ack := xgress.NewAcknowledgement(payload.CircuitId, payload.GetOriginator().Invert())
			ackMsg := ack.Marshall()
			ackMsg.PutUint32Header(ConnIdHeader, connId)
			_, _ = ch.TrySend(msg)
		} else {
			pfxlog.Logger().WithField("connId", int(connId)).WithField("circuitId", payload.CircuitId).
				Debug("unable to dispatch xg payload received for unknown edge conn id")
		}
	} else {
		pfxlog.Logger().WithError(err).WithField("connId", int(connId)).
			Debug("unable to dispatch xg payload received for unknown edge conn id")
	}
}

func (mux *ConnMuxImpl[T]) HandleNotFoundConnInspect(connId uint32, msg *channel.Message, ch channel.Channel) {
	pfxlog.Logger().WithField("connId", int(connId)).Info("no conn found for connection inspect")
	resp := NewConnInspectResponse(connId, ConnTypeInvalid, fmt.Sprintf("invalid conn id [%v]", connId))
	if err := resp.ReplyTo(msg).Send(ch); err != nil {
		logrus.WithFields(GetLoggerFields(msg)).WithError(err).
			Error("failed to send inspect response")
	}
}

func (mux *ConnMuxImpl[T]) HandleInspect(msg *channel.Message, ch channel.Channel) {
	resp := &inspect.SdkInspectResponse{
		Success: true,
		Values:  make(map[string]any),
	}
	requestedValues, _, err := msg.GetStringSliceHeader(InspectRequestValuesHeader)
	if err != nil {
		resp.Errors = append(resp.Errors, err.Error())
		resp.Success = false
		mux.returnInspectResponse(msg, ch, resp)
		return
	}

	for _, requested := range requestedValues {
		lc := strings.ToLower(requested)
		if lc == "circuits" {
			circuitsDetail := &xgress.CircuitsDetail{
				Circuits: make(map[string]*xgress.CircuitDetail),
			}

			for _, sink := range mux.sinks.Items() {
				if circuitInfoSrc, ok := sink.(interface {
					GetCircuitDetail() *xgress.CircuitDetail
				}); ok {
					circuitDetail := circuitInfoSrc.GetCircuitDetail()
					if circuitDetail != nil {
						circuitsDetail.Circuits[circuitDetail.CircuitId] = circuitDetail
					}
				}
			}
			resp.Values[requested] = circuitsDetail
		} else if lc == "context" && mux.contextInspector != nil {
			resp.Values[requested] = mux.contextInspector()
		}
	}

	mux.returnInspectResponse(msg, ch, resp)
}

func (mux *ConnMuxImpl[T]) returnInspectResponse(msg *channel.Message, ch channel.Channel, resp *inspect.SdkInspectResponse) {
	var sender channel.Sender = ch
	if mc, ok := ch.(channel.MultiChannel); ok {
		if sdkChan, ok := mc.GetUnderlayHandler().(SdkChannel); ok {
			sender = sdkChan.GetControlSender()
		}
	}

	reply, err := NewInspectResponse(0, resp)
	if err != nil {
		pfxlog.Logger().WithError(err).Error("failed to create inspect response")
		return
	}
	reply.ReplyTo(msg)

	if err = reply.WithTimeout(5 * time.Second).Send(sender); err != nil {
		pfxlog.Logger().WithError(err).Error("failed to send inspect response")
	}
}

func (mux *ConnMuxImpl[T]) HandleClose(channel.Channel) {
	mux.Close()
}

func (mux *ConnMuxImpl[T]) Add(sink MsgSink[T]) error {
	if mux.closed.Load() {
		return errors.Errorf("mux is closed, can't add sink with id [%v]", sink.Id())
	}

	if !mux.sinks.SetIfAbsent(sink.Id(), sink) {
		return errors.Errorf("sink id %v already in use", sink.Id())
	}
	return nil
}

func (mux *ConnMuxImpl[T]) Remove(sink MsgSink[T]) {
	mux.RemoveByConnId(sink.Id())
}

func (mux *ConnMuxImpl[T]) RemoveByConnId(connId uint32) {
	mux.sinks.Remove(connId)
}

func (mux *ConnMuxImpl[T]) Close() {
	if mux.closed.CompareAndSwap(false, true) {
		// we don't need to lock the mux because due to the atomic bool, only one go-routine will enter this.
		// If the sink HandleMuxClose methods do anything with the mux, like remove themselves, they will acquire
		// their own locks
		sinks := mux.sinks.Items()
		for _, val := range sinks {
			if err := val.HandleMuxClose(); err != nil {
				pfxlog.Logger().
					WithField("connId", val.Id()).
					WithError(err).
					Error("error while closing message sink")
			}
		}
	}
}
