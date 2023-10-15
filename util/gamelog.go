package util

import (
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
