package main

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Get AES-256 key from environment variable
func getKey() ([]byte, error) {
	key := os.Getenv("AES_KEY")
	if len(key) != 32 {
		return nil, fmt.Errorf("invalid key length: AES-256 requires a 32-byte key")
	}
	return []byte(key), nil
}

// Decrypt a single file and remove the encrypted version
func decryptFile(filename string, key []byte) error {
	ciphertext, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading encrypted file: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("error creating cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("error creating GCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return fmt.Errorf("invalid ciphertext size")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	// Remove ".enc" to restore original filename
	decryptedFilename := strings.TrimSuffix(filename, ".enc")
	err = os.WriteFile(decryptedFilename, plaintext, 0644)
	if err != nil {
		return fmt.Errorf("error writing decrypted file: %w", err)
	}

	// Delete encrypted file
	err = os.Remove(filename)
	if err != nil {
		return fmt.Errorf("error deleting encrypted file: %w", err)
	}

	fmt.Println("Decrypted:", decryptedFilename)
	return nil
}

// Recursively decrypt all encrypted files in a folder
func decryptFolder(folder string, key []byte) {
	filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, ".enc") {
			decryptFile(path, key)
		}
		return nil
	})
}

func main() {
	key, err := getKey()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Define major user folders
	folders := []string{
		filepath.Join(os.Getenv("USERPROFILE"), "Documents"),
		filepath.Join(os.Getenv("USERPROFILE"), "Downloads"),
		filepath.Join(os.Getenv("USERPROFILE"), "Desktop"),
		filepath.Join(os.Getenv("USERPROFILE"), "Pictures"),
		filepath.Join(os.Getenv("USERPROFILE"), "Videos"),
		filepath.Join(os.Getenv("USERPROFILE"), "Music"),
	}

	// Decrypt all files in target folders
	for _, folder := range folders {
		fmt.Println("Decrypting files in:", folder)
		decryptFolder(folder, key)
	}

	fmt.Println("All files decrypted successfully!")
}
