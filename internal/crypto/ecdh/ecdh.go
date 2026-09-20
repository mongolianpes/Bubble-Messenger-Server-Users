package ecdh

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

func GenerateKeys(clientPubKeyBytesWithGarbage string) (string, string, error) {
	clientPubKeyBytes, err := base64.RawURLEncoding.DecodeString(clientPubKeyBytesWithGarbage)
	if err != nil {
		return "", "", errors.New("Invalid public key format")
	}

	clientPublicKey, err := ecdh.P256().NewPublicKey(clientPubKeyBytes)
	if err != nil {
		return "", "", errors.New("Invalid public key")
	}

	serverPrivateKey, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return "", "", errors.New("Failed to generate server key")
	}

	sharedSecret, err := serverPrivateKey.ECDH(clientPublicKey)
	if err != nil {
		return "", "", errors.New("Failed to compute secret")
	}

	serverPublicKey := base64.RawURLEncoding.EncodeToString(serverPrivateKey.PublicKey().Bytes())

	return fmt.Sprintf("%x", sharedSecret), serverPublicKey, nil
}
