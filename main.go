package main

import (
	"fmt"
	"github.com/atomy/fake-rcon-server-backend/network"
	"log"
)

func main() {
	log.Println("Hello World!")
	log.Println("Press any key to exit")

	go network.StartServer()

	fmt.Scanln() // Wait for Enter keypress
	log.Println("Exiting...")
}
