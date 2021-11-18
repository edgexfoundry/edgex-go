package secretstore

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/helper"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"
	jwt "github.com/golang-jwt/jwt/v4"
)

// KongAdminAPI is the main struct used to generate the configuration file
type KongAdminAPI struct {
	secrets     KongAdminAPISecrets
	prefixes    KongAdminAPIPrefixes
	paths       KongAdminAPIPaths
	jwtDuration string
}

// KongAdminAPIPrefixes is used for prettier & scoped logging messages
type KongAdminAPIPrefixes struct {
	cmdText  string
	errText  string
	infoText string
}

// KongAdminAPIKey is the struct that holds all the secrets
type KongAdminAPISecrets struct {
	public  ECKey
	private ECKey
	jwt     JWT
}

// KongAdmingAPIPaths are mapped to the paths defined in configuration.toml
type KongAdminAPIPaths struct {
	template string
	config   string
	jwt      string
}

// ECKey is a simple struct to hold public/private keys in different formats
type ECKey struct {
	encoded []byte
	pem     string
}

// JWT is a simple struct to hold the necessary JWT string values
type JWT struct {
	issuer string
	signed string
}

// NewKongAdminAPI is a generic constructor to return an instance of KongAdminAPI
func NewKongAdminAPI(config config.KongAdminInfo) KongAdminAPI {
	return KongAdminAPI{
		paths: KongAdminAPIPaths{
			config:   config.ConfigFilePath,
			template: config.ConfigTemplatePath,
			jwt:      config.ConfigJWTPath,
		},
		jwtDuration: config.ConfigJWTDuration,
	}
}

// Setup creates a Kong declarative configuration file and a JWT.
//
// The configuration file is a Kong declarative configuration file in YML format
// that is written to an area where the Kong service has access and can import
// the file.
//
// The JWT is saved to a JSON based file
func (k *KongAdminAPI) Setup() error {

	// Verify prefixes and paths
	err := k.verifyPrefixes()
	if err != nil {
		return err
	}

	err = k.verifyPaths()
	if err != nil {
		return err
	}

	// Generate the keys for the Kong Admin API
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("%s Failed to generate private/public key pair: %w", k.prefixes.errText, err)
	}

	// Get an encoded version of the keys
	k.secrets.private.encoded, err = x509.MarshalECPrivateKey(key)
	if err != nil {
		return fmt.Errorf("%s Failed to marshal the private key: %w", k.prefixes.errText, err)
	}

	k.secrets.public.encoded, err = x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return fmt.Errorf("%s Failed to marshal the public key: %w", k.prefixes.errText, err)
	}

	// PEM encode the keys
	k.secrets.private.pem = string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: k.secrets.private.encoded}))
	k.secrets.public.pem = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: k.secrets.public.encoded}))

	// When complete, clear the secrets portion of the KongAdminAPI struct
	defer k.clearSecrets()

	// Read in the configuration template
	configTemplateBytes, err := os.ReadFile(k.paths.template)
	if err != nil {
		return fmt.Errorf("%s Failed to read config template from file %s: %w", k.prefixes.errText, k.paths.template, err)
	}

	// Set random string for JWT issuer payload value - the default issuer value assigned
	// by Kong is 32 bytes in length - simply mirroring that for consistency
	k.secrets.jwt.issuer = helper.GeneratePseudoRandomString(32)

	// Insert public key
	configTemplateText := strings.Replace(string(configTemplateBytes),
		"<<INSERT-ADMIN-PUBLIC-KEY>>", strings.ReplaceAll(strings.TrimSpace(k.secrets.public.pem), "\n", "\n    "), -1)

	// Insert issuer
	configTemplateText = strings.Replace(configTemplateText,
		"<<INSERT-ADMIN-JWT-ISSUER-KEY>>", k.secrets.jwt.issuer, -1)

	// Write the config file to the configured save path
	// note: config file contains no confidential data -- 0644 since it is owned by root
	err = os.WriteFile(k.paths.config, []byte(configTemplateText), 0644) // nolint:gosec
	if err != nil {
		return fmt.Errorf("%s Failed to write config template to file %s: %w", k.prefixes.errText, k.paths.config, err)
	}

	// Create the signed JWT
	k.secrets.jwt.signed, err = k.createJWT(k.secrets.private.pem, k.secrets.jwt.issuer)
	if err != nil {
		return fmt.Errorf("%s Failed to create signed JSON Web Token: %w", k.prefixes.errText, err)
	}

	// Write JWT to secret file (used solely by "admin" group in Kong)
	err = os.WriteFile(k.paths.jwt, []byte(k.secrets.jwt.signed), 0600)
	if err != nil {
		return fmt.Errorf("%s Failed to write JWT to file %s: %w", k.prefixes.errText, k.paths.jwt, err)
	}

	return nil
}

