package parsec

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/asn1"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/identity/engines"
	"github.com/parallaxsecond/parsec-client-go/parsec"
	"github.com/parallaxsecond/parsec-client-go/parsec/algorithm"
	"io"
	"math/big"
	"net/url"
	"sync"
)

const EngineId = "parsec"

var parsecEngine = &engine{}
var log = pfxlog.ContextLogger("parsec")

func init() {
	engines.RegisterEngine(parsecEngine)
}

type engine struct {
	client *parsec.BasicClient
	initer sync.Once
}

type parsecKey struct {
	name string
	pub  crypto.PublicKey
}

func (e *engine) Id() string {
	return EngineId
}

func (e *engine) LoadKey(key *url.URL) (crypto.PrivateKey, error) {
	log.Infof("loadig key: %v", key)
	keyName := key.Opaque

	bc := e.getClient()

	pubBytes, err := bc.PsaExportPublicKey(keyName)
	if err != nil {
		return nil, err
	}
	x, y := elliptic.Unmarshal(elliptic.P256(), pubBytes)

	pub := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	return &parsecKey{
		name: keyName,
		pub:  pub,
	}, nil
}

func (e *engine) getClient() *parsec.BasicClient {
	e.initer.Do(func() {
		log.Infof("initializing client")
		config := parsec.NewClientConfig()
		config.Authenticator(parsec.NewUnixPeerAuthenticator())
		bc, err := parsec.CreateConfiguredClient(config)
		if err != nil {
			log.Fatal(err)
		}
		e.client = bc
	})

	return e.client
}

func (pk *parsecKey) Public() crypto.PublicKey {
	return pk.pub
}

func (pk *parsecKey) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) (signature []byte, err error) {
	log.Infof("key[%s] signing %d bytes", pk.name, len(digest))
	bc := parsecEngine.getClient()
	algo := algorithm.NewAsymmetricSignature().Ecdsa(algorithm.HashAlgorithmTypeSHA256).GetAsymmetricSignature()

	sigBytes, err := bc.PsaSignHash(pk.name, digest, algo)
	if err != nil {
		return nil, err
	}

	var sig struct {
		R, S *big.Int
	}

	n := len(sigBytes) / 2
	sig.R = new(big.Int)
	sig.R.SetBytes(sigBytes[:n])
	sig.S = new(big.Int)
	sig.S.SetBytes(sigBytes[n:])

	return asn1.Marshal(sig)
}
