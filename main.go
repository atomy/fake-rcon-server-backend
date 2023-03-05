package main

import (
	"fmt"
	"os"
	"strings"
	"tf2-rcon/db"
	"tf2-rcon/gpt"
	"tf2-rcon/network"
	"tf2-rcon/utils"
	"time"
)

const teamSwitchMessage = "You have switched to team BLU and will receive 500 experience points at the end of the round for changing teams."

func main() {

	// Connect to the DB
	client := db.Connect()

	// Get the rcon host
	rconHost := network.DetermineRconHost()
	if rconHost == "Nothing" {
		utils.ErrorHandler(utils.ErrMissingRconHost)
	}

	fmt.Printf("Rcon Host: %s\n", rconHost)

	// Connect to the rcon host
	conn := network.RconConnect(rconHost)

	// Get the current player name
	res := network.RconExecute(conn, "name")
	playerName := strings.Split(res, " ")[2]
	playerName = strings.TrimSuffix(strings.TrimPrefix(playerName, `"`), `"`)

	// Get log path
	tf2LogPath := utils.LogPathDection()

	// Empty the log file
	utils.EmptyLog(tf2LogPath)

	// Tail the log
	t := utils.TailLog(tf2LogPath)

	selfCommandMap := map[string]func(args string){
		"!gpt": func(args string) {
			if !gpt.IsAvailable() {
				fmt.Println("!gpt is unavailable, cause env *OPENAI_APIKEY* is not set!")
				return
			}

			fmt.Println("!gpt - requesting:", args)
			response, err := gpt.Ask(args)
			fmt.Println("!gpt - requesting:", args, "- Response:", response)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error occured while gpt-communication:", err)
			} else {
				// Split the original string into chunks of 121 characters
				for i := 0; i < len(response); i += 121 {
					end := i + 121
					if end > len(response) {
						end = len(response)
					}
					chunk := response[i:end]
					fmt.Println(chunk)

					// on first run only delay 500 ms
					if i == 0 {
						time.Sleep(500 * time.Millisecond)
						network.RconExecute(conn, ("say \"GPT " + chunk + "\""))
					} else { // delay 1000 ms cause else we may get supressed
						time.Sleep(1000 * time.Millisecond)
						network.RconExecute(conn, ("say \"GPT " + chunk + "\""))
						break // only execute this once, we dont want to spam
					}
				}
			}
		},
		"!test": func(args string) {
			// 500 ms seems to work often, but not always, so lets be safe and use 1k
			time.Sleep(1000 * time.Millisecond)
			network.RconExecute(conn, ("say \"Test confirmed!\""))
		},
	}

	// Loop through the text of each received line
	for line := range t.Lines {
		// Run the status command when the lobby is updated or a player connects
		if strings.Contains(line.Text, "Lobby updated") || strings.Contains(line.Text, "connected") {
			network.RconExecute(conn, "status")
		} else if utils.Steam3IDMatcher(line.Text) && utils.PlayerNameMatcher(line.Text) { // Match all the players' steamID and name from the output of the status command
			// Convert Steam 32 ID to Steam 64 ID
			steamID := utils.Steam3IDToSteam64(utils.Steam3IDFindString(line.Text))

			// Find the player's userName
			userNameStrintToParse := strings.Fields(line.Text)
			userNameNotTrimmed := strings.Join(userNameStrintToParse[2:len(userNameStrintToParse)-5], " ")
			userName := strings.Trim(userNameNotTrimmed, "\"")

			// Add the player to the DB
			db.AddPlayer(client, steamID, userName)

			fmt.Println("SteamID: ", steamID, " UserName: ", userName)
		} else if len(line.Text) > len(playerName)+5 && line.Text[0:len(playerName)] == playerName { // that's my own say stuff
			// check if it starts with "!"
			if string(line.Text[len(playerName)+4]) == "!" {
				// command string, e.g. !gpt
				completeCommand := line.Text[len(playerName)+4:]

				// when command is too long, we skip
				if len(completeCommand) > 128 {
					continue
				}

				command, args := utils.GetCommandAndArgs(completeCommand)
				cmdFunc := selfCommandMap[command]

				if cmdFunc == nil {
					continue
				} else {
					cmdFunc(args)
				}
			} else if strings.Contains(line.Text, teamSwitchMessage) { // when you get team switched forcefully, thank gaben for the bonusxp!
				network.RconExecute(conn, ("say \"Thanks gaben for bonusxp!\""))
			}
		} else {
			fmt.Println("Unknown:", line.Text)
		}
	}
}

// // Function 3
// if strings.Contains(line.Text, "killed") &&
// 	strings.Contains(line.Text, "(crit)") &&
// 	strings.Contains(line.Text, playerName) {

// 	killer := strings.Split(line.Text, "killed")
// 	theKiller := killer[0]

// 	if theKiller == playerName {
// 		theKiller = ""
// 	}

// 	msg := utils.PickRandomMessage("crit")
// 	network.RconExecute(conn, ("say" + " " + "\"" + " " + msg + "\""))

// }

// if utils.Steam3IDMatcher(line.Text) {
// 	steamID := utils.Steam3IDToSteam64(utils.Steam3IDFindString(line.Text))
// 	fmt.Println(steamID)
// 	db.DBAddPlayer(client, steamID)
// }

// if utils.UserNameMatcher(line.Text) {
// 	userName := utils.UserNameFindString(line.Text)
// 	fmt.Println(userName)
// }
