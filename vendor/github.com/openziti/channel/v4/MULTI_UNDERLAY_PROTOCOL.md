# Multi-Underlay Protocol Documentation

## Overview

The multi-underlay protocol is a channel-layer feature that allows a single logical channel to maintain multiple concurrent underlying transport connections (underlays). This provides resilience, load distribution, and the ability to use different transport types simultaneously.

## Core Concepts

### Multi-Channel
A `MultiChannel` represents a single logical communication channel that can contain multiple underlying transport connections. It implements the standard `Channel` interface while internally managing multiple `Underlay` instances.

### Underlay
An `Underlay` represents a single underlying transport connection (TCP, WebSocket, UDP, etc.). Multiple underlays can be grouped together under a single multi-channel using a shared group secret.
 
## Connection ID
A short ID, used to identify which logical channel the connection belongs to.

### Group Secret
A secret generated on the server side when the first connection for a channel is made. It must be presented back to the server when making additional connections. As it is a UUID, it is nearly impossible to guess. Because it's 
generated on first connection, it also changes if a server loses all connection to a channel. Since clients may not lose all connections simultaenously, this also serves as a signal that the client must start afresh, as 
the old group secret won't be recognized and new connections will be rejected.

## Connection Setup Flow

### Initial Connection (First Underlay)

1. **Dialer Initiation**
   - Creates first underlay connection to listener
   - Sends Hello message with required headers

2. **Header Requirements for First Connection**
   ```go
   headers := Headers{
       TypeHeader:             []byte("default"),        // Underlay type identifier
       IsGroupedHeader:        {1},                      // Mark as grouped connection
       IsFirstGroupConnection: {1},                      // Mark as first in group
       ConnectionIdHeader:     []byte(connectionId),     // Unique group identifier
   }
   ```

3. **Listener Processing**
   - Receives Hello message
   - Validates `IsFirstGroupConnection` header
   - Generates group secret if not provided
   - Creates new `MultiChannel` instance
   - Responds with acknowledgment containing group secret

4. **Response Headers**
   ```go
   response.Headers[GroupSecretHeader] = groupSecret    // Generated secret
   response.Headers[IsGroupedHeader]   = {1}           // Confirm grouped
   response.Headers[TypeHeader]        = underlayType  // Echo type
   ```

### Additional Underlays

1. **Subsequent Connections**
   - Use existing group's connection ID and secret
   - Must provide correct group secret for validation

2. **Header Requirements for Additional Connections**
   ```go
   headers := Headers{
       TypeHeader:         []byte(underlayType),     // Type (e.g., "priority")
       ConnectionIdHeader: []byte(groupId),          // Same group ID
       GroupSecretHeader:  groupSecret,              // Shared secret
       IsGroupedHeader:    {1},                      // Mark as grouped
       // Note: IsFirstGroupConnection NOT set
   }
   ```

3. **Validation Process**
   - Listener checks if group exists by connection ID
   - Validates group secret matches
   - Accepts underlay into existing multi-channel
   - Rejects if secret doesn't match or channel is closed

## Required Headers

### Core Headers (message.go:41-52)

| Header | Value | Description | Required When |
|--------|-------|-------------|---------------|
| `ConnectionIdHeader` (0) | string | Unique identifier for the channel group | All grouped connections |
| `TypeHeader` (7) | string | Type of underlay (e.g., "default", "priority") | All connections |
| `IsGroupedHeader` (9) | boolean | Indicates this is a grouped connection | All grouped connections |
| `GroupSecretHeader` (10) | []byte | Cryptographic group identifier | All connections after first |
| `IsFirstGroupConnection` (11) | boolean | Marks the first connection in a group | First connection only |

### Header Usage by Connection Type

#### First Connection
- `ConnectionIdHeader`: Generated or provided connection ID
- `TypeHeader`: Underlay type (typically "default")
- `IsGroupedHeader`: Must be `true`
- `IsFirstGroupConnection`: Must be `true`
- `GroupSecretHeader`: Optional (generated if not provided)

#### Subsequent Connections
- `ConnectionIdHeader`: Same as first connection
- `TypeHeader`: Underlay type (can be different, e.g., "priority")
- `IsGroupedHeader`: Must be `true`
- `GroupSecretHeader`: Must match group secret
- `IsFirstGroupConnection`: Must NOT be set

## Protocol Flow Diagrams

