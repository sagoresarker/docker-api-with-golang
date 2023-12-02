package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
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

func execCommand(cli *client.Client, containerID string, command []string, options string) ([]string, error) {
	ctx := context.Background()

	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          command,
	}

	if strings.Contains(options, "-i") {
		execConfig.AttachStdin = true
	}
	if strings.Contains(options, "-t") {
		execConfig.Tty = true
	}

	execIDResp, err := cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return []string{}, err
	}

	attachResp, err := cli.ContainerExecAttach(ctx, execIDResp.ID, types.ExecStartCheck{})
	if err != nil {
		return []string{}, err
	}
	defer attachResp.Close()

	var outputBuf, errorBuf bytes.Buffer

	_, err = stdcopy.StdCopy(&outputBuf, &errorBuf, attachResp.Reader)
	if err != nil {
		return []string{}, err
	}
	output := outputBuf.String() + errorBuf.String()

	lines := strings.Split(string(output), "\n")

	return lines, nil
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
			continue
		}

		options := segments[2]
		containerID := segments[3]
		command := segments[4:]

		output, err := execCommand(cli, containerID, command, options)
		if err != nil {
			log.Printf("Error executing command: %v", err)
			return
		}
		outputStr := strings.Join(output, "\n")
		err = ws.WriteMessage(websocket.TextMessage, []byte(outputStr))
		if err != nil {
			log.Printf("Error writing message: %v", err)
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
