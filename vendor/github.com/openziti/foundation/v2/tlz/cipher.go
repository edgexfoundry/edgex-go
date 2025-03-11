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

package tlz

import (
	"crypto/tls"
	"golang.org/x/sys/cpu"
	"sync"
)

var once = sync.Once{}
var defaultCipherSuites []uint16

var additionalCipherSuites = []uint16{
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,

	//Below here are cipher suites required for TLS 1.1 and TLS 1.0, they are disabled as the minimum
	//TLS version is 1.2 and they open TLS servers up to BEAST/LUCKY13 vulnerabilities.

	//tls.TLS_RSA_WITH_AES_128_GCM_SHA256, //no PFS
	//tls.TLS_RSA_WITH_AES_256_GCM_SHA384, //no PFS

	//cipher block chaining (CBC)
	//tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	//tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
	//tls.TLS_RSA_WITH_AES_256_CBC_SHA, //no PFS

	//low bit length CBCs for Java 7
	//tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
	//tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	//tls.TLS_RSA_WITH_AES_128_CBC_SHA, //no PFS
	//tls.TLS_RSA_WITH_AES_128_CBC_SHA256, //no PFS
	//tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
	//tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
}

func GetMinTlsVersion() uint16 {
	return tls.VersionTLS12
}

// Note: This will only affect TLS1.2 and lower, TLS1.3 has a separate smaller cipher set managed by Go.
func GetCipherSuites() []uint16 {
	once.Do(setDefaultCipherSuites)
	return defaultCipherSuites
}

func setDefaultCipherSuites() {
	var acceleratedSuites []uint16

	var (
		hasGCMAsmAMD64 = cpu.X86.HasAES && cpu.X86.HasPCLMULQDQ
		hasGCMAsmARM64 = cpu.ARM64.HasAES && cpu.ARM64.HasPMULL
		hasGCMAsmS390X = cpu.S390X.HasAES && cpu.S390X.HasAESCBC && cpu.S390X.HasAESCTR && (cpu.S390X.HasGHASH || cpu.S390X.HasAESGCM)

		hasGCMAsm = hasGCMAsmAMD64 || hasGCMAsmARM64 || hasGCMAsmS390X
	)

	if hasGCMAsm {
		acceleratedSuites = []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		}
	} else {
		acceleratedSuites = []uint16{
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		}
	}

	var suites []uint16
	suites = append(suites, acceleratedSuites...)
	suites = append(suites, additionalCipherSuites...)

	defaultCipherSuites = suites
}
