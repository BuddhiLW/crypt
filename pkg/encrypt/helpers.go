package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	// "github.com/BuddhiLW/crypt/pkg/encrypt"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/comp"
	"github.com/rwxrob/bonzai/vars"
)

const (
	EncryptEnv     = `ENCRYPT_ENV`
	EncryptDataVar = `encrypted-data`

	QREnv     = `QR_ENV`
	QRDataVar = `qr-data`

	QRBinEnv     = `QR_BIN_ENV`
	QRBinDataVar = `qr-bin-data`

	EmbeddedImagePathEnv = `EMBEDDED_IMAGE_PATH_ENV`
	EmbeddedImagePathVar = `embedded-image-path`

	DCTEnv         = `DCT_ENV`
	DCTStrategyVar = `dct-strategy`

	QRSizeVar = `qr-size`
)

// **🔹 Encrypt AES (Ensure Output is Correct)**
func EncryptMessage(secret, key string) (string, error) {
	keyBytes := deriveKey([]byte(key), 32)

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nil, nonce, []byte(secret), nil)
	finalCipher := append(nonce, ciphertext...)

	// Encode to Base64
	base64Cipher := base64.StdEncoding.EncodeToString(finalCipher)

	// Debug: Print the base64 result
	fmt.Println("🔐 Encrypted Base64 Output:", base64Cipher)

	return base64Cipher, nil
}

// deriveKey ensures the key is always a valid AES key size (AES-256 = 32 bytes)
func deriveKey(key []byte, length int) []byte {
	derived := make([]byte, length)
	copy(derived, key)
	return derived
}

var EncryptCmd = &bonzai.Cmd{
	Name:  "encrypt",
	Alias: "e",
	Short: `encrypt information`,
	Comp:  comp.Cmds,
	Cmds: []*bonzai.Cmd{
		TextCmd,
		FileCmd,
		StrategyCmd,
		help.Cmd,
		vars.Cmd,
	},
}

var TextCmd = &bonzai.Cmd{
	Name:  "text",
	Alias: "t",
	Short: "encrypt text using AES",
	Cmds: []*bonzai.Cmd{
		QRCodeCmd,
		help.Cmd,
		vars.Cmd,
	},
	Comp:  comp.Cmds,
	Usage: `encrypt text <input> <key>`,
	Long: `
Encrypt text using AES encryption.

Usage: encrypt text <input> <key>
Where key must be longer than 15 characters.
`,
	Do: func(x *bonzai.Cmd, args ...string) error {
		if len(args) < 2 {
			return fmt.Errorf("usage: encrypt text <input> <key>")
		}
		if len(args[1]) < 16 {
			return fmt.Errorf("key (password) must be greater or equal to 16 characters")
		}

		encrypted, err := EncryptMessage(args[0], args[1])
		if err != nil {
			return err
		}

		if err := vars.Set(EncryptDataVar, encrypted, EncryptEnv); err != nil {
			return fmt.Errorf("failed to store encrypted data: %w", err)
		}

		fmt.Printf("DEBUG TextCmd: Just stored encrypted data, length=%d, data='%.50s...'\n", len(encrypted), encrypted)

		if len(args) > 2 {
			fmt.Println(args[2:])
			// if (args[2] == QRCodeCmd.Name) {
			QRCodeCmd.Do(x, args[3:]...)
			// }
		}

		return nil
	},
}

// FileCmd encrypts a file's contents using AES and stores the base64 ciphertext in vars
var FileCmd = &bonzai.Cmd{
	Name:  "file",
	Alias: "f",
	Comp:  comp.Cmds,
	Short: "encrypt file contents using AES",
	Cmds: []*bonzai.Cmd{
		QRCodeCmd,
		vars.Cmd.AsHidden(),
		help.Cmd.AsHidden(),
	},
	Long: `
encrypt file using AES.

Usage: encrypt file <path> <key>; in which |key|>=16 characters
`,
	Do: func(x *bonzai.Cmd, args ...string) error {
		if len(args) < 2 {
			return fmt.Errorf("usage: encrypt file <path> <key>")
		}
		if len(args[1]) < 16 {
			return fmt.Errorf("key (password) must be greater or equal to 16 characters")
		}

		// Read file
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		encrypted, err := EncryptMessage(string(data), args[1])
		if err != nil {
			return err
		}

		if err := vars.Set(EncryptDataVar, encrypted, EncryptEnv); err != nil {
			return fmt.Errorf("failed to store encrypted data: %w", err)
		}

		if len(args) > 2 {
			fmt.Println(args[2:])
			QRCodeCmd.Do(x, args[3:]...)
		}

		return nil
	},
}

