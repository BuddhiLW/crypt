package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/comp"
	"github.com/rwxrob/bonzai/vars"
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

// EncryptMessage encrypts a message using AES
func EncryptMessage(secret, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nil, nonce, []byte(secret), nil)
	finalCipher := append(nonce, ciphertext...)
	return base64.StdEncoding.EncodeToString(finalCipher), nil
}

var EncryptCmd = &bonzai.Cmd{
	Name:  "encrypt",
	Alias: "e",
	Short: `encrypt information`,
	Comp:  comp.Cmds,
	// Vars: bonzai.Vars{
	// 	{
	// 		K: EncryptDataVar,
	// 		V: `foo`,
	// 		E: EncryptEnv,
	// 		S: `data to be used and transformed in encryption steps`,
	// 		P: true,
	// 		// I: true,
	// 	},
	// },
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
		data := vars.Fetch(EncryptEnv, EncryptDataVar, "zoo fall")
		qrcode, err := CreateQRCodeBytes(data)
		vars.Data.Set(QRBinEnv, QRBinDataVar)

		switch args[0] {
		case EmbedCmd.Name:
			EmbedCmd.Do(x, args[1:]...)
		}
		return nil
	},
}

var EmbedCmd = &bonzai.Cmd{
	Name: `embed`,
	// Alias: `bin`,
	Comp:  comp.Cmds,
	Short: `dct (Discrete Cosine Transform) embedding`,
	Cmds: []*bonzai.Cmd{
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
		},
		{
			K: EmbeddedImagePathVar,
			V: `/tmp/embedded-image.png`,
			E: EmbeddedImagePathEnv,
			S: `path to output: embedded image`,
			P: true,
		},
	},
	Long: `
embed data, through DCT method, in an image.

Usages:
- encrypt text <input> <key> qrcode binary embed <input-image>;
- encrypt text <input> <key> qrcode binary embed <input-image> <output-image>;
`,
	Do: func(x *bonzai.Cmd, args ...string) error {
		fmt.Println("--- embedding ---")
		fmt.Println(args)
		// switch args[0] {
		// case EmbedCmd.Name:
		// 	EmbedCmd.Do(x, args[1:]...)
		// }
		return nil
	},
}
