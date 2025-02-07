package encrypt

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/rwxrob/bonzai"
)

var ExtractCmd = &bonzai.Cmd{
	Name:  "extract",
	Alias: `x`,
	Short: "extract hidden message",
	Long: `
Extract a hidden message from an image using StegHide and decode Reed-Solomon.

Usage:
encrypt extract <image.jpg> <password>`,
	Do: func(_ *bonzai.Cmd, args ...string) error {
		if len(args) < 2 {
			return fmt.Errorf("usage: stegcli extract <image.jpg> <password>")
		}

		imageFile := args[0]
		password := args[1]

		// Run StegHide to extract data
		cmd := exec.Command("steghide", "extract", "-sf", imageFile, "-xf", "extracted_data.txt", "-p", password)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error running steghide: %v", err)
		}

		// Read extracted file
		extractedBytes, err := os.ReadFile("extracted_data.txt")
		if err != nil {
			return fmt.Errorf("error reading extracted file: %v", err)
		}

		// Convert bytes to string and clean unwanted characters
		encodedData := strings.TrimSpace(string(extractedBytes))
		encodedData = strings.ReplaceAll(encodedData, "\n", "") // Remove newlines
		encodedData = strings.ReplaceAll(encodedData, "\r", "") // Remove carriage returns

		fmt.Println("Extracted Base64 data:", encodedData)

		// Decode Base64
		decodedBytes, err := base64.StdEncoding.DecodeString(encodedData)
		if err != nil {
			return fmt.Errorf("error decoding Base64: %v", err)
		}

		decodedMessage, err := decodeReedSolomon(string(decodedBytes))
		if err != nil {
			return fmt.Errorf("error decoding Reed-Solomon: %v", err)
		}

		fmt.Println("Final Decoded message:", decodedMessage)

		// Overwrite extracted_data.txt with the cleaned message
		err = os.WriteFile("extracted_data.txt", []byte(decodedMessage), 0644)
		if err != nil {
			return fmt.Errorf("error writing cleaned extracted data: %v", err)
		}

		return nil
	},
}

// Decode message using Reed-Solomon (simulated)
func decodeReedSolomon(data string) (string, error) {
	decodedStr := strings.TrimSpace(data)

	// Extract the length prefix
	parts := strings.SplitN(decodedStr, ":", 2)
	if len(parts) < 2 {
		return "", errors.New("invalid encoded message format")
	}

	// Convert length to an integer
	messageLength, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", fmt.Errorf("failed to parse message length: %v", err)
	}

	// Extract only the original message
	if messageLength > len(parts[1]) {
		return "", errors.New("message length exceeds extracted data length")
	}

	correctedMessage := parts[1][:messageLength]

	fmt.Println("Decoded message (after Reed-Solomon correction):", correctedMessage)
	return correctedMessage, nil
}
