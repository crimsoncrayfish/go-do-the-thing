package security

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/logger"
	"os"

	"github.com/google/uuid"
)

var keyProviderLogger = logger.NewLogger("KeyProvider")

type SecretKeyProvider struct {
	keyList map[string]*rsa.PrivateKey
	kidList []string
}

var source = "KeyProvider"

func newKeyProvider(env string) *SecretKeyProvider {
	keylist, err := loadKeysFromEnv()
	if err != nil {
		keyProviderLogger.Error("FATAL: Could not load security keys", "error", err)
		panic(err)
	}
	keys := make(map[string]*rsa.PrivateKey)
	kids := make([]string, len(keylist))

	for i, key := range keylist {
		kid := uuid.New().String()
		keys[kid] = key
		kids[i] = kid
	}
	keyProviderLogger.Info("KeyProvider created", "keys_loaded", len(keylist))
	return &SecretKeyProvider{
		keyList: keys,
		kidList: kids,
	}
}

func (skp *SecretKeyProvider) getKey() *rsa.PrivateKey {
	if len(skp.kidList) == 0 {
		return nil
	}
	return skp.keyList[skp.kidList[0]]
}

func parseRSAKeys(privateKeyPEM, publicKeyPEM []byte) (*rsa.PrivateKey, error) {
	privatePemBlock, _ := pem.Decode(privateKeyPEM)
	if privatePemBlock == nil {
		return nil, errors.New(errors.ErrKeysNotLoadedError, "failed to decode private key PEM block %s", privateKeyPEM)
	}
	privateKeyAny, err := x509.ParsePKCS8PrivateKey(privatePemBlock.Bytes)
	if err != nil {
		return nil, errors.New(errors.ErrKeysNotLoadedError, "could not parse private key: %w", err)
	}
	privateKey, ok := privateKeyAny.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New(errors.ErrKeysNotLoadedError, "key is not an RSA private key")
	}

	publicPemBlock, _ := pem.Decode(publicKeyPEM)
	if publicPemBlock == nil {
		return nil, errors.New(errors.ErrKeysNotLoadedError, "failed to decode public key PEM block")
	}
	publicKeyAny, err := x509.ParsePKIXPublicKey(publicPemBlock.Bytes)
	if err != nil {
		return nil, errors.New(errors.ErrKeysNotLoadedError, "could not parse public key: %w", err)
	}
	publicKey, ok := publicKeyAny.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New(errors.ErrKeysNotLoadedError, "public key is not an RSA public key")
	}

	privateKey.PublicKey = *publicKey
	return privateKey, nil
}

func loadKeysFromEnv() ([]*rsa.PrivateKey, error) {
	keyProviderLogger.Info("Reading keys from environment variables")

	privateKeyB64 := os.Getenv("JWT_PRIVATE_KEY")
	publicKeyB64 := os.Getenv("JWT_PUBLIC_KEY")

	if privateKeyB64 == "" || publicKeyB64 == "" {
		return nil, errors.New(errors.ErrKeysNotLoadedError, "JWT_PRIVATE_KEY or JWT_PUBLIC_KEY env var not set")
	}

	privateKeyStr, err := base64.StdEncoding.DecodeString(privateKeyB64)
	if err != nil {
		return nil, errors.New(errors.ErrKeysNotLoadedError, "failed to decode private key from base64: %w", err)
	}

	publicKeyStr, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		return nil, errors.New(errors.ErrKeysNotLoadedError, "failed to decode public key from base64: %w", err)
	}

	key, err := parseRSAKeys([]byte(privateKeyStr), []byte(publicKeyStr))
	if err != nil {
		return nil, err
	}

	return []*rsa.PrivateKey{key}, nil
}
