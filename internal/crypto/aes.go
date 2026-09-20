package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"
	// mathRand "math/rand"
)

// var symbolsForGarbage = map[string]int{
// 	"n": 4,
// 	"f": 6,
// 	"s": 15,
// 	"b": 15,
// 	"t": 58,
// 	"4": 45,
// 	"d": 42,
// 	"i": 32,
// 	"m": 11,
// 	"2": 23,
// 	"h": 49,
// 	"r": 2,
// 	"l": 37,
// 	"k": 53,
// 	"z": 20,
// 	"u": 40,
// 	"8": 53,
// 	"y": 44,
// 	"c": 46,
// 	"a": 27,
// 	"1": 14,
// 	"v": 34,
// 	"g": 10,
// 	"e": 41,
// 	"q": 32,
// }

// func ByteEncrypt(text, key []byte) []byte {
// 	// generate a new aes cipher using our 32 byte long key
// 	c, err := aes.NewCipher(key)
// 	// if there are any errors, handle them
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	// gcm or Galois/Counter Mode, is a mode of operation
// 	// for symmetric key cryptographic block ciphers
// 	// - https://en.wikipedia.org/wiki/Galois/Counter_Mode
// 	gcm, err := cipher.NewGCM(c)
// 	// if any error generating new GCM
// 	// handle them
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	// creates a new byte array the size of the nonce
// 	// which must be passed to Seal
// 	nonce := make([]byte, gcm.NonceSize())
// 	// populates our nonce with a cryptographically secure
// 	// random sequence
// 	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
// 		fmt.Println(err)
// 	}

// 	// here we encrypt our text using the Seal function
// 	// Seal encrypts and authenticates plaintext, authenticates the
// 	// additional data and appends the result to dst, returning the updated
// 	// slice. The nonce must be NonceSize() bytes long and unique for all
// 	// time, for a given key.
// 	return gcm.Seal(nonce, nonce, text, nil)
// }

// func ByteDecrypt(ciphertext, key []byte) []byte {
// 	// ciphertext, err := ioutil.ReadFile("myfile.data")
// 	// if our program was unable to read the file
// 	// print out the reason why it can't
// 	// if err != nil {
// 	// 	fmt.Println(err)
// 	// }

// 	c, err := aes.NewCipher(key)
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	gcm, err := cipher.NewGCM(c)
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	nonceSize := gcm.NonceSize()
// 	if len(ciphertext) < nonceSize {
// 		fmt.Println(err)
// 	}

// 	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
// 	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	return plaintext
// }

// func GetGarbage(key string) string {
// 	for numberKey, symbolKey := range key {
// 		for symbolMap, numberMap := range symbolsForGarbage {
// 			if string(symbolKey) == symbolMap {

// 				randomNumber := 11 - mathRand.Intn(10)
// 				garbage, _ := StringEncrypt(string(randomNumber), GenerateKey())
// 				key = key[:126] + garbage[:numberMap] + key[126:]

// 				encryptRandomNumber, _ := StringEncrypt(string(randomNumber), GenerateKey())
// 				encryptRandomNumber = encryptRandomNumber + encryptRandomNumber
// 				key = key[:numberKey+1] + encryptRandomNumber[:256-len(key)] + key[numberKey+1:]

// 				return key
// 			}
// 		}
// 	}
// 	randomNumber := 11 - mathRand.Intn(10)
// 	garbage, _ := StringEncrypt(string(randomNumber), GenerateKey())
// 	garbage = garbage + garbage + garbage
// 	return key[:126] + garbage + garbage[:256-len(key)] + key[126:]
// }

// func DelGarbage(key string) string {
// 	for numberKey, symbolKey := range key {
// 		for symbolMap, numberMap := range symbolsForGarbage {
// 			if string(symbolKey) == symbolMap {
// 				key = key[:numberKey+1] + key[numberKey+1+(256-128-numberMap):]

