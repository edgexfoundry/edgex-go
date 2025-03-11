//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package headers

import (
	"context"
	stdErrs "errors"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"

	"github.com/golang-jwt/jwt/v5"
)

// VerifyJWT validates the JWT issued by security-proxy-auth by using the verification key provided by the security-proxy-auth service
func VerifyJWT(token string,
	issuer string,
	alg string,
	dic *di.Container,
	ctx context.Context) errors.EdgeX {
	lc := container.LoggingClientFrom(dic.Get)

	verifyKey, edgexErr := GetVerificationKey(dic, issuer, alg, ctx)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}

	err := ParseJWT(token, verifyKey, &jwt.MapClaims{}, jwt.WithExpirationRequired())
	if err != nil {
		if stdErrs.Is(err, jwt.ErrTokenExpired) {
			// Skip the JWT expired error
			lc.Debug("JWT is valid but expired")
			return nil
		} else {
			if stdErrs.Is(err, jwt.ErrTokenMalformed) ||
				stdErrs.Is(err, jwt.ErrTokenUnverifiable) ||
				stdErrs.Is(err, jwt.ErrTokenSignatureInvalid) ||
				stdErrs.Is(err, jwt.ErrTokenRequiredClaimMissing) {
				lc.Errorf("Invalid jwt : %v\n", err)
				return errors.NewCommonEdgeX(errors.KindUnauthorized, "invalid jwt", err)
			}
			lc.Errorf("Error occurred while validating JWT: %v", err)
			return errors.NewCommonEdgeX(errors.Kind(err), "failed to parse jwt", err)
		}
	}
	return nil
}

// ParseJWT parses and validates the JWT with the passed ParserOptions and returns the token which implements the Claim interface
func ParseJWT(token string, verifyKey any, claims jwt.Claims, parserOption ...jwt.ParserOption) error {
	_, err := jwt.ParseWithClaims(token, claims, func(_ *jwt.Token) (any, error) {
		return verifyKey, nil
	}, parserOption...)
	if err != nil {
		return err
	}

	issuer, err := claims.GetIssuer()
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to retrieve the issuer", err)
	}
	if len(issuer) == 0 {
		return errors.NewCommonEdgeX(errors.KindUnauthorized, "issuer is empty", err)
	}
	return nil
}
