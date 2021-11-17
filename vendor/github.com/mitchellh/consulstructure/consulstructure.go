package consulstructure

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	consul "github.com/hashicorp/consul/api"
	"github.com/mitchellh/copystructure"
	"github.com/mitchellh/mapstructure"
)

// Decoder is the structure for decoding Consul data into a Go structure.
//
// Please read the documentation carefully for each field. Fields that
// aren't set properly can result in broken behavior.
//
// See Run for information on how to start the Decoder.
type Decoder struct {
	// Target is the target structure for configuration to be written to.
	// After supplying this configuration option, you should never write to
	// this value again. When configuration changes are detected, Decoder
	// will deep copy this structure before writing any changes and sending
	// them on UpdateCh.
	Target interface{}

	// Prefix is the key prefix in Consul where data will live. Decoder
	// will perform a blocking list query on this prefix and will also
	// request ALL DATA in this prefix. Be very careful that this prefix
	// contains only the configuration data for this application.
	//
	// A prefix of "" is not allowed, since that will request all data in
	// Consul and should never be used as the configuration root.
	//
	// If Prefix doesn't end in '/', then it will be appended. The Decoder
	// treats '/' as the path separator in Consul. This is used to find
	// nested values as well.
	Prefix string

	// UpdateCh is the channel where updates to the configuration are sent.
	// An initial value will be sent on the first read of data from Consul.
	//
	// The value sent on UpdateCh is initially a deep copy of Target so
	// there are no race issues around reading/writing values that come
	// on this channel.
	//
	// ErrCh is sent errors that the Decoder experiences. The Decoder
	// will otherwise ignore errors and continue running in an attempt to
	// stabilize, but you can choose to log or exit on errors if you'd
	// like. Temporary errors such as network issues aren't ever reported.
	UpdateCh chan<- interface{}
	ErrCh    chan<- error

	// QuiescencePeriod is the period of time to wait for Consul changes
	// to quiesce (achieve a stable, unchanging state) before triggering
	// an update.
	//
	// QuiescenceTimeout is the max time to wait for the QuiescencePeriod
	// to be reached before forcing an update anyways. For example, if
	// Period is set to 500ms and Timeout is set to 5s, then if data is
	// continously being updated for over 5s (causing the Period to never
	// be reached), the decoder will trigger an update anyways.
	//
	// If neither of these is set, they will default to 500ms and 5s,
	// respectively.
	QuiescencePeriod  time.Duration
	QuiescenceTimeout time.Duration

	// Consul is the configuration to use for initializing the Consul client.
	// If this is nil, then a default configuration will be used that
	// accesses Consul locally without any ACL token. For default values,
	// see consul.DefaultConfig.
	Consul *consul.Config

	lock   sync.Mutex
	quitCh chan<- struct{}
	doneCh <-chan struct{}
}

// Close stops any running Run method. If none are running, this does
// nothing. Otherwise, this will block until it stops. After Close is
// called, Run can be started again at any time.
//
// This could really be named "Stop" but we have implemented it as Close
// so that it implements io.Closer.
func (d *Decoder) Close() error {
	d.lock.Lock()

	if d.doneCh == nil {
		// No Run is running
		d.lock.Unlock()
		return nil
	}

	if d.quitCh != nil {
		// If we have a quit channel, close it to signal the quit. We
		// then set it to nil so we can never get a double close.
		close(d.quitCh)
		d.quitCh = nil
	}

	// Now just wait for Run to tell us it has exited
	doneCh := d.doneCh
	d.lock.Unlock()
	<-doneCh
	return nil
}

