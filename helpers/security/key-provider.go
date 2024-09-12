package security

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"go-do-the-thing/helpers/slog"
	"os"

	"github.com/google/uuid"
)

type SecretKeyProvider struct {
	keyList map[string]*rsa.PrivateKey
	kidList []string
	logger  *slog.Logger
}

func newKeyProvider(keysLocation string) (*SecretKeyProvider, error) {
	logger := slog.NewLogger("KeyProvider")

	keys := make(map[string]*rsa.PrivateKey)
	kids := make([]string, 10)

	keysList, err := readKeys(keysLocation, logger)
	if err != nil {
		return nil, err
	}
	for i, key := range keysList {
		kid := uuid.New().String()
		keys[kid] = key
		kids[i] = kid
	}

	return &SecretKeyProvider{
		keyList: keys,
		kidList: kids,
		logger:  logger,
	}, nil
}

func (skp *SecretKeyProvider) getKey() *rsa.PrivateKey {
	// todo improve this so it doesnt just use the first kid
	return skp.keyList[skp.kidList[0]]
}

func readKeys(keysLocation string, logger *slog.Logger) ([]*rsa.PrivateKey, error) {
	keys := make([]*rsa.PrivateKey, 10)
	// TODO add code to read from mulitple locations for multiple rotating keys
	key, err := readKey(keysLocation, logger)
	if err != nil {
		return nil, err
	}

	keys[0] = key
	return keys, nil
}

func readKey(keyLocation string, logger *slog.Logger) (*rsa.PrivateKey, error) {
	privateKeyName := "private.key"
	privateKeyFile, err := os.ReadFile(keyLocation + privateKeyName)
	logger.Info("Reading file at " + keyLocation + privateKeyName)
	if err != nil {
		logger.Error(err, "Could not read private key at location %s", keyLocation+privateKeyName)
		return nil, err
	}
	privatePem, _ := pem.Decode(privateKeyFile)
	if privatePem == nil {
		err := errors.New("failed to decode private key file")

		logger.Error(err,
			"Failed to decode private key file content for file at %s",
			keyLocation+privateKeyName,
		)
		return nil, err
	}
	privateKeyAny, err := x509.ParsePKCS8PrivateKey(privatePem.Bytes)
	if err != nil {
		logger.Error(err, "Could not parse private key file. Content potentially malformed")
		return nil, err
	}
	privateKey, ok := privateKeyAny.(*rsa.PrivateKey)
	if !ok {
		err := errors.New("Incorrect key type")

		logger.Error(err, "The private key at location '%s' is not an RSA private key", keyLocation+privateKeyName)
		return nil, err
	}

	publicKeyName := "public.key"
	publicKeyFile, err := os.ReadFile(keyLocation + publicKeyName)
	if err != nil {
		logger.Error(err, "Could not read public key at location %s", keyLocation+privateKeyName)
		return nil, err
	}
	publicKeyPem, _ := pem.Decode(publicKeyFile)
	if publicKeyPem == nil {
		err := errors.New("failed to decode public key file")

		logger.Error(err,
			"Failed to decode public key file content for file at %s",
			keyLocation+publicKeyName,
		)
		return nil, err
	}
	certificate, err := x509.ParsePKIXPublicKey(publicKeyPem.Bytes)
	if err != nil {
		logger.Error(err, "Could not parse public key. Content potentially malformed")
		return nil, err
	}
	publicKey, ok := certificate.(*rsa.PublicKey)
	if !ok {
		err := errors.New("public key was not rsa public key")

		logger.Error(err, "Public key was not an RSA public key")
		return nil, err
	}

	privateKey.PublicKey = *publicKey

	return privateKey, nil
}
