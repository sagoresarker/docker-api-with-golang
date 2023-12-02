package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gorilla/websocket"

	"github.com/google/shlex"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	latestContainerID string
)

// type WebSocketMessage struct {
// 	Operation   string   `json:"operation"`
// 	ContainerID string   `json:"containerID"`
// 	Command     []string `json:"command"`
// }

func execCommand(cli *client.Client, containerID string, command []string, options string, ws *websocket.Conn) error {
	ctx := context.Background()

	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          command,
	}

	// Check if the options include "-i" or "-t"
	if strings.Contains(options, "i") {
		execConfig.AttachStdin = true
	}
	if strings.Contains(options, "t") {
		execConfig.Tty = true
	}

	// Check if the options include "-d"
	if strings.Contains(options, "d") {
		execConfig.Detach = true
	}

	execIDResp, err := cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return err
	}

	attachResp, err := cli.ContainerExecAttach(ctx, execIDResp.ID, types.ExecStartCheck{})
	if err != nil {
		return err
	}
	defer attachResp.Close()

	// Create a channel to signal when the command is done
	done := make(chan error)

	// Start a goroutine to read the output
	go func() {
		defer close(done)

		// Use a Scanner to read the output line by line
		scanner := bufio.NewScanner(attachResp.Reader)
		for scanner.Scan() {
			line := scanner.Text()

			// Write each line to the WebSocket connection
			err := ws.WriteMessage(websocket.TextMessage, []byte(line))
			if err != nil {
				done <- err
				return
			}
		}

		if err := scanner.Err(); err != nil {
			done <- err
			return
		}
		// Log before sending the __END__ message
		log.Println("Sending __END__ message")

		// Send the __END__ message after the command finishes executing
		err := ws.WriteMessage(websocket.TextMessage, []byte("__END__"))
		if err != nil {
			done <- err
			return
		}
	}()

	// Wait for the command to finish
	if err := <-done; err != nil {
		return err
	}

	return nil
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err)
	}

	for {
		_, p, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			return
		}

		segments, err := shlex.Split(string(p))
		if err != nil {
			log.Printf("Error splitting command: %v", err)
			return
		}
		fmt.Println(segments)

		if len(segments) < 4 || segments[0] != "docker" || segments[1] != "exec" {
			errMsg := "Invalid command. Format should be: 'docker exec [OPTIONS] CONTAINER COMMAND [ARG...]'"
			log.Println(errMsg)
			err := ws.WriteJSON(map[string]string{"error": errMsg})
			if err != nil {
				log.Printf("Error writing JSON: %v", err)
			}

			// Send the __END__ message
			err = ws.WriteMessage(websocket.TextMessage, []byte("__END__"))
			if err != nil {
				log.Printf("Error sending __END__ message: %v", err)
			}

			continue
		}

		var options string
		var containerID string
		var command []string

		if strings.HasPrefix(segments[2], "-") {
			options = segments[2]
			containerID = segments[3]
			command = segments[4:]
		} else {
			containerID = segments[2]
			command = segments[3:]
		}

		outputErr := execCommand(cli, containerID, command, options, ws)
		if outputErr != nil {
			log.Printf("Error executing command: %v", outputErr)
			return
		}
	}
}

func main() {
	http.HandleFunc("/ws", handleConnections)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