var QRCodeCmd = &bonzai.Cmd{
	Name:  `qrcode`,
	Alias: `qr`,
	Comp:  comp.Cmds,
	Short: `qrcode related commands`,
	Cmds: []*bonzai.Cmd{
		// CreateQRCmd,
		CreateQRBinaryCmd,
		// DecodeQRCmd,
		vars.Cmd.AsHidden(),
		help.Cmd.AsHidden(),
	},
	Vars: bonzai.Vars{
		{
			K: QRDataVar,
			V: `foo`,
			E: QREnv,
			S: `binary data representing a qrcode`,
			P: true,
			// I: true,
		},
	},
	Long: `
create QRCode holding data. Either just create one and return it, or create and transform in a binary data.

Usages:
- encrypt text <input> <key> qrcode create;
- encrypt text <input> <key> qrcode create binary;
- encrypt text <input> <key> qrcode binary;
`,
	Do: func(x *bonzai.Cmd, args ...string) error {

		// default:
		data := vars.Fetch(EncryptEnv, EncryptDataVar, "zoo fall")
		// generate PNG qrcode with ECC fallback
		_, err := WriteQRCodeWithFallback(data, 256, "/tmp/qr.png")
		if err != nil {
			return fmt.Errorf("failed to generate QR: %w", err)
		}
		fmt.Println("Wrote qrcode to /tmp/qr.png")

		switch args[0] {
		case CreateQRBinaryCmd.Name:
			CreateQRBinaryCmd.Do(x, args[1:]...)
		}
		return nil
	},
}

var CreateQRBinaryCmd = &bonzai.Cmd{
	Name:  `binary`,
	Alias: `bin`,
	Comp:  comp.Cmds,
	Short: `qrcode related commands`,
	Cmds: []*bonzai.Cmd{
		EmbedCmd,
		vars.Cmd.AsHidden(),
		help.Cmd.AsHidden(),
	},
	Vars: bonzai.Vars{
		{
			K: QRBinDataVar,
			V: `foo`,
			E: QRBinEnv,
			S: `binary data representing a qrcode`,
			P: true,
			// I: true,
		},
	},
	Long: `
created QRCode then transform it in binary data.

Usages:
- encrypt text <input> <key> qrcode binary embed <input-image>;
`,
	Do: func(x *bonzai.Cmd, args ...string) error {
		fmt.Println("--- binary ---")
		fmt.Println(args)

		// Create QRCode binary (which can be converted in a png etc.) from EncryptDataVar
		// data := vars.Fetch(EncryptEnv, EncryptDataVar, "zoo fall")
		// qrcode, err := CreateQRCodeBytes(data)
		// vars.Data.Set(QRBinDataVar, qrcode)

		switch args[0] {
		case EmbedCmd.Name:
			EmbedCmd.Do(x, args[1:]...)
		}
		return nil
	},
}

var StrategyCmd = &bonzai.Cmd{
	Name:  `strategy`,
	Alias: `s`,
	Short: `set DCT embedding strategy (single; multi coef)`,
	Usage: `strategy <single|multi>`,
	Long: `
Sets the DCT embedding strategy for steganography:

- single-coefficient: Original approach (1 bit per DCT block)
- multi-coefficient: Enhanced approach (4 bits per DCT block, 4x capacity)

The multi-coefficient strategy provides 4x the capacity but may be slightly
more detectable. Use 'multi' for larger payloads that need High ECC.
`,
	Do: func(x *bonzai.Cmd, args ...string) error {
		if len(args) < 1 {
			current, _ := vars.Get(DCTStrategyVar, DCTEnv)
			// if current == "" {
			// 	current = "single-coefficient"
			// }
			fmt.Printf("Current DCT strategy: %s\n", current)
			fmt.Println("Usage: strategy <single|multi>")
			return nil
		}

		strategy := args[0]
		switch strategy {
		case "single", "single-coefficient":
			if err := vars.Set(DCTStrategyVar, "single-coefficient", DCTEnv); err != nil {
				return fmt.Errorf("failed to set strategy: %w", err)
			}
			fmt.Println("DCT strategy set to: single-coefficient (1x capacity)")
		case "multi", "multi-coefficient":
			if err := vars.Set(DCTStrategyVar, "multi-coefficient", DCTEnv); err != nil {
				return fmt.Errorf("failed to set strategy: %w", err)
			}
			fmt.Println("DCT strategy set to: multi-coefficient (4x capacity)")
		default:
			return fmt.Errorf("invalid strategy '%s'. Use 'single' or 'multi'", strategy)
		}

		return nil
	},
}

