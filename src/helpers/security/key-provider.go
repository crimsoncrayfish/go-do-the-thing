package security

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/slog"
	"os"

	"github.com/google/uuid"
)

type SecretKeyProvider struct {
	keyList map[string]*rsa.PrivateKey
	kidList []string
	logger  slog.Logger
}

var source = assert.Source{"KeyProvider"}

func newKeyProvider(keysLocation string) *SecretKeyProvider {
	logger := slog.NewLogger(source.Name)

	keys := make(map[string]*rsa.PrivateKey)
	kids := make([]string, 10)

	keysList := readKeys(keysLocation, logger)

	for i, key := range keysList {
		kid := uuid.New().String()
		keys[kid] = key
		kids[i] = kid
	}

	return &SecretKeyProvider{
		keyList: keys,
		kidList: kids,
		logger:  logger,
	}
}

func (skp *SecretKeyProvider) getKey() *rsa.PrivateKey {
	// TODO: improve this so it doesnt just use the first kid
	return skp.keyList[skp.kidList[0]]
}

func readKeys(keysLocation string, logger slog.Logger) []*rsa.PrivateKey {
	keys := make([]*rsa.PrivateKey, 10)
	// TODO: add code to read from mulitple locations for multiple rotating keys
	key := readKey(keysLocation, logger)

	keys[0] = key
	return keys
}

func readKey(keyLocation string, logger slog.Logger) *rsa.PrivateKey {
	privateKeyName := "private.key"
	logger.Info("Reading file at $s$s", keyLocation, privateKeyName)

	privateKeyFile, err := os.ReadFile(keyLocation + privateKeyName)
	assert.NoError(err, source, "Could not read private key at location %s", keyLocation+privateKeyName)

	privatePem, _ := pem.Decode(privateKeyFile)
	assert.IsTrue(privatePem != nil, source,
		"Failed to decode private key file content for file at %s",
		keyLocation+privateKeyName)

	privateKeyAny, err := x509.ParsePKCS8PrivateKey(privatePem.Bytes)
	assert.NoError(err, source, "Could not parse private key file. Content potentially malformed")

	privateKey, ok := privateKeyAny.(*rsa.PrivateKey)
	assert.IsTrue(ok, source, "The private key at location '%s' is not an RSA private key", keyLocation+privateKeyName)

	publicKeyName := "public.key"
	publicKeyFile, err := os.ReadFile(keyLocation + publicKeyName)
	assert.NoError(err, source, "Could not read public key at location %s", keyLocation+privateKeyName)

	publicKeyPem, _ := pem.Decode(publicKeyFile)
	assert.IsTrue(publicKeyPem != nil, source,
		"Failed to decode public key file content for file at %s",
		keyLocation+publicKeyName)

	certificate, err := x509.ParsePKIXPublicKey(publicKeyPem.Bytes)
	assert.NoError(err, source, "Could not parse public key. Content potentially malformed")

	publicKey, ok := certificate.(*rsa.PublicKey)
	assert.IsTrue(ok, source, "Public key was not an RSA public key")
	privateKey.PublicKey = *publicKey

	logger.Info("Successfully read keys")
	return privateKey
}
