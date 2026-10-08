package ecdh

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"testing"
)

func TestGenerateKeys(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		clientPrivateKey, err := ecdh.P256().GenerateKey(rand.Reader)
		if err != nil {
			t.Fatalf("failed to generate client key: %v", err)
		}

		clientPubKeyBytes := clientPrivateKey.PublicKey().Bytes()
		clientPubKeyB64 := base64.RawURLEncoding.EncodeToString(clientPubKeyBytes)

		// 2. Вызываем тестируемую функцию
		sharedSecretHex, serverPubKeyB64, err := GenerateKeys(clientPubKeyB64)
		if err != nil {
			t.Fatalf("GenerateKeys returned error: %v", err)
		}

		if sharedSecretHex == "" {
			t.Fatal("shared secret is empty")
		}
		if serverPubKeyB64 == "" {
			t.Fatal("server public key is empty")
		}

		// 3. Декодируем публичный ключ сервера
		serverPubKeyBytes, err := base64.RawURLEncoding.DecodeString(serverPubKeyB64)
		if err != nil {
			t.Fatalf("invalid server public key encoding: %v", err)
		}

		serverPublicKey, err := ecdh.P256().NewPublicKey(serverPubKeyBytes)
		if err != nil {
			t.Fatalf("invalid server public key: %v", err)
		}

		// 4. Вычисляем shared secret на стороне клиента
		clientSharedSecret, err := clientPrivateKey.ECDH(serverPublicKey)
		if err != nil {
			t.Fatalf("client ECDH failed: %v", err)
		}

		expectedHex := hex.EncodeToString(clientSharedSecret)

		// 5. Сравниваем
		if sharedSecretHex != expectedHex {
			t.Errorf("shared secrets do not match\nwant: %s\ngot:  %s", expectedHex, sharedSecretHex)
		}
	})

	t.Run("invalidbase64", func(t *testing.T) {
		_, _, err := GenerateKeys("это-не-base64!!!")
		if err == nil {
			t.Fatal("expected error for invalid base64")
		}
		if !errors.Is(err, errors.New("Invalid public key format")) && err.Error() != "Invalid public key format" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("invalid public key", func(t *testing.T) {
		garbage := base64.RawURLEncoding.EncodeToString([]byte("not-a-valid-p256-public-key"))
		_, _, err := GenerateKeys(garbage)
		if err == nil {
			t.Fatal("expected error for invalid public key")
		}
		if err.Error() != "Invalid public key" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("empty", func(t *testing.T) {
		_, _, err := GenerateKeys("")
		if err == nil {
			t.Fatal("expected error for empty input")
		}
	})
}
