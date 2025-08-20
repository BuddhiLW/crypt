package decrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"os"
	"strings"

	"github.com/BuddhiLW/crypt/pkg/encrypt"
	"github.com/liyue201/goqr"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/comp"
	"github.com/rwxrob/bonzai/vars"
)

const (
	DecryptEnv     = `DECRYPT_ENV`
	DecryptDataVar = `decrypted-data`
	DefaultQRPath  = `/tmp/extracted_qr.png`
)

// **🔹 Main Decrypt Command**
var DecryptCmd = &bonzai.Cmd{
	Name:  "decrypt",
	Alias: "d",
	Short: "decrypt embedded QR Code from an image",
	Comp:  comp.Cmds,
	Cmds: []*bonzai.Cmd{
		ImageCmd,
		DirectCmd,
		vars.Cmd.AsHidden(),
		help.Cmd.AsHidden(),
	},
}

// **🔹 Decrypt Image Command**
var ImageCmd = &bonzai.Cmd{
	Name:  "image",
	Alias: "i",
	Short: "extract QR Code from an image",
	Comp:  comp.Cmds,
	Cmds: []*bonzai.Cmd{
		ExtractCmd,
		vars.Cmd.AsHidden(),
		help.Cmd.AsHidden(),
	},
	Vars: bonzai.Vars{
		{
			K: DecryptDataVar,
			V: `foo`,
			E: DecryptEnv,
			S: `decrypted data extracted from an image`,
			P: true,
		},
	},
	Do: func(x *bonzai.Cmd, args ...string) error {
		if len(args) < 1 {
			return fmt.Errorf("usage: decrypt image <input-image> [<output-qrcode>, defaults to /tmp/extracted_qr.png]")
		}

		inputImage := args[0]
		outputQR := DefaultQRPath
		if len(args) > 1 && args[1] != ExtractCmd.Name {
			outputQR = args[1] // Override default output if provided
		}

		fmt.Println("Extracting QR code from image:", inputImage)

		// **Step 1: Extract QR Code from JPEG**
		err := ExtractQRCodeFromJPEG(inputImage, outputQR)
		if err != nil {
			return fmt.Errorf("failed to extract QR code from image: %w", err)
		}

		fmt.Println("Extracted QR Code saved to:", outputQR)

		if args[1] == ExtractCmd.Name {
			// **Step 2: Decrypt Text from QR Code**
			return ExtractCmd.Do(ExtractCmd, args[2:]...)
		}

		if args[2] == ExtractCmd.Name {
			// **Step 2: Decrypt Text from QR Code**
			return ExtractCmd.Do(ExtractCmd, args[3:]...)
		}

		return nil
	},
}

// **🔹 Extract Command for Decrypting Text**
var ExtractCmd = &bonzai.Cmd{
	Name:  "extract",
	Alias: "x",
	Short: "decrypt text from extracted QR code",
	Comp:  comp.Cmds,
	Cmds: []*bonzai.Cmd{
		TextCmd,
		vars.Cmd.AsHidden(),
		help.Cmd.AsHidden(),
	},
	Do: func(x *bonzai.Cmd, args ...string) error {
		// Ensure at least one argument is provided (the decryption key)
		if len(args) < 2 {
			fmt.Println("error: provide a password (key) to be used to decrypt the message.")
			return fmt.Errorf("usage: decrypt image <input> [<output>] extract text <key>")
		}

		if args[0] == TextCmd.Name {
			return TextCmd.Do(TextCmd, args[1:]...)
		}
		return nil
	},
}

// **🔹 Extract Text Command**
var TextCmd = &bonzai.Cmd{
	Name:  "text",
	Alias: "t",
	Short: "decrypt text using AES",
	Do: func(x *bonzai.Cmd, args ...string) error {
		if len(args) < 1 {
			return fmt.Errorf("usage: decrypt image <input> [<output>] extract text <key>")
		}

		key := args[0]
		fmt.Println("key", key)
		outputQR := DefaultQRPath // Use default unless overridden

		// Read the QR Code
		qrText, err := ReadQRCode(outputQR)
		if err != nil {
			return fmt.Errorf("failed to read QR code: %w", err)
		}

		// Debugging: Print extracted QR code content
		fmt.Println("🛠️ Extracted QR Code (Base64):", qrText)

		// **Decrypt the extracted text**
		decryptedText, err := DecryptAES(qrText, key)
		if err != nil {
			return fmt.Errorf("failed to decrypt text: %w", err)
		}

		// Cache decrypted text
		vars.Data.Set(DecryptDataVar, decryptedText)
		fmt.Println("Decrypted Text:", decryptedText)

		return nil
	},
}

