package client

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/gorilla/websocket"
)

func StartClient(host string, port int, name string) {
	addr := fmt.Sprintf("ws://%s:%d/ws?name=%s", host, port, name)
	fmt.Printf("Connecting to %s as %s...\n", addr, name)

	// connect to server
	conn, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		log.Fatal("dial: ", err)
	}
	defer conn.Close()

	// independent go routine to read messages
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("read error: ", err)
				return
			}
			fmt.Printf("\n[Broadcast] %s\n>", string(msg))
		}
	}()

	// Read user input & send to server
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(">")
	for scanner.Scan() {
		text := scanner.Text()
		if text == "/quit" {
			fmt.Println("Exiting...")
			return
		}
		// Write the message to server
		err := conn.WriteMessage(websocket.TextMessage, []byte(text))
		if err != nil {
			log.Println("Write error", err)
			return
		}
		fmt.Print(">")
	}

	if err := scanner.Err(); err != nil {
		log.Println("stdin error:", err)
	}
}
