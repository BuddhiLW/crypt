package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	// "github.com/BuddhiLW/crypt/pkg/encrypt"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/comp"
	"github.com/rwxrob/bonzai/vars"
	"github.com/skip2/go-qrcode"
	// "os"
	// "os/exec"
	// "strconv"
	// "github.com/skip2/go-qrcode"
	// "errors"
	// "image"
	// "image/png"
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
		vars.Cmd.AsHidden(),
		help.Cmd.AsHidden(),
	},
}

var TextCmd = &bonzai.Cmd{
	Name:  "text",
	Alias: "t",
	Comp:  comp.Cmds,
	Short: "encrypt text using AES",
	Cmds: []*bonzai.Cmd{
		QRCodeCmd,
		vars.Cmd.AsHidden(),
		help.Cmd.AsHidden(),
	},
	Vars: bonzai.Vars{
		{
			K: EncryptDataVar,
			V: `foo`,
			E: EncryptEnv,
			S: `data to be used and transformed in encryption steps`,
			P: true,
			// I: true,
		},
	},
	Long: `
encrypt text using AES.

Usage: encrypt text <input> <key>; in which |key|>15 characters
`,
	Do: func(x *bonzai.Cmd, args ...string) error {
		if len(args) < 2 {
			return fmt.Errorf("usage: encrypt text <input> <key>")
		}
		if len(args[1]) < 16 {
			return fmt.Errorf("key (password) must be greater or equal to 16 characters")
		}

		encrypted, err := EncryptMessage(args[0], args[1])
		vars.Data.Set(EncryptDataVar, encrypted)
		// s := vars.Fetch(EncryptDataVar, EncryptDataVar, "zoo fall")
		if err != nil {
			return err
		}

		if len(args) > 2 {
			fmt.Println(args[2:])
			// if (args[2] == QRCodeCmd.Name) {
			QRCodeCmd.Do(x, args[3:]...)
			// }
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
		// qrcode, err := CreateQRCodeBytes(data)
		// DOING: generate PNG qrcode
		err := qrcode.WriteFile(data, qrcode.Highest, 256, "/tmp/qr.png")
		fmt.Print("Wrote qrcode to /tmp/qr.png")
		if err != nil {
			return err
		}

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

var EmbedCmd = &bonzai.Cmd{
	Name:  `embed`,
	Comp:  comp.Cmds,
	Short: `DCT (Discrete Cosine Transform) embedding`,
	Cmds: []*bonzai.Cmd{
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
		fmt.Println("--- Embedding QR Code into JPEG ---")
		fmt.Println(args)

		// Ensure input image is provided
		if len(args) < 1 {
			return fmt.Errorf("missing input image path")
		}
		inputImage := args[0]

		qrData := vars.Fetch(EncryptEnv, EncryptDataVar, "zoo fall")
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
		err := EmbedQRCodeInJPEG(inputImage, outputImage, qrData)
		if err != nil {
			return fmt.Errorf("failed to embed QR code in JPEG: %w", err)
		}

		fmt.Println("QR code embedded in:", outputImage)
		return nil
	},
}
