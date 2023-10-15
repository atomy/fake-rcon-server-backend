package network

import (
	"bufio"
	"encoding/binary"
	"github.com/atomy/fake-rcon-server-backend/util"
	"log"
	"net"
)

// Define constants for packet types.
const (
	SERVERDATA_AUTH           = 3
	SERVERDATA_AUTH_RESPONSE  = 2
	SERVERDATA_RESPONSE_VALUE = 0
)

// StartServer Startup the TCP server.
func StartServer() {
	// Define the address and port to listen on
	address := "192.168.2.40"
	port := "27015"

	// Create a TCP listener
	listener, err := net.Listen("tcp", address+":"+port)
	if err != nil {
		log.Println("Error creating listener:", err)
		return
	}

	defer func(listener net.Listener) {
		_ = listener.Close()
	}(listener)

	log.Println("TCP server is listening on " + address + ":" + port)

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		// Handle the incoming connection in a goroutine
		go handleConnection(conn)
	}
}

// Handle new connection and do the communication.
func handleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// RCON password (change this to your server's RCON password)
	rconPassword := "123"

	// Read and authenticate the RCON password
	var passwordSize, packetID, packetType int32
	err := binary.Read(reader, binary.LittleEndian, &passwordSize)
	if err != nil {
		log.Println("Error reading password size:", err)
		return
	}

	// Read the packet ID and type for the password packet.
	err = binary.Read(reader, binary.LittleEndian, &packetID)
	if err != nil {
		log.Println("Error reading packet id:", err)
		return
	}

	err = binary.Read(reader, binary.LittleEndian, &packetType)
	if err != nil {
		log.Println("Error reading packet type:", err)
		return
	}

	if SERVERDATA_AUTH != packetType {
		log.Printf("Error, received wrong package-type: %d (expecting: %d)\n", packetType, SERVERDATA_AUTH)
		return
	}

	// Read the password.
	password := make([]byte, passwordSize-10) // Subtract 10 as explained in the protocol

	_, err = reader.Read(password)
	if err != nil {
		log.Println("Error reading password:", err)
		return
	}

	// Read remaining bytes and discard.
	_, err = reader.ReadByte()
	if err != nil {
		// Ignore.
	}
	_, err = reader.ReadByte()
	if err != nil {
		// Ignore.
	}

	// When the server receives an auth request, it will respond with an empty SERVERDATA_RESPONSE_VALUE.
	sendRCONResponse(writer, packetID, SERVERDATA_RESPONSE_VALUE, "")

	// Check supplied password.
	if string(password) != rconPassword {
		log.Printf("RCON password incorrect (given password: %s)\n", string(password))
		// If authentication was successful, the ID assigned by the request. If auth failed, -1 (0xFF FF FF FF)
		// Send an AUTH_RESPONSE with failure
		sendRCONResponse(writer, -1, SERVERDATA_AUTH_RESPONSE, "")
		return
	}

	log.Println("RCON password authenticated")

	// If authentication was successful, respond with an AUTH_RESPONSE
	// Send an AUTH_RESPONSE with success
	sendRCONResponse(writer, packetID, SERVERDATA_AUTH_RESPONSE, "")

	for {
		// Read the full packet
		var packetSize, packetID, packetType int32

		// Read packetSize bytes from wire.
		err := binary.Read(reader, binary.LittleEndian, &packetSize)
		if err != nil {
			log.Println("Error reading packetSize:", err)
			return
		} else {
			log.Printf("Read packetSize is: %d\n", packetSize)
		}

		// Read packetID bytes from wire.
		err = binary.Read(reader, binary.LittleEndian, &packetID)
		if err != nil {
			log.Println("Error reading packetID:", err)
			return
		} else {
			log.Printf("Read packetID is: %d\n", packetID)
		}

		// Read packetType bytes from wire.
		err = binary.Read(reader, binary.LittleEndian, &packetType)
		if err != nil {
			log.Println("Error reading packetType:", err)
			return
		} else {
			log.Printf("Read packetType is: %d\n", packetType)
		}

		// Read the packet body based on the size
		bodySize := int(packetSize) - 10 // Subtract 10 as explained in the protocol
		body := make([]byte, bodySize)
		_, err = reader.Read(body)

		if err != nil {
			log.Println("Error reading packet body:", err)
			return
		}

		// Read remaining bytes and discard.
		_, err = reader.ReadByte()
		if err != nil {
			// Ignore.
		}
		_, err = reader.ReadByte()
		if err != nil {
			// Ignore.
		}

		// Process the packet
		switch packetType {
		case 2: // SERVERDATA_EXECCOMMAND
			log.Println("Received RCON command:", string(body))
			response := getResponseForCommand(string(body))

			// Create a trimmed version for logging (up to 32 characters)
			responseForLog := response
			if len(responseForLog) > 32 {
				responseForLog = responseForLog[:32]
				responseForLog = responseForLog + "..."
			}
			log.Printf("Answering command '%s' with response: %s\n", string(body), responseForLog)

			// Send answer over the wire.
			sendRCONResponse(writer, packetID, SERVERDATA_RESPONSE_VALUE, response)
		default:
			log.Println("Unknown packet type:", packetType)
		}
	}
}

// Get string response for given command and return it.
func getResponseForCommand(command string) string {
	// Process the command
	switch command {
	case "status":
		fileContent, err := util.ReadFileToString("data/status-response.txt")
		if err != nil {
			// Handle the error
			log.Println("Error while calling ReadFileToString(), Error:", err)
			return ""
		}

		return fileContent
	case "name":
		fileContent, err := util.ReadFileToString("data/name-response.txt")
		if err != nil {
			// Handle the error
			log.Println("Error while calling ReadFileToString(), Error:", err)
			return ""
		}

		return fileContent
	default:
		log.Panicf("ERROR, unknown command: '%s'\n", command)
		return ""
	}
}

// Send Rcon-paket with given parameters over the network.
func sendRCONResponse(writer *bufio.Writer, packetID, packetType int32, response string) {
	responseSize := int32(len(response) + 10)

	err := binary.Write(writer, binary.LittleEndian, responseSize)
	if err != nil {
		log.Println("Unable to write responseSize:", err)
		return
	}

	err = binary.Write(writer, binary.LittleEndian, packetID)
	if err != nil {
		log.Println("Unable to write packetID:", err)
		return
	}

	err = binary.Write(writer, binary.LittleEndian, packetType)
	if err != nil {
		log.Println("Unable to write packetType:", err)
		return
	}

	// If response is an empty string, we have to explicitly write a null-byte.
	if len(response) <= 0 {
		err := writer.WriteByte(byte(0x00))
		if err != nil {
			log.Println("Unable to writer.WriteByte(0x00):", err)
			return
		}
	} else {
		// Append a null byte (0x00) to the string
		responseWithNull := response + string(byte(0)) + string(byte(0))
		_, err = writer.WriteString(responseWithNull)
		if err != nil {
			log.Println("Unable to writer.WriteString():", err)
			return
		}
	}

	err = writer.Flush()
	if err != nil {
		log.Println("Unable to flush:", err)
		return
	}
}
