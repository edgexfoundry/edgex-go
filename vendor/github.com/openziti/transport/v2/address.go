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

package transport

import (
	"fmt"
	"github.com/openziti/identity"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
	"io"
	"math"
	"net"
	"strings"
	"time"
)

const (
	KeyProxy                    = "proxy"
	KeyProtocol                 = "protocol"
	KeyCachedProxyConfiguration = "cachedProxyConfiguration"

	KeyHandshakeTimeout       = "handshakeTimeout"
	KeyCachedHandshakeTimeout = "cachedHandshakeTimeout"
)

type Configuration map[interface{}]interface{}

// Protocols returns supported or requested application protocols (used for ALPN support)
func (self Configuration) Protocols() []string {
	if self == nil {
		return nil
	}

	p, found := self[KeyProtocol]
	if found {
		switch v := p.(type) {
		case string:
			return []string{v}
		case []string:
			return v
		default:
			panic("invalid transport.Configuration[protocols] type")
		}
	}
	return nil
}

func (self Configuration) GetProxyConfiguration() (*ProxyConfiguration, error) {
	if self == nil {
		return nil, nil
	}

	if val, found := self[KeyCachedProxyConfiguration]; found {
		return val.(*ProxyConfiguration), nil
	}

	val, found := self[KeyProxy]
	if !found {
		return nil, nil
	}

	cfg, ok := val.(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("invalid proxy configuration value, should be map")
	}

	result, err := LoadProxyConfiguration(cfg)
	if err != nil {
		return nil, err
	}

	self[KeyCachedProxyConfiguration] = result

	return result, nil
}

func (self Configuration) GetHandshakeTimeout() (time.Duration, error) {
	if val, ok := self[KeyCachedHandshakeTimeout]; ok {
		if timeout, ok := val.(time.Duration); ok {
			return timeout, nil
		}
	}

	if val, ok := self[KeyHandshakeTimeout]; ok {
		if strVal, ok := val.(string); ok {
			timeout, err := time.ParseDuration(strVal)
			if err == nil {
				self[KeyCachedHandshakeTimeout] = timeout
			}
			return timeout, errors.Wrapf(err, "unable to parse handshake timeout '%s' to duration", strVal)
		} else {
			return 0, errors.New("invalid handshake timeout, must be string value")
		}
	}

	return 0, nil
}

func (self Configuration) GetUIntValue(first string, rest ...string) (uint, bool, error) {
	val, err := self.GetValue(first, rest...)
	if val == nil {
		return 0, false, err
	}

	intVal, ok := val.(int)
	if !ok {
		key := strings.Join(self.toSlice(first, rest...), ":")
		return 0, false, errors.Errorf("value for key %s should be int, not %v of type '%T'", key, val, val)
	}
	if intVal < 0 {
		key := strings.Join(self.toSlice(first, rest...), ":")
		return 0, false, errors.Errorf("value for key %s should be positive, not %v ", key, intVal)
	}

	if intVal > math.MaxInt {
		key := strings.Join(self.toSlice(first, rest...), ":")
		return 0, false, errors.Errorf("value for key %s should be less or equal to %d, not %v ", key, math.MaxInt, intVal)
	}

	return uint(intVal), true, nil
}

func (self Configuration) GetInt64Value(first string, rest ...string) (int64, bool, error) {
	val, err := self.GetValue(first, rest...)
	if val == nil {
		return 0, false, err
	}

	intVal, ok := val.(int)
	if !ok {
		key := strings.Join(self.toSlice(first, rest...), ":")
		return 0, false, errors.Errorf("value for key %s should be int, not %v of type '%T'", key, val, val)
	}
	return int64(intVal), true, nil
}

func (self Configuration) toSlice(first string, rest ...string) []string {
	if len(rest) == 0 {
		return []string{first}
	}

	key := make([]string, len(rest)+1)
	key[0] = first
	copy(key[1:], rest)
	return key
}

func (self Configuration) GetValue(first string, rest ...string) (interface{}, error) {
	return self.getValue(0, self.toSlice(first, rest...)...)
}

func (self Configuration) getValue(index int, key ...string) (interface{}, error) {
	if index == len(key)-1 {
		return self[key[index]], nil
	}

	val, ok := self[key[index]]
	if !ok {
		return nil, nil
	}

	subMap, ok := val.(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid transport configuration value. %s should be map, not %T",
			strings.Join(key[:index+1], ":"), val)
	}

	return Configuration(subMap).getValue(index+1, key...)
}

type ProxyType string

const (
	ProxyTypeNone        ProxyType = "none"
	ProxyTypeHttpConnect ProxyType = "http"
)

