package encrypt_test

import (
	"fmt"
	"testing"

	"github.com/BuddhiLW/crypt/pkg/encrypt"
	"github.com/rwxrob/bonzai"
)

func TestTextCmdStructure(t *testing.T) {
	t.Parallel()

	cmd := encrypt.TextCmd

	// Test command structure
	if cmd.Name != "text" {
		t.Errorf("expected command name 'text', got '%s'", cmd.Name)
	}

	if cmd.Alias != "t" {
		t.Errorf("expected alias 't', got '%s'", cmd.Alias)
	}

	if cmd.Short == "" {
		t.Error("command should have a short description")
	}

	if cmd.Usage == "" {
		t.Error("command should have usage information")
	}

	if cmd.Long == "" {
		t.Error("command should have a long description")
	}
}

func TestFileCmdStructure(t *testing.T) {
	t.Parallel()

	cmd := encrypt.FileCmd

	// Test command structure
	if cmd.Name != "file" {
		t.Errorf("expected command name 'file', got '%s'", cmd.Name)
	}

	if cmd.Alias != "f" {
		t.Errorf("expected alias 'f', got '%s'", cmd.Alias)
	}

	if cmd.Short == "" {
		t.Error("command should have a short description")
	}

	if cmd.Long == "" {
		t.Error("command should have a long description")
	}
}

func TestQRCodeCmdStructure(t *testing.T) {
	t.Parallel()

	cmd := encrypt.QRCodeCmd

	// Test command structure
	if cmd.Name != "qrcode" {
		t.Errorf("expected command name 'qrcode', got '%s'", cmd.Name)
	}

	if cmd.Alias != "qr" {
		t.Errorf("expected alias 'qr', got '%s'", cmd.Alias)
	}

	if cmd.Short == "" {
		t.Error("command should have a short description")
	}

	if cmd.Long == "" {
		t.Error("command should have a long description")
	}

	// Test subcommands
	if len(cmd.Cmds) == 0 {
		t.Error("qrcode command should have subcommands")
	}
}

func TestEmbedCmdStructure(t *testing.T) {
	t.Parallel()

	cmd := encrypt.EmbedCmd

	// Test command structure
	if cmd.Name != "embed" {
		t.Errorf("expected command name 'embed', got '%s'", cmd.Name)
	}

	if cmd.Short == "" {
		t.Error("command should have a short description")
	}

	if cmd.Long == "" {
		t.Error("command should have a long description")
	}

	// Test subcommands
	if len(cmd.Cmds) == 0 {
		t.Error("embed command should have subcommands")
	}
}

func TestMultiQRCmdStructure(t *testing.T) {
	t.Parallel()

	cmd := encrypt.MultiQRCmd

	// Test command structure
	if cmd.Name != "multiqr" {
		t.Errorf("expected command name 'multiqr', got '%s'", cmd.Name)
	}

	if cmd.Alias != "mq" {
		t.Errorf("expected alias 'mq', got '%s'", cmd.Alias)
	}

	if cmd.Short == "" {
		t.Error("command should have a short description")
	}

	if cmd.Long == "" {
		t.Error("command should have a long description")
	}
}

func TestEncryptCmdHierarchy(t *testing.T) {
	t.Parallel()

	cmd := encrypt.EncryptCmd

	// Test root command structure
	if cmd.Name != "encrypt" {
		t.Errorf("expected command name 'encrypt', got '%s'", cmd.Name)
	}

	if cmd.Alias != "e" {
		t.Errorf("expected alias 'e', got '%s'", cmd.Alias)
	}

	if cmd.Short == "" {
		t.Error("command should have a short description")
	}

	// Test subcommands exist
	subcommandNames := make(map[string]bool)
	for _, subcmd := range cmd.Cmds {
		subcommandNames[subcmd.Name] = true
	}

	expectedSubcommands := []string{"text", "file", "embed"}
	for _, expected := range expectedSubcommands {
		if !subcommandNames[expected] {
			t.Errorf("missing expected subcommand: %s", expected)
		}
	}
}

func TestCommandValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cmd      *bonzai.Cmd
		validate func(*bonzai.Cmd) error
	}{
		{
			name: "text command validation",
			cmd:  encrypt.TextCmd,
			validate: func(cmd *bonzai.Cmd) error {
				if cmd.Do == nil {
					return fmt.Errorf("command should have a Do function")
				}
				return nil
			},
		},
		{
			name: "file command validation",
			cmd:  encrypt.FileCmd,
			validate: func(cmd *bonzai.Cmd) error {
				if cmd.Do == nil {
					return fmt.Errorf("command should have a Do function")
				}
				return nil
			},
		},
		{
			name: "qrcode command validation",
			cmd:  encrypt.QRCodeCmd,
			validate: func(cmd *bonzai.Cmd) error {
				if cmd.Do == nil {
					return fmt.Errorf("command should have a Do function")
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.validate(tt.cmd); err != nil {
				t.Errorf("command validation failed: %v", err)
			}
		})
	}
}
