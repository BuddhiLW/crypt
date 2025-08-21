package encrypt

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BuddhiLW/crypt/pkg/core"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/comp"
	"github.com/rwxrob/bonzai/vars"
)

const (
	MultiQREnv     = `MULTIQR_ENV`
	MultiQRDataVar = `multiqr-data`
	DecryptEnv     = `DECRYPT_ENV`
	DecryptDataVar = `decrypted-data`
)

// EnhancedMultiQRCmd provides advanced multi-QR functionality
var EnhancedMultiQRCmd = &bonzai.Cmd{
	Name:  "multiqr",
	Alias: "mq",
	Short: "enhanced multi-QR grid embedding with hash-based identification",
	Comp:  comp.Cmds,
	Cmds: []*bonzai.Cmd{
		MultiQREmbedCmd,
		MultiQRExtractCmd,
		MultiQRScanCmd,
		help.Cmd,
		vars.Cmd,
	},
	Long: `
Enhanced multi-QR functionality with hash-based file identification.

Features:
- Hash-based chunk identification for reliable reconstruction
- Directory scanning for automatic chunk discovery
- Metadata validation and integrity checking
- Compression resilience with High ECC

Usage:
- encrypt text <data> <key> multiqr embed <input.jpg> <output-dir>
- decrypt multiqr extract <metadata-file> <chunk-dir> <key>
- decrypt multiqr scan <directory> <key>
`,
	Do: func(x *bonzai.Cmd, args ...string) error {
		fmt.Printf("DEBUG: EnhancedMultiQRCmd called with args: %v\n", args)

		if len(args) == 0 {
			return fmt.Errorf("usage: multiqr <embed|extract|scan> [args...]")
		}

		switch args[0] {
		case MultiQREmbedCmd.Name:
			return MultiQREmbedCmd.Do(x, args[1:]...)
		case MultiQRExtractCmd.Name:
			return MultiQRExtractCmd.Do(x, args[1:]...)
		case MultiQRScanCmd.Name:
			return MultiQRScanCmd.Do(x, args[1:]...)
		default:
			return fmt.Errorf("unknown subcommand: %s", args[0])
		}
	},
}

// MultiQREmbedCmd embeds data as multiple QR codes with metadata
var MultiQREmbedCmd = &bonzai.Cmd{
	Name:  "embed",
	Short: "embed data as multiple QR codes with metadata",
	Long: `
Embed encrypted data as multiple QR codes with hash-based metadata.

Usage: encrypt text <data> <key> multiqr embed <input.jpg> <output-dir>

Creates:
- metadata.qr: Contains chunk information and hashes
- chunk_*.qr: Individual QR codes for each data chunk
`,
	Do: func(x *bonzai.Cmd, args ...string) error {
		fmt.Printf("DEBUG: MultiQREmbedCmd called with args: %v\n", args)

		if len(args) < 2 {
			return fmt.Errorf("usage: encrypt text <data> <key> multiqr embed <input.jpg> <output-dir>")
		}

		inputImage := args[0]
		outputDir := args[1]

		fmt.Printf("DEBUG: inputImage=%s, outputDir=%s\n", inputImage, outputDir)

		// Get encrypted data from vars
		encryptedData, err := vars.Get(EncryptDataVar, EncryptEnv)
		if err != nil || encryptedData == "" {
			return fmt.Errorf("no encrypted data found. Run 'encrypt text <input> <key>' first")
		}

		fmt.Printf("DEBUG: Got encrypted data, length=%d\n", len(encryptedData))

		// Create output directory
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		fmt.Printf("DEBUG: Created output directory\n")

		// Create service using factory
		factory := core.NewServiceFactory()
		service := factory.CreateSteganographyService(MultiQREnv)

		fmt.Printf("DEBUG: Created service\n")

		// Embed using enhanced multi-QR
		err = service.EmbedMultiQRWithMetadata(inputImage, outputDir, encryptedData, MultiQREnv)
		if err != nil {
			return fmt.Errorf("failed to embed multi-QR: %w", err)
		}

		fmt.Printf("✅ Enhanced multi-QR embedded successfully in: %s\n", outputDir)
		return nil
	},
}

// MultiQRExtractCmd extracts data from multiple QR codes using metadata
var MultiQRExtractCmd = &bonzai.Cmd{
	Name:  "extract",
	Short: "extract data from multiple QR codes using metadata",
	Long: `
Extract data from multiple QR codes using hash-based metadata.

Usage: decrypt multiqr extract <metadata-file> <chunk-dir> <key>

Requires:
- metadata.qr: Contains chunk information and hashes
- chunk_*.qr: Individual QR codes for each data chunk
`,
	Do: func(x *bonzai.Cmd, args ...string) error {
		if len(args) < 3 {
			return fmt.Errorf("usage: decrypt multiqr extract <metadata-file> <chunk-dir> <key>")
		}

		metadataFile := args[0]
		chunkDir := args[1]
		key := args[2]

		// Create service using factory
		factory := core.NewServiceFactory()
		service := factory.CreateSteganographyService(MultiQREnv)

		// Extract using enhanced multi-QR
		extractedData, err := service.ExtractMultiQRWithMetadata(metadataFile, chunkDir, key, MultiQREnv)
		if err != nil {
			return fmt.Errorf("failed to extract multi-QR: %w", err)
		}

		// Store extracted data
		if err := vars.Set(DecryptDataVar, extractedData, DecryptEnv); err != nil {
			return fmt.Errorf("failed to store extracted data: %w", err)
		}

		fmt.Printf("✅ Enhanced multi-QR extracted successfully: %s\n", extractedData)
		return nil
	},
}

// MultiQRScanCmd scans directory and automatically extracts data
var MultiQRScanCmd = &bonzai.Cmd{
	Name:  "scan",
	Short: "scan for QR files",
	Long: `
Scan directory for QR files and automatically extract multi-QR data.

Usage: decrypt multiqr scan <directory> <key>

Automatically:
- Finds metadata QR file
- Discovers chunk QR files
- Validates file integrity
- Extracts and decrypts data
`,
	Do: func(x *bonzai.Cmd, args ...string) error {
		if len(args) < 2 {
			return fmt.Errorf("usage: decrypt multiqr scan <directory> <key>")
		}

		directory := args[0]
		key := args[1]

		// Create service using factory
		factory := core.NewServiceFactory()
		service := factory.CreateSteganographyService(MultiQREnv)

		// Scan and extract
		extractedData, err := service.ScanAndExtractMultiQR(directory, key, MultiQREnv)
		if err != nil {
			return fmt.Errorf("failed to scan and extract multi-QR: %w", err)
		}

		// Store extracted data
		if err := vars.Set(DecryptDataVar, extractedData, DecryptEnv); err != nil {
			return fmt.Errorf("failed to store extracted data: %w", err)
		}

		fmt.Printf("✅ Multi-QR data scanned and extracted successfully: %s\n", extractedData)
		return nil
	},
}

// Helper function to get chunk files from directory
func getChunkFiles(dirPath string) ([]string, error) {
	var chunkFiles []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".qr") {
			// Skip metadata file
			if !strings.Contains(strings.ToLower(filepath.Base(path)), "metadata") {
				chunkFiles = append(chunkFiles, path)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	return chunkFiles, nil
}
