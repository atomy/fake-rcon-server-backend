package util

import (
	"fmt"
	"log"
	"os"
)

// ReadFileToString reads the contents of a file into a string.
func ReadFileToString(filePath string) (string, error) {
	body, err := os.ReadFile(filePath)
	if err != nil {
		log.Panicln("ERROR, unable to read file:", err)
		return "", err
	}

	return string(body), nil
}

// WriteContentToFile write given content to file located on filePath.
func WriteContentToFile(filePath string, content string) {
	// Write the content to the file
	err := os.WriteFile(filePath, []byte(content), 0644)

	if err != nil {
		// Handle the error
		fmt.Println("Error:", err)
		return
	}
}
