package encrypt

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/rwxrob/bonzai"
)

var EmbedCmd = &bonzai.Cmd{
	Name:  "embed",
	Alias: `e`,
	Short: "embed hidden message",
	Long: `
Embed a hidden message in an image using StegHide and Reed-Solomon

Usage:
encrypt embed <input.jpg> <output.jpg> <message> <password>`,
	Do: func(_ *bonzai.Cmd, args ...string) error {
		if len(args) < 4 {
			return fmt.Errorf("usage: stegcli embed <input.jpg> <output.jpg> <message> <password>")
		}

		inputFile := args[0]
		outputFile := args[1]
		message := args[2]
		password := args[3]

		// Apply Reed-Solomon encoding
		fmt.Println("Original message:", message)

		encodedMsg, err := encodeReedSolomon(message)
		if err != nil {
			fmt.Printf("Reed-Solomon encoding error: %v\n", err)
			return fmt.Errorf("error encoding message: %v", err)
		}

		fmt.Println("Encoded message:", encodedMsg)

		// Ensure hidden_data.txt is created
		err = os.WriteFile("hidden_data.txt", []byte(encodedMsg), 0644)
		if err != nil {
			fmt.Printf("Error writing hidden_data.txt: %v\n", err)
			return fmt.Errorf("error writing hidden_data.txt: %v", err)
		}

		fmt.Println("Message successfully written to hidden_data.txt")

		// Run StegHide to embed data
		cmd := exec.Command("steghide", "embed", "-cf", inputFile, "-ef", "hidden_data.txt", "-sf", outputFile, "-p", password)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error running steghide: %v", err)
		}

		fmt.Println("Data embedded successfully in", outputFile)
		return nil
	},
}

// Encode message using Reed-Solomon (simulated)
func encodeReedSolomon(data string) (string, error) {
	if len(data) == 0 {
		return "", errors.New("input data is empty")
	}

	// Store the length of the original message
	messageLength := strconv.Itoa(len(data))

	// Append the message length before encoding
	formattedData := messageLength + ":" + data

	// Encode as Base64 for better embedding
	encoded := base64.StdEncoding.EncodeToString([]byte(formattedData))

	fmt.Println("Reed-Solomon (simulated) Encoded:", encoded)
	return encoded, nil
}