// createJWT creates a JWT from a given private key and issuer payload value
func (k *KongAdminAPI) createJWT(privateKey string, issuer string) (signedToken string, err error) {

	// Setup JWT generation variables
	now := time.Now()
	duration, err := time.ParseDuration(k.jwtDuration)
	if err != nil {
		return "", fmt.Errorf("%s Could not parse JWT duration: %w", k.prefixes.errText, err)
	}

	// Sanity check - parse & check EC key
	eckey, err := jwt.ParseECPrivateKeyFromPEM([]byte(privateKey))
	if err != nil {
		return "", fmt.Errorf("%s Could not parse private key: %w", k.prefixes.errText, err)
	}
	if eckey.Params().BitSize != 256 {
		return "", fmt.Errorf("%s EC key bit size is incorrect (%d instead of 256)", k.prefixes.errText, eckey.Params().BitSize)
	}

	// Create JWT
	token := jwt.NewWithClaims(jwt.SigningMethodES256, &jwt.RegisteredClaims{
		Issuer:    issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
	})

	// Save JWT to string return
	signedToken, err = token.SignedString(eckey)
	if err != nil {
		return "", fmt.Errorf("%s Could not sign JWT: %w", k.prefixes.errText, err)
	}

	return
}

// clearSecrets clears the key structs by setting them to empty struct values
func (k *KongAdminAPI) clearSecrets() {
	k.secrets = KongAdminAPISecrets{}
}

// verifyPrefixes confirms that there are values set for the prefixes
func (k *KongAdminAPI) verifyPrefixes() error {

	base := "[build-kong-admin]"

	if k.prefixes.cmdText == "" {
		k.prefixes.cmdText = base
	}

	if k.prefixes.errText == "" {
		k.prefixes.errText = base + "[error]"
	}

	if k.prefixes.infoText == "" {
		k.prefixes.infoText = base + "[info]"
	}

	return nil
}

// verifyPaths confirms values are set for the paths, and that those paths exist
func (k *KongAdminAPI) verifyPaths() error {

	// Is the template path set?
	if k.paths.template == "" {
		return fmt.Errorf("%s Template file path (<KongAdminAPI>.paths.template) is not set - check configuration.toml", k.prefixes.errText)
	}

	// Does the template path/file exist? It should be set.
	if !helper.CheckIfFileExists(k.paths.template) {
		return fmt.Errorf("%s Template file path (<KongAdminAPI>.paths.template) does not exist - check configuration.toml", k.prefixes.errText)
	}

	// Is the config file path set?
	if k.paths.config == "" {
		return fmt.Errorf("%s Kong configuration file path (<KongAdminAPI>.paths.config) is not set - check configuration.toml", k.prefixes.errText)
	}

	// Does the config file path directory exist? If not, create it.
	if !helper.CheckIfFileExists(k.paths.config) {
		dirErr := os.MkdirAll(path.Dir(k.paths.config), 0755)
		if dirErr != nil {
			return fmt.Errorf("%s Kong configuration file path (<KongAdminAPI>.paths.config) could not be created", k.prefixes.errText)
		}
	}

	// Is the JWT path set?
	if k.paths.jwt == "" {
		return fmt.Errorf("%s JWT file path (<KongAdminAPI>.paths.jwt) is not set - check configuration.toml", k.prefixes.errText)
	}

	// Does the JWT file path directory exist? If not, create it.
	if _, err := os.Stat(k.paths.jwt); os.IsNotExist(err) {
		dirErr := os.MkdirAll(path.Dir(k.paths.jwt), 0755)
		if dirErr != nil {
			return fmt.Errorf("%s JWT file path (<KongAdminAPI>.paths.jwt) could not be created", k.prefixes.errText)
		}
	}

	return nil
}
