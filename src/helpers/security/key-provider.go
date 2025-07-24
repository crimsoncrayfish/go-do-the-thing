package security

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/slog"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type SecretKeyProvider struct {
	keyList map[string]*rsa.PrivateKey
	kidList []string
	logger  slog.Logger
}

var source = "KeyProvider"

func newKeyProvider(env, working_dir string) *SecretKeyProvider {
	logger := slog.NewLogger(source)
	var keylist []*rsa.PrivateKey
	var err error
	if env == "development" {
		keylist, err = loadKeysFromFile(filepath.Join(working_dir, "keys"), logger)
	} else {
		keylist, err = loadKeysFromEnv(logger)
	}

	if err != nil {
		logger.Fatal("FATAL: Could not load security keys: %v", err)
	}
	keys := make(map[string]*rsa.PrivateKey)
	kids := make([]string, len(keylist))

	for i, key := range keylist {
		kid := uuid.New().String()
		keys[kid] = key
		kids[i] = kid
	}
	logger.Info("KeyProvider created", "keys_loaded", len(keylist))
	return &SecretKeyProvider{
		keyList: keys,
		kidList: kids,
		logger:  logger,
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
		return nil, errors.New(errors.ErrKeysNotLoadedError, "failed to decode private key PEM block")
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

func loadKeysFromFile(keysLocation string, logger slog.Logger) ([]*rsa.PrivateKey, error) {
	logger.Info("Reading keys from filesystem: %s", keysLocation)

	privateKeyFile, err := os.ReadFile(filepath.Join(keysLocation, "private.key"))
	if err != nil {
		return nil, errors.New(errors.ErrKeysNotLoadedError, "could not read private.key: %w", err)
	}

	publicKeyFile, err := os.ReadFile(filepath.Join(keysLocation, "public.key"))
	if err != nil {
		return nil, errors.New(errors.ErrKeysNotLoadedError, "could not read public.key: %w", err)
	}

	key, err := parseRSAKeys(privateKeyFile, publicKeyFile)
	if err != nil {
		return nil, err
	}

	return []*rsa.PrivateKey{key}, nil
}

func loadKeysFromEnv(logger slog.Logger) ([]*rsa.PrivateKey, error) {
	logger.Info("Reading keys from environment variables")

	privateKeyStr := os.Getenv("JWT_PRIVATE_KEY")
	publicKeyStr := os.Getenv("JWT_PUBLIC_KEY")

	if privateKeyStr == "" || publicKeyStr == "" {
		return nil, errors.New(errors.ErrKeysNotLoadedError, "JWT_PRIVATE_KEY or JWT_PUBLIC_KEY env var not set")
	}

	key, err := parseRSAKeys([]byte(privateKeyStr), []byte(publicKeyStr))
	if err != nil {
		return nil, err
	}

	return []*rsa.PrivateKey{key}, nil
}
