package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Request struct {
	Method     string              `json:"method,omitempty"`
	Headers    map[string][]string `json:"headers,omitempty"`
	Body       string              `json:"body,omitempty"`
	Query      url.Values          `json:"query,omitempty"`
	StatusCode int                 `json:"status_code,omitempty"`
}

type Response struct {
	Headers    map[string][]string `json:"headers,omitempty"`
	Body       string              `json:"body,omitempty"`
	StatusCode int                 `json:"status_code,omitempty"`
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

	// TODO: Move somewhere else
	version = "0.0.3"
)

var (
	megahookURL       = "http://" + megahookDomain
	megahookWebsocket = "ws://" + megahookDomain + "/ws"
)

var showHelp bool
var name string
var showVersion bool

const helpUsage = "Megahook help"
const nameUsage = "Name to use for megahook URL"
const versionUsage = "Megahook version"

func parseFlags() {
	flag.BoolVar(&showHelp, "help", false, helpUsage)
	flag.BoolVar(&showHelp, "h", false, helpUsage+" (shorthand)")

	flag.BoolVar(&showVersion, "version", false, versionUsage)
	flag.BoolVar(&showVersion, "v", false, versionUsage+" (shorthand)")

	flag.StringVar(&name, "name", "", nameUsage)
	flag.StringVar(&name, "n", "", nameUsage+" (shorthand)")

	flag.Parse()
}

func main() {
	parseFlags()

	if showVersion {
		fmt.Printf("%v\n", version)
		return
	}

	args := flag.Args()

	if len(args) == 0 {
		showHelp = true
	}

	if showHelp {
		fmt.Printf(strings.TrimSpace(`
Name

	megahook - Util for forwarding webhooks to your local environment

Usage

	megahook [options] <webhook_url> [name]

Options/Flags

	-h --help
		What you are seeing now

	-v --version
		Version of megahook client

	-n --name
		Name to use for the Megahook URL. If taken or not provided, a random uuid v4 will be used.
		`) + "\n")

		return
	}

	local := args[0]

	if name == "" && len(args) > 1 {
		name = args[1]
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

		response, err := client.Do(request)
		if err != nil {
			fmt.Printf("Error doing request: %v\n", err)

			// Send response back to server so it can reach the originator
			res := &Response{
				Body:       "",
				StatusCode: response.StatusCode,
			}

			if err = conn.WriteJSON(res); err != nil {
				fmt.Printf("Error sending response to server: %v\n", err)
			}

			continue
		}

		reader := bufio.NewReader(response.Body)
		body := ""

		for {
			s, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					body += s
					break
				}

				fmt.Printf("Error reading from body: %v\n", err)
				break
			}

			body += s
		}

		// Send response back to server so it can reach the originator
		res := &Response{
			Headers:    map[string][]string(response.Header),
			Body:       body,
			StatusCode: response.StatusCode,
		}

		if err = conn.WriteJSON(res); err != nil {
			fmt.Printf("Error sending response to server: %v\n", err)
			continue
		}
	}
}