### First Connection Setup
```
Dialer                                  Listener
  |                                        |
  |-- Hello (IsFirstGroupConnection) ----->|
  |                                        |- Generate GroupSecret
  |                                        |- Create MultiChannel
  |<-- Result (GroupSecret) ---------------|
  |                                        |
  |-- Start Message Exchange ------------->|
```

### Additional Connection Setup
```
Dialer                                  Listener
  |                                        |
  |-- Hello (GroupSecret) ---------------->|
  |                                        |- Validate GroupSecret
  |                                        |- Add to existing MultiChannel
  |<-- Result (Confirmed) -----------------|
  |                                        |
  |-- Start Message Exchange ------------->|
```

## Implementation Details

### MultiChannel Creation (multi.go:95-141)
```go
// Configuration for creating a new multi-channel
type MultiChannelConfig struct {
    LogicalName     string
    Options         *Options
    UnderlayHandler UnderlayHandler
    BindHandler     BindHandler
    Underlay        Underlay  // First underlay
}
```

### Underlay Acceptance (multi.go:143-171)
```go
func (mc *multiChannelImpl) AcceptUnderlay(underlay Underlay) error {
    // Validate group secret
    groupSecret := underlay.Headers()[GroupSecretHeader]
    if !bytes.Equal(groupSecret, mc.groupSecret) {
        return fmt.Errorf("incorrect group secret")
    }
    
    // Add to underlay collection
    mc.underlays.Append(underlay)
    mc.startMultiplex(underlay)
    
    return nil
}
```

### Listener Processing (multi_listener.go:34-85)
```go
func (ml *MultiListener) AcceptUnderlay(underlay Underlay) {
    isGrouped, _ := Headers(underlay.Headers()).GetBoolHeader(IsGroupedHeader)
    
    if !isGrouped {
        // Handle non-grouped connection
        return
    }
    
    chId := underlay.ConnectionId()
    
    if existingChannel, exists := ml.channels[chId]; exists {
        // Add to existing multi-channel
        existingChannel.AcceptUnderlay(underlay)
    } else {
        // Create new multi-channel if first connection
        isFirst, _ := Headers(underlay.Headers()).GetBoolHeader(IsFirstGroupConnection)
        if !isFirst {
            // Reject - not first connection but no existing channel
            underlay.Close()
            return
        }
        
        // Create new multi-channel
        mc, err := ml.multiChannelFactory(underlay, closeCallback)
        if err == nil {
            ml.channels[chId] = mc
        }
    }
}
```

## Security Considerations

1. **Group Secret Validation**: All connections must provide the correct group secret
2. **First Connection Protection**: Only connections marked with `IsFirstGroupConnection` can create new groups
3. **Connection Limits**: Implementations may limit the number of underlays per group
4. **Timeout Handling**: Connections have deadlines during the handshake process

## Error Conditions

### Invalid Group Secret
- **Condition**: Provided group secret doesn't match existing group
- **Action**: Close underlay connection immediately
- **Error**: "incorrect group secret"

### No Existing Group for Non-First Connection
- **Condition**: Connection lacks `IsFirstGroupConnection` but no group exists
- **Action**: Close underlay connection
- **Log**: "no existing channel found for underlay, but isFirstGroupConnection not set"

### Channel Already Closed
- **Condition**: Attempting to add underlay to closed multi-channel
- **Action**: Close underlay connection
- **Error**: "multi-channel is closed"

## Testing

The test suite in `multi_test.go` demonstrates:
- Creating multi-channel with initial underlay
- Adding additional underlays of different types
- Message routing across multiple underlays
- Priority handling with different underlay types
- Connection failure and recovery scenarios

## Underlay Types and Constraints

### Underlay Constraints (multi.go:547-643)
```go
type UnderlayConstraints struct {
    types           map[string]underlayConstraint
    applyInProgress atomic.Bool
    lastDial        concurrenz.AtomicValue[time.Time]
}

type underlayConstraint struct {
    numDesired int  // Target number of connections
    minAllowed int  // Minimum required connections
}
```

Example constraint configuration:
```go
constraints.AddConstraint("default", 2, 1)   // Want 2, need at least 1
constraints.AddConstraint("priority", 1, 0)  // Want 1, can have 0
```

## Message Routing

Messages are distributed across underlays based on:
1. **Priority**: Priority messages use priority underlays when available
2. **Load Balancing**: Default messages distributed across available underlays
3. **Retry Logic**: Failed sends are retried on different underlays
4. **Type Affinity**: Different underlay types can handle different message flows

This multi-underlay architecture provides robust, scalable communication with automatic failover and load distribution capabilities.
