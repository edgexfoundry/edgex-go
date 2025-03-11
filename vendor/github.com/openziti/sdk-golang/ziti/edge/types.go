package edge

import (
	"fmt"
	"github.com/michaelquigley/pfxlog"
	"github.com/mitchellh/mapstructure"
	"github.com/openziti/edge-api/rest_model"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
	"net"
	"strings"
)

const InterceptV1 = "intercept.v1"

type InterceptV1Config struct {
	Addresses   []ZitiAddress
	PortRanges  []*PortRange
	Protocols   []string
	SourceIp    *string
	DialOptions *InterceptDialOptions `json:"dialOptions"`
	Service     *rest_model.ServiceDetail
}

// Match returns the matching score of the given target address against this intercept. A negative one (-1) is returned
// if no match is found. If the address is matched, a 32bit integer with upper bits set to the hostname match and lower
// bits to port match.
func (intercept *InterceptV1Config) Match(network, hostname string, port uint16) int {
	if !slices.Contains(intercept.Protocols, network) {
		return -1
	}

	var target any
	ip := net.ParseIP(hostname)
	if len(ip) != 0 {
		target = ip
	} else {
		target = hostname
	}

	addrScore := -1
	for _, address := range intercept.Addresses {
		score := address.Matches(target)
		if score == -1 {
			continue
		}

		if score == 0 {
			addrScore = 0
			break
		}

		if addrScore == -1 || score < addrScore {
			addrScore = score
		}
	}

	if addrScore == -1 {
		return -1
	}

	portScore := -1
	for _, portRange := range intercept.PortRanges {
		score := portRange.Match(port)
		if score == -1 {
			continue
		}

		if score == 0 {
			portScore = 0
			break
		}

		if portScore == -1 || score < portScore {
			portScore = score
		}
	}
	if portScore == -1 {
		return -1
	}

	return int(uint(addrScore)<<16 | (uint(portScore) & 0xFFFF))
}

type ZitiAddress struct {
	cidr   *net.IPNet
	ip     net.IP
	domain DomainName
}

func (self *ZitiAddress) Matches(v any) int {
	if ip, ok := v.(net.IP); ok {
		if self.ip != nil {
			if ip.Equal(self.ip) {
				return 0
			} else {
				return -1
			}
		}

		if self.cidr != nil {
			if self.cidr.Contains(ip) {
				ones, bits := self.cidr.Mask.Size()
				return bits - ones
			} else {
				return -1
			}
		}
	} else if hostname, ok := v.(string); ok {
		return self.domain.Match(strings.ToLower(hostname))
	}

	return -1
}

type DomainName string

func (dn DomainName) Match(hostname string) int {
	if len(dn) == 0 {
		return -1
	}

	if dn[0] == '*' {
		domain := string([]byte(dn)[1:])
		if strings.HasSuffix(hostname, domain) {
			return len(hostname) - len(domain)
		} else {
			return -1
		}
	} else {
		if hostname == string(dn) {
			return 0
		} else {
			return -1
		}
	}
}

type PortRange struct {
	Low  uint16
	High uint16
}

func (pr *PortRange) Match(port uint16) int {
	if pr.Low <= port && port <= pr.High {
		return int(pr.High - pr.Low)
	}
	return -1
}

type InterceptDialOptions struct {
	ConnectTimeoutSeconds *int
	Identity              *string
}

func ParseServiceConfig(service *rest_model.ServiceDetail, configType string, target interface{}) (bool, error) {
	logger := pfxlog.Logger().WithField("serviceId", *service.ID).WithField("serviceName", *service.Name)
	if service.Config == nil {
		logger.Debug("no service configs defined for service")
		return false, nil
	}

	configMap, found := service.Config[configType]
	if !found {
		logger.Debugf("no service config of type %v defined for service", configType)
		return false, nil
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result: target,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.TextUnmarshallerHookFunc(),
			mapstructure.StringToTimeDurationHookFunc()),
	})

	if err != nil {
		logger.WithError(err).Debugf("unable to setup decoder for service configuration for type %v defined for service", configType)
		return true, errors.Wrap(err, "unable to setup decoder for service config structure")
	}

	if err := decoder.Decode(configMap); err != nil {
		logger.WithError(err).Debugf("unable to decode service configuration for type %v defined for service", configType)
		return true, errors.Wrap(err, "unable to decode service config structure")
	}
	return true, nil
}

type ClientConfig struct {
	Protocol string
	Hostname ZitiAddress
	Port     int
}

func (s *ClientConfig) String() string {
	return fmt.Sprintf("%v:%v:%v", s.Protocol, s.Hostname, s.Port)
}

func (self *ClientConfig) ToInterceptV1Config() *InterceptV1Config {

	return &InterceptV1Config{
		Protocols:  []string{"tcp", "udp"},
		Addresses:  []ZitiAddress{self.Hostname},
		PortRanges: []*PortRange{{Low: uint16(self.Port), High: uint16(self.Port)}},
	}
}

func NewZitiAddress(str string) (*ZitiAddress, error) {
	addr := &ZitiAddress{}
	err := addr.UnmarshalText([]byte(str))
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func (self *ZitiAddress) UnmarshalText(data []byte) error {
	v := string(data)
	if _, cidr, err := net.ParseCIDR(v); err == nil {
		self.cidr = cidr
		return nil
	}

	if ip := net.ParseIP(v); ip != nil {
		self.ip = ip
		return nil
	}

	// minimum valid hostname is `a.b`
	// minimum valid domain name is '*.c'
	if len(v) < 3 {
		return errors.New("invalid address")
	}

	if v[0] == '*' && v[1] != '.' {
		return errors.Errorf("invalid wildcard domain '%s'", v)
	}

	self.domain = DomainName(strings.ToLower(v))
	return nil
}
