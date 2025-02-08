package decrypt

import (
	"fmt"

	"github.com/rwxrob/bonzai"
)

var DecryptCmd = &bonzai.Cmd{
	Name:  "decrypt",
	Alias: "e",
	Short: "encrypt text and process through QR, binary, and image embedding",
	Cmds: []*bonzai.Cmd{
		ImageCmd,
	},
}

var ImageCmd = &bonzai.Cmd{
	Name:  "image",
	Alias: "i",
	Short: "decrypt Image",
	Do: func(_ *bonzai.Cmd, args ...string) error {
		if len(args) < 2 {
			return fmt.Errorf("usage: encrypt image <input> ...")
		}
		fmt.Println("Decrypt image:", args[0])

		// extract binary
		// tranform binary in qrcode
		// read qrcode
		// decrypt text
		// return decrypted-text
		return nil
	},
}