// **🔹 Read QR Code from Image (Fix Base64 Formatting)**
func ReadQRCode(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open QR image: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	qrCodes, err := goqr.Recognize(img)
	if err != nil {
		return "", fmt.Errorf("failed to decode QR: %w", err)
	}

	if len(qrCodes) > 0 {
		decodedString := string(qrCodes[0].Payload)

		// **Ensure proper Base64 formatting**
		cleanedString := normalizeBase64(decodedString)

		return cleanedString, nil
	}

	return "", fmt.Errorf("no QR code found in image")
}

// **🔹 Normalize Base64 to Ensure Correct Padding**
func normalizeBase64(encoded string) string {
	// Remove unwanted characters
	encoded = strings.TrimSpace(encoded)
	encoded = strings.ReplaceAll(encoded, "\n", "")
	encoded = strings.ReplaceAll(encoded, "\r", "")

	// Ensure proper Base64 padding
	missingPadding := len(encoded) % 4
	if missingPadding > 0 {
		encoded += strings.Repeat("=", 4-missingPadding)
	}

	return encoded
}

// **🔹 Decrypt AES (Reversible from Encrypt)**
func DecryptAES(encryptedBase64, key string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	keyBytes := deriveKey([]byte(key), 32)

	if len(ciphertext) < 12 {
		return "", errors.New("invalid ciphertext: too short")
	}

	nonce := ciphertext[:12]
	ciphertext = ciphertext[12:]

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return string(plaintext), nil
}

// **🔹 Direct DCT Decryption Command**
var DirectCmd = &bonzai.Cmd{
	Name:  "direct",
	Short: "extract and decrypt data from DCT coefficients",
	Comp:  comp.Cmds,
	Cmds: []*bonzai.Cmd{
		vars.Cmd.AsHidden(),
		help.Cmd.AsHidden(),
	},
	Long: `
Extract and decrypt data that was embedded directly into JPEG DCT coefficients 
without QR code overhead (high capacity method).

Usage: decrypt direct <image> <key>
`,
	Do: func(x *bonzai.Cmd, args ...string) error {
		fmt.Println("--- Direct DCT Extraction & Decryption ---")

		if len(args) < 2 {
			return fmt.Errorf("usage: direct <image> <key>")
		}

		imagePath := args[0]
		key := args[1]

		if len(key) < 16 {
			return fmt.Errorf("key (password) must be greater or equal to 16 characters")
		}

		// Extract encrypted data directly from DCT coefficients
		fmt.Printf("Extracting data from: %s\n", imagePath)
		encryptedData, err := encrypt.ExtractDataDirectlyFromDCT(imagePath)
		if err != nil {
			return fmt.Errorf("direct DCT extraction failed: %w", err)
		}

		fmt.Printf("Extracted %d bytes of encrypted data\n", len(encryptedData))

		// Decrypt the extracted data
		decryptedData, err := DecryptAES(encryptedData, key)
		if err != nil {
			return fmt.Errorf("decryption failed: %w", err)
		}

		fmt.Printf("Successfully decrypted %d bytes\n", len(decryptedData))
		fmt.Printf("Decrypted data:\n%s\n", decryptedData)

		// Store decrypted data in vars for potential further use
		if err := vars.Set(DecryptDataVar, decryptedData, DecryptEnv); err != nil {
			fmt.Printf("Warning: failed to store decrypted data: %v\n", err)
		}

		return nil
	},
}

// **🔹 Ensures Key is Always 32 Bytes**
func deriveKey(key []byte, length int) []byte {
	derived := make([]byte, length)
	copy(derived, key)
	return derived
}
