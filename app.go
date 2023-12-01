package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	latestContainerID string
)

type WebSocketMessage struct {
	Operation   string   `json:"operation"`
	ContainerID string   `json:"containerID"`
	Command     []string `json:"command"`
}

func execCommand(cli *client.Client, containerID string, command []string) (string, error) {
	ctx := context.Background()

	// Create exec instance
	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          command,
	}
	execIDResp, err := cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return "", err
	}

	// Attach to exec instance
	attachResp, err := cli.ContainerExecAttach(ctx, execIDResp.ID, types.ExecStartCheck{})
	if err != nil {
		return "", err
	}
	defer attachResp.Close()

	// Read output
	output, err := io.ReadAll(attachResp.Reader)
	if err != nil {
		return "", err
	}

	return string(output), nil
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
		var msg WebSocketMessage
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading JSON: %v", err)
			return
		}

		switch msg.Operation {
		case "exec":
			var containerToExec string
			if msg.ContainerID != "" {
				containerToExec = msg.ContainerID
			} else if latestContainerID != "" {
				containerToExec = latestContainerID
			} else {
				errMsg := "No containers available to execute command."
				log.Println(errMsg)
				err := ws.WriteJSON(map[string]string{"error": errMsg})
				if err != nil {
					log.Printf("Error writing JSON: %v", err)
				}
				continue
			}

			output, err := execCommand(cli, containerToExec, msg.Command)
			if err != nil {
				log.Printf("Error executing command: %v", err)
				return
			}
			err = ws.WriteJSON(map[string]string{"output": output})
			if err != nil {
				log.Printf("Error writing JSON: %v", err)
			}
		default:
			log.Println("Invalid operation.")
			fmt.Println("Invalid operation.")
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
