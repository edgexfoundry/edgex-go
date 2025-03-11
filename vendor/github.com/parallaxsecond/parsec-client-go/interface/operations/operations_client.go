// Copyright 2021 Contributors to the Parsec project.
// SPDX-License-Identifier: Apache-2.0

package operations

import (
	"bytes"

	"github.com/parallaxsecond/parsec-client-go/interface/auth"
	connection "github.com/parallaxsecond/parsec-client-go/interface/connection"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/deleteclient"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/listauthenticators"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/listclients"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/listkeys"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/listopcodes"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/listproviders"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/ping"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/psaaeaddecrypt"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/psaaeadencrypt"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/psaalgorithm"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/psaasymmetricdecrypt"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/psaasymmetricencrypt"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/psacipherdecrypt"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/psacipherencrypt"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/psadestroykey"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/psaexportkey"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/psaexportpublickey"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/psageneratekey"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/psageneraterandom"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/psahashcompute"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/psaimportkey"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/psakeyattributes"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/psamaccompute"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/psamacverify"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/psarawkeyagreement"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/psasignhash"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/psasignmessage"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/psaverifyhash"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/psaverifymessage"
	"github.com/parallaxsecond/parsec-client-go/interface/requests"
	"google.golang.org/protobuf/proto"
)

// Client is a Parsec client representing a connection and set of API implementations
type Client struct {
	conn connection.Connection
}

// InitClient initializes a Parsec client
func InitClient() (*Client, error) {
	conn, err := connection.NewDefaultConnection()
	if err != nil {
		return nil, err
	}
	client := &Client{
		conn: conn,
	}

	return client, nil
}

// InitClient initializes a Parsec client
func InitClientFromConnection(conn connection.Connection) (*Client, error) {
	client := &Client{
		conn: conn,
	}

	return client, nil
}

func (c *Client) Close() error {
	// Just in case
	return c.conn.Close()
}

// Ping server and return wire protocol major and minor version number
func (c Client) Ping(provider requests.ProviderID, authenticator auth.Authenticator) (uint8, uint8, error) { //nolint:gocritic
	req := &ping.Operation{}
	resp := &ping.Result{}
	err := c.operation(provider, authenticator, requests.OpPing, req, resp)
	if err != nil {
		return 0, 0, err
	}

	return uint8(resp.WireProtocolVersionMaj), uint8(resp.WireProtocolVersionMin), nil
}

// ListProviders returns a list of the providers supported by the server.
func (c Client) ListProviders(provider requests.ProviderID, authenticator auth.Authenticator) ([]*listproviders.ProviderInfo, error) {
	req := &listproviders.Operation{}
	resp := &listproviders.Result{}
	err := c.operation(provider, authenticator, requests.OpListProviders, req, resp)
	if err != nil {
		return nil, err
	}

	return resp.GetProviders(), nil
}

// ListOpcodes list the opcodes for a provider
func (c Client) ListOpcodes(provider requests.ProviderID, authenticator auth.Authenticator, providerID uint32) ([]uint32, error) {
	req := &listopcodes.Operation{ProviderId: providerID}
	resp := &listopcodes.Result{}
	err := c.operation(provider, authenticator, requests.OpListOpcodes, req, resp)
	if err != nil {
		return nil, err
	}

	return resp.GetOpcodes(), nil
}

// ListClients lists the clients
func (c Client) ListClients(provider requests.ProviderID, authenticator auth.Authenticator) ([]string, error) {
	req := &listclients.Operation{}
	resp := &listclients.Result{}
	err := c.operation(provider, authenticator, requests.OpListClients, req, resp)
	if err != nil {
		return nil, err
	}
	return resp.GetClients(), nil
}

func (c Client) DeleteClient(provider requests.ProviderID, authenticator auth.Authenticator, client string) error {
	req := &deleteclient.Operation{Client: client}
	resp := &deleteclient.Result{}

	return c.operation(provider, authenticator, requests.OpDeleteClient, req, resp)
}

// ListKeys obtain keys stored for current application
func (c Client) ListKeys(provider requests.ProviderID, authenticator auth.Authenticator) ([]*listkeys.KeyInfo, error) {
	req := &listkeys.Operation{}
	resp := &listkeys.Result{}
	err := c.operation(provider, authenticator, requests.OpListKeys, req, resp)
	if err != nil {
		return nil, err
	}
	return resp.GetKeys(), nil
}

