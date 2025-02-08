package cmd

import (
	"github.com/BuddhiLW/crypt/pkg/decrypt"
	"github.com/BuddhiLW/crypt/pkg/encrypt"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/comp"
	"github.com/rwxrob/bonzai/vars"
)

var RootCmd = &bonzai.Cmd{
	Name: "crypt",
	Long: `
A CLI for steganography; focused on EAS, QRCode and DCT. Encryption and Decryption flows are supported.

Multiple methods fail to preserver the embbeded information, after even a little compression.

Here, a working "empirical" (opinionated?) workflow that survives heavy compression, is supported and proposed.
`,
	Comp: comp.Cmds,
	Cmds: []*bonzai.Cmd{encrypt.EncryptCmd, decrypt.DecryptCmd, vars.Cmd, help.Cmd},
}
