# flera

## Introduction

flera is a lightweight networking library for Go, designed to simplify the creation of multiplayer applications.
It leverages both TCP and UDP for reliable and fast communication, making it suitable for a wide range of applications, from real-time games to chat servers.

## Features

- **Dual-Protocol Support:** Utilizes TCP for reliable data transmission and UDP for fast, potentially lossy updates.
- **Easy-to-Use API:** Simplifies server and client setup with intuitive functions for connecting, sending, and receiving data.
- **Handler-Based Communication:** Allows registration of handler functions for specific message types, making it easy to manage different network events.
- **Concurrency-Safe:** Designed with concurrency in mind, ensuring safe operation in multi-client environments.

## Installation
```
bash
go get github.com/VincentBrodin/flera
```
## Usage

### Server Setup

To set up a server with flera, you need to create a server instance, register handlers for incoming messages, and start listening for connections.
Here's a simplified example based on the TicTacToe server (`example/tictactoe/server/game_server.go`):
```
go
package main

import (
	"flera/pkg/server"
	"fmt"
)

const (
	UPDATE_STATE uint32 = 2
	MOUSE_POS    uint32 = 3
)

func main() {
	s := server.New()

	s.Register(UPDATE_STATE, UpdateState)
	s.Register(MOUSE_POS, MousePos)

	s.OnConn = OnConn
	s.OnDisConn = OnDisConn

	fmt.Println(s.Start(":2489"))
}

func UpdateState(s *server.Server, connId uint32, data []byte) error {
	// Handle state updates from clients
	fmt.Printf("Received state update from client %d: %v\n", connId, data)
	// Example: Broadcast the updated state to all clients
	return s.BroadcastSafe(UPDATE_STATE, data)
}

func MousePos(s *server.Server, connId uint32, data []byte) error {
	// Handle mouse position updates from clients for fast updates
	return s.BroadcastFast(MOUSE_POS, append([]byte{byte(connId)}, data...))
}

func OnConn(s *server.Server, connId uint32) {
	fmt.Printf("Client %d connected\n", connId)
	// Additional connection handling logic (e.g., assigning teams)
}

func OnDisConn(s *server.Server, connId uint32) {
	fmt.Printf("Client %d disconnected\n", connId)
	// Additional disconnection handling logic
}
```
This example demonstrates:

- Creating a new server instance using `server.New()`.
- Registering handlers for `UPDATE_STATE` and `MOUSE_POS` messages using `s.Register()`.
- Defining handler functions (`UpdateState`, `MousePos`) to process incoming data.
- Using `s.BroadcastSafe()` to send reliable updates (e.g., game state) to all clients via TCP.
- Using `s.BroadcastFast()` to send fast updates (e.g., mouse position) via UDP.
- Setting `OnConn` and `OnDisConn` event handlers to manage client connections and disconnections.
- Starting the server and listening for connections on port 2489 using `s.Start(":2489")`.

### Client Setup

To set up a client, you need to create a client instance, register handlers for incoming messages, and connect to the server. Here's a simplified example based on the TicTacToe client (`example/tictactoe/client/game_client.go`):
```
go
package main

import (
	"flera/pkg/client"
	"fmt"
)

const (
	SET_TEAM     uint32 = 1
	UPDATE_STATE uint32 = 2
	MOUSE_POS    uint32 = 3
)

func main() {
	c := client.New()

	c.Register(SET_TEAM, SetTeam)
	c.Register(UPDATE_STATE, UpdateState)
	c.Register(MOUSE_POS, MousePos)

	if err := c.Connect(":2489"); err != nil {
		fmt.Println("Failed to connect:", err)
		return
	}
	fmt.Println("Connected to server!")

	// Client logic for sending updates and interacting with the server
	// Example: Send a state update
	// c.SendSafe(UPDATE_STATE, []byte("some data"))

	// Keep the client running (e.g., with a game loop or by listening for input)
	select {} 
}

func SetTeam(c *client.Client, data []byte) error {
	fmt.Printf("Received team assignment: %v\n", data)
	// Handle team assignment from the server
	return nil
}

func UpdateState(c *client.Client, data []byte) error {
	fmt.Printf("Received state update: %v\n", data)
	// Handle game state updates from the server
	return nil
}

func MousePos(c *client.Client, data []byte) error {
	fmt.Printf("Received mouse position update: %v\n", data)
	// Handle mouse position updates from other clients
	return nil
}
```
This example demonstrates:

- Creating a new client instance using `client.New()`.
- Registering handlers for `SET_TEAM`, `UPDATE_STATE`, and `MOUSE_POS` messages using `c.Register()`.
- Defining handler functions (`SetTeam`, `UpdateState`, `MousePos`) to process incoming data.
- Connecting to the server on port 2489 using `c.Connect(":2489")`.
- Sending data to the server using `c.SendSafe()` (for reliable updates via TCP).  `c.SendFast()` is also available for UDP.
- The `select {}` statement keeps the client running indefinitely. In a real application, you would replace this with your main loop or interaction logic.

### Example - TicTacToe

For a complete example, refer to the TicTacToe implementation in the `example/tictactoe` directory. It showcases how to build a simple multiplayer game using flera, including:

- Server-side game logic for managing the game state and player turns.
- Client-side rendering and user input handling.
- Communication between the server and clients for updating the game state and player actions.

## Use Cases

flera can be used to build various multiplayer applications, including:

- **Real-time Games:** Games requiring low-latency updates and interaction, such as action games or simulations.
- **Chat Applications:** Servers for handling text or voice communication between multiple users.
- **Collaborative Tools:** Applications where multiple users need to interact with a shared data set in real-time.
- **Multiplayer Servers:** Servers for any application requiring real-time communication and data synchronization between clients.

## Contributing

Contributions are welcome! Please submit pull requests or open issues on the project's GitHub repository.
