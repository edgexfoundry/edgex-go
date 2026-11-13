//
// Copyright (C) 2026 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0
//

package application

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

// reservedNameRunes lists the characters that change the semantics of the
// message-bus topics and subjects built from metadata entity names
// (https://github.com/edgexfoundry/edgex-go/issues/5313):
//
//   - '+' and '#' are the MQTT wildcards, so a device (or profile, service,
//     provision watcher) named with one of them makes every subscription
//     filter that embeds the name match other devices' topics as well;
//   - '/' separates MQTT topic levels, so it breaks the level structure the
//     clients and subscriptions rely on;
//   - '*' and '>' are the NATS wildcards and '.' the NATS token separator;
//   - '.' is also reserved by the Redis topics / RegexCmd matching;
//   - whitespace is not allowed in NATS subjects at all.
//
// All of these are percent-encoded on the topics when the
// EnableNameFieldEscape configuration is enabled (see common.URLEncode), so
// the write-time validation below only rejects them while that safety net is
// disabled - which is the default.
const reservedNameRunes = "+#/*.>"

// validateEntityName rejects metadata entity names whose characters cannot be
// carried by the message-bus topics built from them. When name field escape is
// enabled the characters are percent-encoded instead and every name is
// accepted.
func validateEntityName(name string, escapeEnabled bool) errors.EdgeX {
	// The empty name check is independent of the escaping safety net: an
	// escaped name still has to be a name.
	if name == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "name is required", nil)
	}

	if escapeEnabled {
		return nil
	}

	// IndexFunc reports a byte offset, while the offending character has to be
	// reported as a rune; capture it in the predicate instead of indexing the
	// rune slice with a byte offset.
	var reserved rune
	idx := strings.IndexFunc(name, func(r rune) bool {
		if unicode.IsSpace(r) || strings.ContainsRune(reservedNameRunes, r) {
			reserved = r
			return true
		}
		return false
	})
	if idx >= 0 {
		return errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("name '%s' contains the reserved character %q; metadata names must not contain whitespace or any of %+q while EnableNameFieldEscape is disabled, because the message-bus topics are built from them", name, reserved, reservedNameRunes), nil)
	}
	return nil
}
