package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

const (
	connHost = "localhost"
	connPort = 6667
	connType = "tcp"
)

func main() {
	l, err := net.Listen(connType, fmt.Sprintf("%s:%d", connHost, connPort))
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	// Close the listener when the application closes.
	defer l.Close()
	fmt.Printf("Listening on %s:%d\n", connHost, connPort)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		// Handle connections in a new goroutine.
		go handleConnection(conn)
	}
}

func splitCommand(s, sep string) (string, string) {
	data := strings.Split(s, sep)
	if len(data) == 1 {
		return data[0], ""
	}
	return data[0], data[1]
}

// Handles incoming requests.
func handleConnection(conn net.Conn) {
	// Close the connection when you're done with it.
	clientNick := ""
	quit := false
	buf := bufio.NewReader(conn)
	for !quit {
		message, err := buf.ReadString('\n')
		// Read the incoming connection into the buffer.
		if err != nil {
			fmt.Println("Error reading: ", err.Error())
			quit = true
		}

		//messages := strings.Split(string(buf[:reqLen]), "\r\n")
		message = strings.Split(message, "\r")[0]
		response := ""

		subCommand, data := splitCommand(message, " ")
		if len(message) == 0 {
			continue
		}
		fmt.Printf("Parsing command message %s\n", message)
		switch subCommand {
		case "CAP":
			subCommand, _ := splitCommand(data, " ")
			switch subCommand {
			case "LS":
				response += ":tmi.twitch.tv CAP * LS :twitch.tv/commands twitch.tv/tags\r\n"
				//response += ":tmi.twitch.tv 410 tmi.twitch.tv :Invalid CAP command. TMoohI always runs twitch.tv/commands and twitch.tv/tags\r\n"
				break
			case "END":
				if len(clientNick) == 0 {
					response += "SQUIT 127.0.0.1 :No NICK sent\r\n"
					fmt.Printf("Disconnecting client, didn't send a NICK")
					quit = true
					break
				}
				response += fmt.Sprintf(":tmi.twitch.tv 001 %s :Welcome to ttv bouncer!\r\n", clientNick)
				response += fmt.Sprintf(":tmi.twitch.tv 002 %s :Your host is tmi.twitch.tv\r\n", clientNick)
				response += fmt.Sprintf(":tmi.twitch.tv 003 %s :This server is rather new\r\n", clientNick)
				response += fmt.Sprintf(":tmi.twitch.tv 004 %s :-\r\n", clientNick)
				response += fmt.Sprintf(":tmi.twitch.tv 375 %s :-\r\n", clientNick)
				response += fmt.Sprintf(":tmi.twitch.tv 372 %s :You are in a maze of twisty passages.\r\n", clientNick)
				response += fmt.Sprintf(":tmi.twitch.tv 376 %s :>\r\n", clientNick)
				break
			default:
				response += ":tmi.twitch.tv 410 tmi.twitch.tv :Invalid CAP command\r\n"
			}
		case "QUIT":
			response += "SQUIT\r\n"
			fmt.Printf("Client disconnected %s\n", clientNick)
			quit = true
			break
		case "NICK":
			clientNick = data
			break
		case "PING":
			response += fmt.Sprintf(":tmi.twitch.tv PONG :tmi.twitch.tv :%d\r\n", time.Now().Unix())
			break
		}

		// Send a response back to person contacting us.
		conn.Write([]byte(response))
	}
	conn.Close()
}
