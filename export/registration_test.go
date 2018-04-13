//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package export

import "testing"

func TestRegistrationValid(t *testing.T) {
	var tests = []struct {
		name        string
		compression string
		format      string
		destination string
		encryption  string
		valid       bool
	}{
		{"empty", "", "", "", "", false},
		{"valid", CompZip, FormatJSON, DestMQTT, EncAes, true},
		{"defaultCompression", "", FormatJSON, DestMQTT, EncAes, true},
		{"defaultEncryption", CompZip, FormatJSON, DestMQTT, "", true},
		{"wrongCompresion", "INVALID", FormatJSON, DestMQTT, EncAes, false},
		{"wrongFormat", CompZip, "INVALID", DestMQTT, EncAes, false},
		{"wrongDestination", CompZip, FormatJSON, "INVALID", EncAes, false},
		{"wrongEncryption", CompZip, FormatJSON, DestMQTT, "INVALID", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Registration{}

			if tt.compression != "" {
				r.Compression = tt.compression
			}
			if tt.format != "" {
				r.Format = tt.format
			}
			if tt.destination != "" {
				r.Destination = tt.destination
			}
			if tt.encryption != "" {
				r.Encryption.Algo = tt.encryption
			}
			if valid, err := r.Validate(); valid != tt.valid {
				t.Errorf("Validate should return %v instead of %v. Reg %v, err: %v",
					tt.valid, valid, r, err)
			} else {
				if valid && err != nil {
					t.Errorf("There should not be an error if is is valid: %v", err)
				}
			}
		})
	}
}