// 				key = key[:len(key)-2-numberMap] + key[len(key)-2:]
// 				return key
// 			} else if numberKey >= 126 {
// 				fmt.Print("op")
// 				return key[:126] + key[len(key)-2:]
// 			}
// 		}
// 	}
// 	return key
// }

// func GenerateKey() string {
// 	bytes := make([]byte, 32) //generate a random 32 byte key for AES-256
// 	if _, err := rand.Read(bytes); err != nil {
// 		panic(err.Error())
// 	}
// 	key := hex.EncodeToString(bytes) //encode key in bytes to string and keep as secret, put in a vault
// 	return key
// }

func StringEncrypt(stringToEncrypt []byte, keyString string) (encryptedString string, errEncrypt error) {
	//Since the key is in string, we need to convert decode it to bytes
	key, _ := hex.DecodeString(keyString)
	// plaintext := []byte(stringToEncrypt)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	//Create a new GCM - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	//https://golang.org/pkg/crypto/cipher/#NewGCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	//Create a nonce. Nonce should be from GCM
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	//Encrypt the data using aesGCM.Seal
	//Since we don't want to save the nonce somewhere else in this case, we add it as a prefix to the encrypted data. The first nonce argument in Seal is the prefix.
	ciphertext := aesGCM.Seal(nonce, nonce, stringToEncrypt, nil)
	return fmt.Sprintf("%x", ciphertext), nil
}

func StringEncryptByte(stringToEncrypt []byte, keyString string) (encryptedString []byte, errEncrypt error) {
	//Since the key is in string, we need to convert decode it to bytes
	key, _ := hex.DecodeString(keyString)
	// plaintext := []byte(stringToEncrypt)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}

	//Create a new GCM - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	//https://golang.org/pkg/crypto/cipher/#NewGCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return []byte{}, err
	}

	//Create a nonce. Nonce should be from GCM
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return []byte{}, err
	}

	//Encrypt the data using aesGCM.Seal
	//Since we don't want to save the nonce somewhere else in this case, we add it as a prefix to the encrypted data. The first nonce argument in Seal is the prefix.
	ciphertext := aesGCM.Seal(nonce, nonce, stringToEncrypt, nil)
	return ciphertext, nil
}

func StringDecrypt(encryptedString, keyString string) (decryptedString string, errEncrypt error) {

	key, _ := hex.DecodeString(keyString)
	enc, _ := hex.DecodeString(encryptedString)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	//Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	//Get the nonce size
	nonceSize := aesGCM.NonceSize()

	//Extract the nonce from the encrypted data
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	//Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s", plaintext), nil
}

func StringDecryptByte(enc []byte, keyString string) (decryptedString []byte, errEncrypt error) {
	key, _ := hex.DecodeString(keyString)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}

	//Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return []byte{}, err
	}

	//Get the nonce size
	nonceSize := aesGCM.NonceSize()

	//Extract the nonce from the encrypted data
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	//Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return []byte{}, err
	}

	return plaintext, nil
}

func FileStringEncrypt(stringToEncrypt []byte, pathToFileToEncrypt, userSalt, keyForServerDataBase string) error {
	encryptedString, err := StringEncrypt(stringToEncrypt, keyForServerDataBase)
	if err != nil {
		return err
	}

	if err = os.WriteFile(pathToFileToEncrypt, []byte(encryptedString), 0644); err != nil {
		return err
	}

	return nil
}

func FileStringDecrypt(pathToFileToDecrypt, userSalt, keyForServerDataBase string) (decryptedString string, errEncrypt error) {
	file, err := os.ReadFile(pathToFileToDecrypt)
	if err != nil {
		return "", err
	}

	decryptedString, err = StringDecrypt(string(file), keyForServerDataBase)
	if err != nil {
		return "", err
	}

	return decryptedString, nil
}

func checkDateWhenChangeFile(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Date(0, 0, 0, 0, 0, 0, 0, time.Local), err
	}

	modTime := info.ModTime()
	return modTime, nil
}
