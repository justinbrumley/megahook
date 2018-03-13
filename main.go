package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Missing local webhook url.\n\nmegahook <local_webhook_url> [name]")
		return
	}

	local := os.Args[1]
	name := ""

	if len(os.Args) > 2 {
		name = os.Args[2]
	}

	fmt.Printf("Connecting to server with local url %v and name %v\n", local, name)

	dialer := &websocket.Dialer{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	header := &http.Header{}
	header.Add("Origin", "http://localhost")
	conn, _, err := dialer.Dial("ws://localhost:8080/ws", *header)
	if err != nil {
		fmt.Printf("Error connecting to server: %v\n", err)
		return
	}

	// Write the chosen name to the server to get back a url to use as webhook
	err = conn.WriteMessage(websocket.TextMessage, []byte(name))
	if err != nil {
		fmt.Printf("Error sending message to server: %v\n", err)
		return
	}

	// Receive full url back from server
	_, message, err := conn.ReadMessage()
	if err != nil {
		fmt.Printf("Error reading message from server: %v\n", err)
		return
	}

	fmt.Printf(`
░░░░░░░░░░▄▄█▀▀▄░░░░
░░░░░░░░▄█████▄▄█▄░░░░
░░░░░▄▄▄▀██████▄▄██░░░░
░░▄██░░█░█▀░░▄▄▀█░█░░░▄▄▄▄
▄█████░░██░░░▀▀░▀░█▀▀██▀▀▀█▀▄
█████░█░░▀█░▀▀▀▀▄▀░░░███████▀
░▀▀█▄░██▄▄░▀▀▀▀█▀▀▀▀▀░▀▀▀▀
░▄████████▀▀▀▄▀░░░░
██████░▀▀█▄░░░█▄░░░░
░▀▀▀▀█▄▄▀░██████▄░░░░
░░░░░░░░░█████████░░░░
	`)
	fmt.Printf("\nAll traffic from the following url: \n\n\t%v\n\nwill be forwarded to your local url:\n\n\t%v\n\n", string(message), local)

	// Now we wait
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("Server error: %v\n", err)
			break
		}

		if messageType == websocket.TextMessage {
			fmt.Printf("Received message from server: \n%v\n", string(message))
		}
	}
}