// ListAuthenticators obtain authenticators supported by server
func (c Client) ListAuthenticators(provider requests.ProviderID, authenticator auth.Authenticator) ([]*listauthenticators.AuthenticatorInfo, error) {
	req := &listauthenticators.Operation{}
	resp := &listauthenticators.Result{}
	err := c.operation(provider, authenticator, requests.OpListAuthenticators, req, resp)
	if err != nil {
		return nil, err
	}
	return resp.GetAuthenticators(), nil
}

// PsaGenerateKey create key named name with attributes
func (c Client) PsaGenerateKey(provider requests.ProviderID, authenticator auth.Authenticator, name string, attributes *psakeyattributes.KeyAttributes) error {
	req := &psageneratekey.Operation{
		KeyName:    name,
		Attributes: attributes,
	}
	resp := &psageneratekey.Result{}

	return c.operation(provider, authenticator, requests.OpPsaGenerateKey, req, resp)
}

// PsaDestroyKey destroys a key with given name
func (c Client) PsaDestroyKey(provider requests.ProviderID, authenticator auth.Authenticator, name string) error {
	req := &psadestroykey.Operation{
		KeyName: name,
	}
	resp := &psadestroykey.Result{}

	return c.operation(provider, authenticator, requests.OpPsaDestroyKey, req, resp)
}

// PsaHashCompute calculates a hash of a message using specified algorithm
func (c Client) PsaHashCompute(provider requests.ProviderID, authenticator auth.Authenticator, message []byte, alg psaalgorithm.Algorithm_Hash) ([]byte, error) {
	req := &psahashcompute.Operation{
		Input: message,
		Alg:   alg,
	}
	resp := &psahashcompute.Result{}

	err := c.operation(provider, authenticator, requests.OpPsaHashCompute, req, resp)
	if err != nil {
		return nil, err
	}
	return resp.Hash, nil
}

// PsaSignMessage signs message using signingKey and algorithm, returning the signature.
func (c Client) PsaSignMessage(provider requests.ProviderID, authenticator auth.Authenticator, signingKey string, message []byte, alg *psaalgorithm.Algorithm_AsymmetricSignature) ([]byte, error) {
	req := &psasignmessage.Operation{
		KeyName: signingKey,
		Alg:     alg,
		Message: message,
	}
	resp := &psasignmessage.Result{}

	err := c.operation(provider, authenticator, requests.OpPsaSignMessage, req, resp)

	if err != nil {
		return nil, err
	}
	return resp.Signature, nil
}

// PsaSignHash signs hash using signingKey and algorithm, returning the signature.
func (c Client) PsaSignHash(provider requests.ProviderID, authenticator auth.Authenticator, signingKey string, hash []byte, alg *psaalgorithm.Algorithm_AsymmetricSignature) ([]byte, error) {
	req := &psasignhash.Operation{
		KeyName: signingKey,
		Alg:     alg,
		Hash:    hash,
	}
	resp := &psasignhash.Result{}

	err := c.operation(provider, authenticator, requests.OpPsaSignHash, req, resp)

	if err != nil {
		return nil, err
	}
	return resp.Signature, nil
}

// PsaVerifyMessage verify a signature  of message with verifyingKey using signature algorithm alg.
func (c Client) PsaVerifyMessage(provider requests.ProviderID, authenticator auth.Authenticator, verifyingKey string, message, signature []byte, alg *psaalgorithm.Algorithm_AsymmetricSignature) error {
	req := &psaverifymessage.Operation{
		KeyName:   verifyingKey,
		Message:   message,
		Signature: signature,
		Alg:       alg,
	}
	resp := &psaverifymessage.Result{}

	return c.operation(provider, authenticator, requests.OpPsaVerifyMessage, req, resp)
}

// PsaVerifyHash verify a signature  of hash with verifyingKey using signature algorithm alg.
func (c Client) PsaVerifyHash(provider requests.ProviderID, authenticator auth.Authenticator, verifyingKey string, hash, signature []byte, alg *psaalgorithm.Algorithm_AsymmetricSignature) error {
	req := &psaverifyhash.Operation{
		KeyName:   verifyingKey,
		Hash:      hash,
		Signature: signature,
		Alg:       alg,
	}
	resp := &psaverifymessage.Result{}

	return c.operation(provider, authenticator, requests.OpPsaVerifyHash, req, resp)
}

