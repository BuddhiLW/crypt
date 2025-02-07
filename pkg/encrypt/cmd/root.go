package cmd

import (
	"github.com/BuddhiLW/go-encrypt/pkg/encrypt"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/comp"
	"github.com/rwxrob/bonzai/vars"
)

var RootCmd = &bonzai.Cmd{
	Name: "encrypt",
	Long: "A CLI for Steganography using StegHide and Reed-Solomon encoding",
	Comp: comp.Cmds,
	Cmds: []*bonzai.Cmd{encrypt.EmbedCmd, encrypt.ExtractCmd, vars.Cmd, help.Cmd},
	Def:  help.Cmd,
}
