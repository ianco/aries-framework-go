/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package verifiable

import (
	"fmt"

	"github.com/square/go-jose/v3"
	"github.com/square/go-jose/v3/jwt"
)

// MarshalJWS serializes JWT presentation claims into signed form (JWS)
// todo refactor, do not pass privateKey (https://github.com/hyperledger/aries-framework-go/issues/339)
func marshalJWS(jwtClaims interface{}, signatureAlg JWSAlgorithm, privateKey interface{}, keyID string) (string, error) { //nolint:lll
	joseAlg, err := signatureAlg.jose()
	if err != nil {
		return "", err
	}
	key := jose.SigningKey{Algorithm: joseAlg, Key: privateKey}

	var signerOpts = &jose.SignerOptions{}
	signerOpts.WithType("JWT")
	signerOpts.WithHeader("kid", keyID)

	signer, err := jose.NewSigner(key, signerOpts)
	if err != nil {
		return "", fmt.Errorf("failed to create signer: %w", err)
	}

	// create an instance of Builder that uses the signer
	builder := jwt.Signed(signer).Claims(jwtClaims)

	// validate all ok, sign with the key, and return a compact JWT
	jws, err := builder.CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return jws, nil
}

func verifyJWTSignature(token *jwt.JSONWebToken, fetcher PublicKeyFetcher, issuer string, jwtClaims interface{}) error {
	var keyID string
	for _, h := range token.Headers {
		if h.KeyID != "" {
			keyID = h.KeyID
			break
		}
	}
	publicKey, err := fetcher(issuer, keyID)
	if err != nil {
		return fmt.Errorf("failed to get public key for JWT signature verification: %w", err)
	}
	if err = token.Claims(publicKey, jwtClaims); err != nil {
		return fmt.Errorf("JWT signature verification failed: %w", err)
	}
	return nil
}