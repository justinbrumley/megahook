package main

import (
	"fmt"
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
}