// DirectDCTCmd embeds encrypted data directly into DCT coefficients (no QR overhead)
var DirectDCTCmd = &bonzai.Cmd{
	Name:  `direct`,
	Short: `Direct DCT embedding (high capacity, no QR overhead)`,
	Comp:  comp.Cmds,
	Cmds: []*bonzai.Cmd{
		vars.Cmd.AsHidden(),
		help.Cmd.AsHidden(),
	},
	Long: `
Embed encrypted data directly into JPEG DCT coefficients without QR code overhead.
This method can achieve 5K-10K capacity by using multiple DCT coefficients per 8x8 block.

Usage: encrypt text <data> <key> qrcode binary direct <input.jpg> <output.jpg>
`,
	Do: func(x *bonzai.Cmd, args ...string) error {
		fmt.Println("--- Direct DCT Embedding (High Capacity) ---")

		if len(args) < 2 {
			return fmt.Errorf("usage: direct <input-image> <output-image>")
		}

		inputImage := args[0]
		outputImage := args[1]

		// Get encrypted data
		encryptedData, varErr := vars.Get(EncryptDataVar, EncryptEnv)
		if varErr != nil || encryptedData == "" {
			return fmt.Errorf("no encrypted data found - run encrypt first")
		}

		fmt.Printf("Direct DCT: embedding %d bytes without QR overhead\n", len(encryptedData))

		// Embed directly into DCT coefficients
		err := EmbedDataDirectlyInDCT(inputImage, outputImage, encryptedData)
		if err != nil {
			return fmt.Errorf("direct DCT embedding failed: %w", err)
		}

		fmt.Printf("Successfully embedded %d bytes directly into DCT coefficients: %s\n", len(encryptedData), outputImage)
		return nil
	},
}

var EmbedCmd = &bonzai.Cmd{
	Name:  `embed`,
	Comp:  comp.Cmds,
	Short: `DCT (Discrete Cosine Transform) embedding`,
	Cmds: []*bonzai.Cmd{
		DirectDCTCmd,
		vars.Cmd.AsHidden(),
		help.Cmd.AsHidden(),
	},
	Vars: bonzai.Vars{
		{
			K: QRBinDataVar,
			V: `foo`,
			E: QRBinEnv,
			S: `binary data representing a QR code`,
			P: true,
		},
		{
			K: EmbeddedImagePathVar,
			V: `/tmp/embedded-image.jpg`,
			E: EmbeddedImagePathEnv,
			S: `path to output: embedded image`,
			P: true,
		},
	},
	Long: `
Embed a QR code as binary data into an image using DCT.

Usages:
- encrypt text <input> <key> qrcode binary embed <input-image>;
- encrypt text <input> <key> qrcode binary embed <input-image> <output-image>;
`,
	Do: func(x *bonzai.Cmd, args ...string) error {
		// Check if first argument is a subcommand
		if len(args) > 0 && args[0] == DirectDCTCmd.Name {
			return DirectDCTCmd.Do(x, args[1:]...)
		}

		fmt.Println("--- Embedding QR Code into JPEG ---")
		fmt.Println(args)

		// Ensure input image is provided
		if len(args) < 1 {
			return fmt.Errorf("missing input image path")
		}
		inputImage := args[0]

		qrData, varErr := vars.Get(EncryptDataVar, EncryptEnv)
		if varErr != nil || qrData == "" {
			qrData = "zoo fall" // fallback
			fmt.Printf("DEBUG: Failed to get qrData from vars (error: %v), using fallback\n", varErr)
		} else {
			fmt.Printf("DEBUG: Retrieved qrData from vars using new method, length=%d, first 50 chars: '%.50s'\n", len(qrData), qrData)
		}
		// Get QR binary data
		// qrData := vars.Fetch(QRBinEnv, QRBinDataVar, "default_qr_data")
		// if qrData == "" {
		// 	return fmt.Errorf("failed to fetch QR binary data")
		// }

		// Set output image path
		outputImage := vars.Fetch(EmbeddedImagePathEnv, EmbeddedImagePathVar, "/tmp/embedded-image.jpg")
		if len(args) > 1 {
			outputImage = args[1] // Override output path if provided
		}

		// Embed QR code into the JPEG using DCT
		payloadSize := len(qrData) // Size of the Base64 encrypted data
		fmt.Printf("DEBUG: qrData length: %d, qrData: '%.50s...'\n", len(qrData), qrData)
		err := EmbedQRCodeInJPEG(inputImage, outputImage, qrData, payloadSize)
		if err != nil {
			return fmt.Errorf("failed to embed QR code in JPEG: %w", err)
		}

		fmt.Println("QR code embedded in:", outputImage)
		return nil
	},
}
