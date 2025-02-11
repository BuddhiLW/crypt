package decrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"os"

	"github.com/liyue201/goqr"
	"github.com/rwxrob/bonzai"
	// "github.com/skip2/go-qrcode"
)

var DecryptCmd = &bonzai.Cmd{
	Name:  "decrypt",
	Alias: "e",
	Short: "decrypt image with embedded qrcode",
	Cmds: []*bonzai.Cmd{
		ImageCmd,
	},
}

var ImageCmd = &bonzai.Cmd{
	Name:  "image",
	Alias: "i",
	Short: "decrypt QR Code from an Image",
	Do: func(_ *bonzai.Cmd, args ...string) error {
		if len(args) < 2 {
			return fmt.Errorf("usage: decrypt image <input> <output-qrcode>")
		}

		inputImage := args[0]
		outputQR := args[1]

		fmt.Println("Extracting QR code from image:", inputImage)

		// **Step 1: Extract QR Code from JPEG**
		err := ExtractQRCodeFromJPEG(inputImage, outputQR)
		if err != nil {
			return fmt.Errorf("failed to extract QR code from image: %w", err)
		}

		fmt.Println("Extracted QR Code saved to:", outputQR)

		// **Step 2: Read QR Code from extracted image**
		qrText, err := ReadQRCode(outputQR)
		if err != nil {
			return fmt.Errorf("failed to read QR code: %w", err)
		}

		fmt.Println("Decoded QR Code Content:", qrText)

		// **Step 3 (Optional): Decrypt the extracted text**
		decryptedText, err := DecryptAES(qrText, "mysecurepassword")
		if err != nil {
			return fmt.Errorf("failed to decrypt text: %w", err)
		}
		fmt.Println("Decrypted Text:", decryptedText)

		return nil
	},
}

// ReadQRCode reads the text from a QR Code image
func ReadQRCode(filePath string) (string, error) {
	// Open QR code image
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open QR image: %w", err)
	}
	defer file.Close()

	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Decode QR code
	qrCodes, err := goqr.Recognize(img)
	if err != nil {
		return "", fmt.Errorf("failed to decode QR: %w", err)
	}

	// Return the first QR code found
	if len(qrCodes) > 0 {
		return string(qrCodes[0].Payload), nil
	}

	return "", fmt.Errorf("no QR code found in image")
}

// DecryptAES decrypts a Base64-encoded AES-GCM ciphertext using the provided key.
func DecryptAES(encryptedBase64, key string) (string, error) {
	// Decode Base64 input
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Convert key to 32 bytes (AES-256)
	keyBytes := deriveKey([]byte(key), 32)

	// Ensure ciphertext is large enough to contain nonce
	if len(ciphertext) < 12 {
		return "", errors.New("invalid ciphertext: too short")
	}

	// Extract nonce (first 12 bytes) and encrypted message
	nonce := ciphertext[:12]
	ciphertext = ciphertext[12:]

	// Create AES cipher block
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM cipher
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	// Decrypt the message
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return string(plaintext), nil
}

// deriveKey ensures the key is the required size (AES-256 = 32 bytes)
func deriveKey(key []byte, length int) []byte {
	derived := make([]byte, length)
	copy(derived, key)
	return derived
}