// PsaCipherEncrypt carries out symmetric encryption on plaintext using defined key/algorithm, returning ciphertext
func (c Client) PsaCipherEncrypt(provider requests.ProviderID, authenticator auth.Authenticator, keyName string, alg psaalgorithm.Algorithm_Cipher, plaintext []byte) ([]byte, error) {
	req := &psacipherencrypt.Operation{
		KeyName:   keyName,
		Alg:       alg,
		Plaintext: plaintext,
	}
	resp := &psacipherencrypt.Result{}

	err := c.operation(provider, authenticator, requests.OpPsaCipherEncrypt, req, resp)
	if err != nil {
		return nil, err
	}
	return resp.Ciphertext, nil
}

// PsaCipherDecrypt decrypts symmetrically encrypted ciphertext using defined key/algorithm, returning plaintext
func (c Client) PsaCipherDecrypt(provider requests.ProviderID, authenticator auth.Authenticator, keyName string, alg psaalgorithm.Algorithm_Cipher, ciphertext []byte) ([]byte, error) {
	req := &psacipherdecrypt.Operation{
		KeyName:    keyName,
		Alg:        alg,
		Ciphertext: ciphertext,
	}
	resp := &psacipherdecrypt.Result{}

	err := c.operation(provider, authenticator, requests.OpPsaCipherDecrypt, req, resp)
	if err != nil {
		return nil, err
	}
	return resp.Plaintext, nil
}

func (c Client) PsaAeadDecrypt(provider requests.ProviderID, authenticator auth.Authenticator, keyName string, alg *psaalgorithm.Algorithm_Aead, nonce, additionalData, ciphertext []byte) ([]byte, error) {
	req := &psaaeaddecrypt.Operation{
		KeyName:        keyName,
		Alg:            alg,
		Nonce:          nonce,
		AdditionalData: additionalData,
		Ciphertext:     ciphertext,
	}
	resp := &psaaeaddecrypt.Result{}

	err := c.operation(provider, authenticator, requests.OpPsaAeadDecrypt, req, resp)
	if err != nil {
		return nil, err
	}
	return resp.GetPlaintext(), nil
}

func (c Client) PsaAeadEncrypt(provider requests.ProviderID, authenticator auth.Authenticator, keyName string, alg *psaalgorithm.Algorithm_Aead, nonce, additionalData, plaintext []byte) ([]byte, error) {
	req := &psaaeadencrypt.Operation{
		KeyName:        keyName,
		Alg:            alg,
		Nonce:          nonce,
		AdditionalData: additionalData,
		Plaintext:      plaintext,
	}
	resp := &psaaeadencrypt.Result{}

	err := c.operation(provider, authenticator, requests.OpPsaAeadEncrypt, req, resp)
	if err != nil {
		return nil, err
	}
	return resp.GetCiphertext(), nil
}

