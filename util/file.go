package util

import (
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
