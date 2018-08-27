package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Request struct {
	Method  string              `json:"method,omitempty"`
	Headers map[string][]string `json:"headers,omitempty"`
	Body    string              `json:"body,omitempty"`
	Query   url.Values          `json:"query,omitempty"`
}

const megaman = `
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
`

const (
	megahookDomain = "megahook.in"
	// megahookDomain = "localhost:8080"
	writeTimeout = time.Second * 30
)

var (
	megahookURL       = "http://" + megahookDomain
	megahookWebsocket = "ws://" + megahookDomain + "/ws"
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
	header.Add("Origin", megahookURL)
	conn, resp, err := dialer.Dial(megahookWebsocket, *header)
	if err != nil {
		fmt.Printf("Error connecting to server: %v\n%v\n", err, resp)
		return
	}
	defer conn.Close()

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

	fmt.Println(megaman)
	fmt.Printf("\nAll traffic from the following url: \n\n\t%v\n\nwill be forwarded to your local url:\n\n\t%v\n\n", string(message), local)

	// Start listening for requests
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("Server error: %v\n", err)
			break
		}

		if messageType != websocket.TextMessage {
			continue
		}

		r := bytes.NewBuffer(message)
		d := json.NewDecoder(r)
		req := &Request{}
		d.Decode(req)

		// Print the request out here so it's more
		// obvious to the user that something happened.
		if val, ok := req.Headers["Content-Type"]; ok && val[0] == "application/json" {
			o := make(map[string]interface{})
			json.Unmarshal([]byte(req.Body), &o)
			j, err := json.MarshalIndent(o, "", "  ")
			if err != nil {
				fmt.Printf("Error formatting JSON request: %v\n", err)
				continue
			}

			fmt.Printf("\n%v\n", string(j))
		} else {
			fmt.Printf("\n%v\n", *req)
		}

		client := &http.Client{}
		request, err := http.NewRequest(req.Method, local, bytes.NewBuffer([]byte(req.Body)))
		if err != nil {
			fmt.Printf("Error creating new request: %v\n", err)
			continue
		}

		for key, headers := range req.Headers {
			for _, value := range headers {
				request.Header.Add(key, value)
			}
		}

		request.URL.RawQuery = req.Query.Encode()

		_, err = client.Do(request)
		if err != nil {
			fmt.Printf("Error doing request: %v\n", err)
			continue
		}
	}
}
