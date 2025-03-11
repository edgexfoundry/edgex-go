package wss

import (
	"fmt"
	"github.com/openziti/identity"
	"github.com/openziti/transport/v2"
	"github.com/pkg/errors"
	"time"
)

type Config struct {
	WriteTimeout      time.Duration
	ReadTimeout       time.Duration
	IdleTimeout       time.Duration
	PongTimeout       time.Duration
	PingInterval      time.Duration
	HandshakeTimeout  time.Duration
	ReadBufferSize    int
	WriteBufferSize   int
	EnableCompression bool
	Identity          identity.Identity
}

func NewDefaultConfig() *Config {
	return &Config{
		WriteTimeout:      transport.DefaultWsWriteTimeout,
		ReadTimeout:       transport.DefaultWsReadTimeout,
		IdleTimeout:       transport.DefaultWsIdleTimeout,
		PongTimeout:       transport.DefaultWsPongTimeout,
		PingInterval:      transport.DefaultWsPingInterval,
		HandshakeTimeout:  transport.DefaultWsHandshakeTimeout,
		ReadBufferSize:    transport.DefaultWsReadBufferSize,
		WriteBufferSize:   transport.DefaultWsWriteBufferSize,
		EnableCompression: transport.DefaultWsEnableCompression,
	}
}

func (self *Config) Load(data map[interface{}]interface{}) error {
	if v, found := data["writeTimeout"]; found {
		if i, ok := v.(int); ok {
			self.WriteTimeout = time.Second * time.Duration(i)
		} else {
			return errors.New("invalid 'writeTimeout' value")
		}
	}
	if v, found := data["readTimeout"]; found {
		if i, ok := v.(int); ok {
			self.ReadTimeout = time.Second * time.Duration(i)
		} else {
			return errors.New("invalid 'readTimeout' value")
		}
	}
	if v, found := data["idleTimeout"]; found {
		if i, ok := v.(int); ok {
			self.IdleTimeout = time.Second * time.Duration(i)
		} else {
			return errors.New("invalid 'idleTimeout' value")
		}
	}
	if v, found := data["pongTimeout"]; found {
		if i, ok := v.(int); ok {
			self.PongTimeout = time.Second * time.Duration(i)
		} else {
			return errors.New("invalid 'pongTimeout' value")
		}
	}
	if v, found := data["pingInterval"]; found {
		if i, ok := v.(int); ok {
			self.PingInterval = time.Second * time.Duration(i)
		} else {
			return errors.New("invalid 'pingInterval' value")
		}
	} else {
		self.PingInterval = transport.DefaultWsPingInterval
	}
	if v, found := data["handshakeTimeout"]; found {
		if i, ok := v.(int); ok {
			self.HandshakeTimeout = time.Second * time.Duration(i)
		} else {
			return errors.New("invalid 'handshakeTimeout' value")
		}
	}
	if v, found := data["readBufferSize"]; found {
		if i, ok := v.(int); ok {
			self.ReadBufferSize = i
		} else {
			return errors.New("invalid 'readBufferSize' value")
		}
	}
	if v, found := data["writeBufferSize"]; found {
		if i, ok := v.(int); ok {
			self.WriteBufferSize = i
		} else {
			return errors.New("invalid 'writeBufferSize' value")
		}
	}
	if v, found := data["enableCompression"]; found {
		if i, ok := v.(bool); ok {
			self.EnableCompression = i
		} else {
			return errors.New("invalid 'enableCompression' value")
		}
	}

	if v, found := data["identity"]; found {
		if identityMap, ok := v.(map[interface{}]interface{}); ok {

			identityConfig, err := identity.NewConfigFromMap(identityMap)

			if err != nil {
				return fmt.Errorf("could not load identity section: %w", err)
			}

			if err = identityConfig.ValidateForServerWithPathContext("transport.wss"); err != nil {
				return fmt.Errorf("could not validate identity section: %w", err)
			}

			if self.Identity, err = identity.LoadIdentity(*identityConfig); err != nil {
				return fmt.Errorf("could not load identity section: %w", err)
			}

		} else {
			return errors.New("invalid identity value")
		}
	}

	return nil
}

func (self *Config) Dump(name string) string {
	out := name + "{\n"
	out += fmt.Sprintf("\t%-30s %d\n", "writeTimeout", self.WriteTimeout)
	out += fmt.Sprintf("\t%-30s %d\n", "readTimeout", self.ReadTimeout)
	out += fmt.Sprintf("\t%-30s %d\n", "idleTimeout", self.IdleTimeout)
	out += fmt.Sprintf("\t%-30s %d\n", "pongTimeout", self.PongTimeout)
	out += fmt.Sprintf("\t%-30s %d\n", "pingInterval", self.PingInterval)
	out += fmt.Sprintf("\t%-30s %d\n", "handshakeTimeout", self.HandshakeTimeout)
	out += fmt.Sprintf("\t%-30s %d\n", "readBufferSize", self.ReadBufferSize)
	out += fmt.Sprintf("\t%-30s %d\n", "writeBufferSize", self.WriteBufferSize)
	out += fmt.Sprintf("\t%-30s %t\n", "enableCompression", self.EnableCompression)
	out += fmt.Sprintf("\t%-30s %s\n", "serverCert", self.Identity.GetConfig().ServerCert)
	out += fmt.Sprintf("\t%-30s %s\n", "key", self.Identity.GetConfig().Key)
	out += fmt.Sprintf("\t%-30s %s\n", "server_key", self.Identity.GetConfig().ServerKey)

	for _, altServerCerts := range self.Identity.GetConfig().AltServerCerts {
		out += fmt.Sprintf("\t%-30s %s\n", "alt_server_certs[%d].serverCert", altServerCerts.ServerCert)
		out += fmt.Sprintf("\t%-30s %s\n", "alt_server_certs[%d].server_key", altServerCerts.ServerKey)
	}

	out += "}"
	return out
}
