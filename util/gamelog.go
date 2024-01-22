package util

import (
	"log"
	"os"
)

// WriteToGameLogFile Write given string to game-logfile.
func WriteToGameLogFile(content string) {
	gameLogPath := os.Getenv("GAME_LOGPATH")

	// If env is empty, skip this.
	if gameLogPath == "" {
		return
	}

	WriteContentToFile(gameLogPath, content)
}

// TruncateGameLogFile Truncate game-log for a fresh start.
func TruncateGameLogFile() {
	gameLogPath := os.Getenv("GAME_LOGPATH")

	// If env is empty, skip this.
	if gameLogPath == "" {
		return
	}

	// Open the file for writing
	file, err := os.OpenFile(gameLogPath, os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	// Truncate the file to 100 bytes
	err = file.Truncate(100)
	if err != nil {
		log.Fatal(err)
	}
}