type ProxyConfiguration struct {
	Type    ProxyType
	Address string
	Auth    *proxy.Auth
}

func LoadProxyConfiguration(cfg map[interface{}]interface{}) (*ProxyConfiguration, error) {
	val, found := cfg["type"]
	if !found {
		return nil, errors.New("proxy configuration does not specify proxy type")
	}

	proxyType, ok := val.(string)
	if !ok {
		return nil, errors.New("proxy type must be a string")
	}

	if proxyType == string(ProxyTypeNone) {
		return &ProxyConfiguration{
			Type: ProxyTypeNone,
		}, nil
	}

	result := &ProxyConfiguration{}

	switch proxyType {
	case string(ProxyTypeHttpConnect):
		result.Type = ProxyTypeHttpConnect
	default:
		return nil, errors.Errorf("invalid proxy type %s", proxyType)
	}

	val, found = cfg["address"]
	if !found {
		return nil, errors.Errorf("no address specified for %s proxy", string(result.Type))
	}

	if addr, ok := val.(string); !ok {
		return nil, errors.Errorf("invalid value for %s proxy address [%v], must be string", string(result.Type), val)
	} else {
		result.Address = addr
	}

	if val, found = cfg["username"]; found {
		if username, ok := val.(string); ok {
			result.Auth = &proxy.Auth{
				User: username,
			}
		} else {
			return nil, errors.Errorf("invalid value for %s proxy username [%v], must be string", string(result.Type), val)
		}

		if val, found = cfg["password"]; found {
			if password, ok := val.(string); ok {
				result.Auth.Password = password
			} else {
				return nil, errors.Errorf("invalid value for %s proxy password [%v], must be string", string(result.Type), val)
			}
		}
	}

	return result, nil
}

// Address implements the functionality provided by a generic "address".
type Address interface {
	Dial(name string, i *identity.TokenId, timeout time.Duration, tcfg Configuration) (Conn, error)
	DialWithLocalBinding(name string, binding string, i *identity.TokenId, timeout time.Duration, tcfg Configuration) (Conn, error)
	Listen(name string, i *identity.TokenId, acceptF func(Conn), tcfg Configuration) (io.Closer, error)
	MustListen(name string, i *identity.TokenId, acceptF func(Conn), tcfg Configuration) io.Closer
	String() string
	Type() string
}

type HostPortAddress interface {
	Address
	Hostname() string
	Port() uint16
}

// AddressParser implements the functionality provided by an "address parser".
type AddressParser interface {
	Parse(addressString string) (Address, error)
}

// AddAddressParser adds an AddressParser to the globally-configured address parsers.
func AddAddressParser(addressParser AddressParser) {
	for _, e := range addressParsers {
		if addressParser == e {
			return
		}
	}
	addressParsers = append(addressParsers, addressParser)
}

// ParseAddress uses the globally-configured AddressParser instances to parse an address.
func ParseAddress(addressString string) (Address, error) {
	if addressParsers == nil || len(addressParsers) < 1 {
		return nil, errors.New("no configured address parsers")
	}
	for _, addressParser := range addressParsers {
		address, err := addressParser.Parse(addressString)
		if err == nil {
			return address, nil
		}
	}
	return nil, fmt.Errorf("address (%v) not parsed", addressString)
}

// The globally-configured address parsers.
var addressParsers = make([]AddressParser, 0)

// Resolve a network interface by name or IP address
func ResolveInterface(toResolve string) (*net.Interface, error) {
	// Easy check first - see if the interface is specified by name
	ief, err := net.InterfaceByName(toResolve)

	if err == nil {
		return ief, nil
	}

	// Nope! Scan all network interfaces to if there is an IP match
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		if (iface.Flags & net.FlagUp) == 0 {
			log.Debugf("Interface %s is down, ignoring it for address resolution", iface.Name)
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			log.Warnf("Could not check interface %s (%s)", iface.Name, err)
			continue
		}

		for _, addr := range addrs {
			log.Tracef("Checking interface %s (%s) against %s", iface.Name, addr.String(), toResolve)

			var ip net.IP

			switch addr := addr.(type) {
			case *net.IPAddr:
				ip = addr.IP
			case *net.IPNet:
				ip = addr.IP
			default:
				continue
			}

			if ip.To4() != nil && ip.To4().String() == toResolve {
				log.Debugf("Resolved %s to interface %s", toResolve, iface.Name)
				return &iface, nil
			}
		}
	}

	// Not an IP either, not sure how to resolve this interface
	return nil, errors.Errorf("no network interface found for %s", toResolve)
}