func (c Client) PsaExportKey(provider requests.ProviderID, authenticator auth.Authenticator, keyName string) ([]byte, error) {
	req := &psaexportkey.Operation{
		KeyName: keyName,
	}
	resp := &psaexportkey.Result{}

	err := c.operation(provider, authenticator, requests.OpPsaExportKey, req, resp)
	if err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (c Client) PsaImportKey(provider requests.ProviderID, authenticator auth.Authenticator, keyName string, attributes *psakeyattributes.KeyAttributes, data []byte) error {
	req := &psaimportkey.Operation{
		KeyName:    keyName,
		Attributes: attributes,
		Data:       data,
	}
	resp := &psaimportkey.Result{}

	err := c.operation(provider, authenticator, requests.OpPsaImportKey, req, resp)
	if err != nil {
		return err
	}
	return nil
}

func (c Client) PsaExportPublicKey(provider requests.ProviderID, authenticator auth.Authenticator, keyName string) ([]byte, error) {
	req := &psaexportpublickey.Operation{
		KeyName: keyName,
	}
	resp := &psaexportpublickey.Result{}

	err := c.operation(provider, authenticator, requests.OpPsaExportPublicKey, req, resp)
	if err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (c Client) PsaGenerateRandom(provider requests.ProviderID, authenticator auth.Authenticator, size uint64) ([]byte, error) {
	req := &psageneraterandom.Operation{
		Size: size,
	}
	resp := &psageneraterandom.Result{}

	err := c.operation(provider, authenticator, requests.OpPsaGenerateRandom, req, resp)
	if err != nil {
		return nil, err
	}
	return resp.GetRandomBytes(), nil
}

func (c Client) PsaMACCompute(provider requests.ProviderID, authenticator auth.Authenticator, keyName string, alg *psaalgorithm.Algorithm_Mac, input []byte) ([]byte, error) {
	req := &psamaccompute.Operation{
		KeyName: keyName,
		Alg:     alg,
		Input:   input,
	}
	resp := &psamaccompute.Result{}

	err := c.operation(provider, authenticator, requests.OpPsaMacCompute, req, resp)
	if err != nil {
		return nil, err
	}
	return resp.GetMac(), nil
}

func (c Client) PsaMACVerify(provider requests.ProviderID, authenticator auth.Authenticator, keyName string, alg *psaalgorithm.Algorithm_Mac, input, mac []byte) error {
	req := &psamacverify.Operation{
		KeyName: keyName,
		Alg:     alg,
		Mac:     mac,
		Input:   input,
	}
	resp := &psamacverify.Result{}

	return c.operation(provider, authenticator, requests.OpPsaMacCompute, req, resp)
}

func (c Client) PsaRawKeyAgreement(provider requests.ProviderID, authenticator auth.Authenticator, alg *psaalgorithm.Algorithm_KeyAgreement_Raw, privateKey string, peerKey []byte) ([]byte, error) {
	req := &psarawkeyagreement.Operation{
		Alg:            *alg,
		PrivateKeyName: privateKey,
		PeerKey:        peerKey,
	}
	resp := &psarawkeyagreement.Result{}

	err := c.operation(provider, authenticator, requests.OpPsaRawKeyAgreement, req, resp)
	if err != nil {
		return nil, err
	}
	return resp.GetSharedSecret(), nil
}

func (c Client) PsaAsymmetricDecrypt(provider requests.ProviderID, authenticator auth.Authenticator, keyName string, alg *psaalgorithm.Algorithm_AsymmetricEncryption, salt, ciphertext []byte) ([]byte, error) {
	req := &psaasymmetricdecrypt.Operation{
		KeyName:    keyName,
		Alg:        alg,
		Salt:       salt,
		Ciphertext: ciphertext,
	}
	resp := &psaasymmetricdecrypt.Result{}

	err := c.operation(provider, authenticator, requests.OpPsaAsymmetricDecrypt, req, resp)
	if err != nil {
		return nil, err
	}
	return resp.GetPlaintext(), nil
}

func (c Client) PsaAsymmetricEncrypt(provider requests.ProviderID, authenticator auth.Authenticator, keyName string, alg *psaalgorithm.Algorithm_AsymmetricEncryption, salt, plaintext []byte) ([]byte, error) {
	req := &psaasymmetricencrypt.Operation{
		KeyName:   keyName,
		Alg:       alg,
		Salt:      salt,
		Plaintext: plaintext,
	}
	resp := &psaasymmetricencrypt.Result{}

	err := c.operation(provider, authenticator, requests.OpPsaAsymmetricEncrypt, req, resp)
	if err != nil {
		return nil, err
	}
	return resp.GetCiphertext(), nil
}

func (c Client) operation(provider requests.ProviderID, authenticator auth.Authenticator, op requests.OpCode, request, response proto.Message) error {
	err := c.conn.Open()
	if err != nil {
		return err
	}
	defer c.conn.Close()

	r, err := requests.NewRequest(op, request, authenticator, provider)
	if err != nil {
		return err
	}
	b, err := r.Pack()
	if err != nil {
		return err
	}
	// TODO ensure that we continue writing whole buffer afer a short write
	// https://github.com/parallaxsecond/parsec-client-go/issues/23
	_, err = c.conn.Write(b.Bytes())
	if err != nil {
		return err
	}

	rcvBuf := new(bytes.Buffer)
	_, err = rcvBuf.ReadFrom(c.conn)
	if err != nil {
		return err
	}

	return requests.ParseResponse(op, rcvBuf, response)
}
