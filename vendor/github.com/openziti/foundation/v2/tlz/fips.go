//go:build goexperiment.opensslcrypto && requirefips

package tlz

import (
    "crypto/boring"
    _ "crypto/tls/fipsonly"
)

// returns true if the OpenSSL FIPS provider is active at runtime
func FipsEnabled() bool {
    return boring.Enabled()
}