// Run starts the decoder. This should be started in a goroutine. If a
// runner is already running then this will return immediately.
func (d *Decoder) Run() {
	// If we're already running, exit
	d.lock.Lock()
	if d.doneCh != nil {
		d.lock.Unlock()
		return
	}

	// Setup our quit/done channels so we can be signaled to stop
	quitCh := make(chan struct{})
	doneCh := make(chan struct{})
	d.quitCh = quitCh
	d.doneCh = doneCh
	d.lock.Unlock()

	// When we finish, we close the done channel but also set it to
	// nil to signal that there is nothing running.
	defer func() {
		d.lock.Lock()
		defer d.lock.Unlock()

		close(doneCh)
		d.doneCh = nil
	}()

	// Setup the channels
	updateCh := d.UpdateCh
	errCh, ok := d.errCh()
	if !ok {
		// If we didn't have an error channel, then we close this one
		// when we're done since it is temporary.
		defer close(errCh)
	}

	// If we have an empty prefix, it is an error
	if d.Prefix == "" {
		errCh <- errors.New("prefix can't be empty")
		return
	}
	if d.Prefix[len(d.Prefix)-1] != '/' {
		d.Prefix += "/"
	}

	// Qsc settings
	qscPeriod := d.QuiescencePeriod
	qscTimeout := d.QuiescenceTimeout
	if qscPeriod == 0 {
		qscPeriod = 500 * time.Millisecond
	}
	if qscTimeout == 0 {
		qscTimeout = 5 * time.Second
	}

	// Create the Consul client. If we can't create the Consul client
	// then this is an unrecoverable error and we exit.
	config := d.Consul
	if config == nil {
		config = consul.DefaultConfig()
	}
	client, err := consul.NewClient(config)
	if err != nil {
		errCh <- err
		return
	}

	// The first goroutine we run just sits and waits for updated
	// KVPairs from Consul. This keeps track of the ModifyIndex to use.
	// This doesn't trigger any config updating though.
	pairCh := make(chan consul.KVPairs)
	go func() {
		var waitIndex uint64
		for {
			// Setup our variables and query options for the query
			var pairs consul.KVPairs
			var meta *consul.QueryMeta
			queryOpts := &consul.QueryOptions{
				WaitIndex: waitIndex,
				WaitTime:  defaultWaitTime,
			}

			// Perform a query with exponential backoff to get our pairs
			err := backoff.Retry(func() error {
				// If the quitCh is closed, then just return now and
				// don't make anymore queries to Consul.
				select {
				case <-quitCh:
					return nil
				default:
				}

				// Query
				var err error
				pairs, meta, err = client.KV().List(d.Prefix, queryOpts)

				// Been block waiting so first check if quitCh is closed and need to return now
				select {
				case <-quitCh:
					return nil
				default:
				}

				if err != nil {
					errCh <- err
				}

				return err
			}, d.backOff())
			if err != nil {
				// These get sent by list
				continue
			}

			// Check for quit. If so, quit.
			select {
			case <-quitCh:
				return
			default:
			}

			// If we have the same index, then we didn't find any new values.
			if meta.LastIndex == waitIndex {
				continue
			}

			// Update our wait index
			waitIndex = meta.LastIndex

			// Send the pairs
			pairCh <- pairs
		}
	}()

	// Listen for pair updates, wait the proper quiesence periods, and
	// trigger configuration updates.
	init := false
	var pairs consul.KVPairs
	var qscPeriodCh, qscTimeoutCh <-chan time.Time
	for {
		select {
		case <-quitCh:
			// Exit
			return
		case pairs = <-pairCh:
			// Setup our qsc timers and reloop
			qscPeriodCh = time.After(qscPeriod)
			if qscTimeoutCh == nil {
				qscTimeoutCh = time.After(qscTimeout)
			}

			// If we've initialized already, then we wait for qsc.
			// Otherwise, we go through for the initial config.
			if init {
				continue
			}

			init = true
		case <-qscPeriodCh:
		case <-qscTimeoutCh:
		}

		// Set our timers to nil for the next data
		qscPeriodCh = nil
		qscTimeoutCh = nil

		// Decode and send
		if err := d.decode(updateCh, pairs); err != nil {
			errCh <- err
		}
	}
}

func (d *Decoder) decode(ch chan<- interface{}, pairs consul.KVPairs) error {
	raw := make(map[string]interface{})
	for _, p := range pairs {
		// Trim the prefix off our key first
		key := strings.TrimPrefix(p.Key, d.Prefix)

		// Determine what map we're writing the value to. We split by '/'
		// to determine any sub-maps that need to be created.
		m := raw
		children := strings.Split(key, "/")
		if len(children) > 0 {
			key = children[len(children)-1]
			children = children[:len(children)-1]
			for _, child := range children {
				if m[child] == nil {
					m[child] = make(map[string]interface{})
				}

				subm, ok := m[child].(map[string]interface{})
				if !ok {
					return fmt.Errorf("child is both a data item and dir: %s", child)
				}

				m = subm
			}
		}

		m[key] = string(p.Value)
	}

	// First copy our initial value
	target, err := copystructure.Copy(d.Target)
	if err != nil {
		return err
	}

	// Now decode into it
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           target,
		WeaklyTypedInput: true,
		TagName:          "consul",
	})
	if err != nil {
		return err
	}
	if err := decoder.Decode(raw); err != nil {
		return err
	}

	// Send it
	ch <- target
	return nil
}

func (d *Decoder) errCh() (chan<- error, bool) {
	errCh := d.ErrCh
	ok := true

	// If we have no error channel, create one that we drain.
	if errCh == nil {
		ok = false
		ch := make(chan error)
		errCh = ch
		go func() {
			for range ch {
			}
		}()
	}

	return errCh, ok
}

func (d *Decoder) backOff() backoff.BackOff {
	result := backoff.NewExponentialBackOff()
	result.InitialInterval = 1 * time.Second
	result.MaxInterval = 10 * time.Second
	result.MaxElapsedTime = 0
	return result
}

var (
	// The wait time for us can be quite long.
	defaultWaitTime = 30 * time.Minute
)
