//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package export

import (
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"testing"
)

func TestRegistrationValid(t *testing.T) {
	var tests = []struct {
		name        string
		regName     string
		compression string
		format      string
		destination string
		encryption  string
		valid       bool
	}{
		{"empty", "", "", "", "", "", false},
		{"valid", "reg", models.CompZip, models.FormatJSON, models.DestMQTT, models.EncAes, true},
		{"defaultCompression", "reg", "", models.FormatJSON, models.DestMQTT, models.EncAes, true},
		{"defaultEncryption", "reg", models.CompZip, models.FormatJSON, models.DestMQTT, "", true},
		{"withoutName", "", models.CompZip, models.FormatJSON, models.DestMQTT, models.EncAes, false},
		{"wrongCompresion", "reg", "INVALID", models.FormatJSON, models.DestMQTT, models.EncAes, false},
		{"wrongFormat", "reg", models.CompZip, "INVALID", models.DestMQTT, models.EncAes, false},
		{"wrongDestination", "reg", models.CompZip, models.FormatJSON, "INVALID", models.EncAes, false},
		{"wrongEncryption", "reg", models.CompZip, models.FormatJSON, models.DestMQTT, "INVALID", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := models.Registration{}

			r.Name = tt.regName

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
