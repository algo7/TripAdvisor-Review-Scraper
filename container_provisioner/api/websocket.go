package api

import (
	"bufio"
	"container_provisioner/containers"
	"log"

	"github.com/gofiber/websocket/v2"
)

// wsHandler streams logs to the WebSocket connection
func wsHandler(c *websocket.Conn, containerId string) {
	defer c.Close()

	reader := containers.TailLog(c.Params("id"))

	// Create a scanner to read logs line by line
	scanner := bufio.NewScanner(reader)

	// Stream logs to the WebSocket connection
	for scanner.Scan() {
		err := c.WriteMessage(websocket.TextMessage, scanner.Bytes())
		if err != nil {
			log.Printf("Error sending log to WebSocket: %s\n", err)
			return
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading logs: %s\n", err)
		return
	}
}
