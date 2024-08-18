package security

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
)

type SecretKeyProvider struct {
	keyList map[string]*rsa.PrivateKey
	kidList []string
}

func newKeyProvider(keysLocation string) (*SecretKeyProvider, error) {
	keys := make(map[string]*rsa.PrivateKey)
	kids := make([]string, 10)

	keysList, err := readKeys(keysLocation)
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
	}, nil
}

func (skp *SecretKeyProvider) getKey() *rsa.PrivateKey {
	// todo improve this so it doesnt just use the first kid
	return skp.keyList[skp.kidList[0]]
}

func readKeys(keysLocation string) ([]*rsa.PrivateKey, error) {
	keys := make([]*rsa.PrivateKey, 10)
	// TODO add code to read from mulitple locations for multiple rotating keys
	key, err := readKey(keysLocation)
	if err != nil {
		return nil, err
	}

	keys[0] = key
	return keys, nil
}

func readKey(keyLocation string) (*rsa.PrivateKey, error) {
	privateKeyName := "key.pem"
	privateKeyFile, err := os.ReadFile(keyLocation + privateKeyName)
	fmt.Println("Reading file at " + keyLocation + privateKeyName)
	if err != nil {
		fmt.Printf("Could not read private key at location %s\n", keyLocation+privateKeyName)
		return nil, err
	}
	fmt.Println(len(privateKeyFile))
	privatePem, _ := pem.Decode(privateKeyFile)
	if privatePem == nil {
		fmt.Printf(
			"Failed to decode private key file content for file at %s\n",
			keyLocation+privateKeyName,
		)
		return nil, errors.New("failed to decode private key file")
	}
	privateKeyAny, err := x509.ParsePKCS8PrivateKey(privatePem.Bytes)
	if err != nil {
		fmt.Println("Could not parse private key file. Content potentially malformed")
		return nil, err
	}
	privateKey, ok := privateKeyAny.(*rsa.PrivateKey)
	if !ok {
		fmt.Printf("The private key at location '%s' is not an RSA private key\n")
		return nil, errors.New("Incorrect key type")
	}

	publicKeyName := "cert.pem"
	publicKeyFile, err := os.ReadFile(keyLocation + publicKeyName)
	if err != nil {
		fmt.Printf("Could not read public key at location %s\n", keyLocation+privateKeyName)
		return nil, err
	}
	publicKeyPem, _ := pem.Decode(publicKeyFile)
	if publicKeyPem == nil {
		fmt.Printf(
			"Failed to decode public key file content for file at %s\n",
			keyLocation+publicKeyName,
		)
		return nil, errors.New("failed to decode public key file")
	}
	certificate, err := x509.ParseCertificate(publicKeyPem.Bytes)
	if err != nil {
		fmt.Println("Could not parse public key. Content potentially malformed")
		return nil, err
	}
	publicKey, ok := certificate.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("public key was not rsa public key")
	}

	privateKey.PublicKey = *publicKey

	return privateKey, nil
}
