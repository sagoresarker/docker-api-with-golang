package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	containertypes "github.com/docker/docker/api/types/container"
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

func createContainer(cli *client.Client) (string, error) {
	fmt.Println("Creating container...")
	resp, err := cli.ContainerCreate(
		context.Background(),
		&container.Config{
			Image: "alpine",
			Cmd:   []string{"sleep", "3600"},
		},
		nil, nil, nil, "",
	)
	if err != nil {
		return "", err
	}

	latestContainerID = resp.ID
	fmt.Printf("Container %s created!\n", resp.ID)
	return resp.ID, nil
}

func startContainer(cli *client.Client, containerID string) error {
	fmt.Println("Starting container.......")

	err := cli.ContainerStart(context.Background(), containerID, types.ContainerStartOptions{})

	if err != nil {
		return err
	}

	fmt.Printf("Started container! The Container ID is :%s\n", containerID)
	return nil
}

func stopContainer(cli *client.Client, containerID string) error {
	fmt.Println("Stopping container.......")

	noWaitTimeout := 10
	err := cli.ContainerStop(context.Background(), containerID, containertypes.StopOptions{Timeout: &noWaitTimeout})
	if err != nil {
		return err
	}

	fmt.Printf("Stopped container! The Container ID was: %s\n", containerID)
	return nil
}

func deleteContainer(cli *client.Client, containerID string) error {
	fmt.Println("Deleting container.......")

	err := cli.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{})

	if err != nil {
		return err
	}

	fmt.Printf("Deleted container! The Container ID was :%s\n", containerID)
	return nil
}

func listRunningContainers(cli *client.Client) ([]types.Container, error) {
	ctx := context.Background()

	// Get list of running containers
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: false})
	if err != nil {
		return nil, err
	}

	fmt.Printf("Found %d running containers\n", len(containers))

	return containers, nil
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
		case "create":
			containerID, err := createContainer(cli)
			if err != nil {
				log.Printf("Error creating container: %v", err)
				return
			}
			err = ws.WriteJSON(map[string]string{"containerID": containerID})
			if err != nil {
				log.Printf("Error writing JSON: %v", err)
			}

		case "start":
			var containerToStart string
			if msg.ContainerID != "" {
				containerToStart = msg.ContainerID
			} else if latestContainerID != "" {
				containerToStart = latestContainerID
			} else {
				log.Println("No containers available to start.")
				continue
			}

			err = startContainer(cli, containerToStart)
			if err != nil {
				log.Printf("Error starting container: %v", err)
				return
			}
			err = ws.WriteJSON(map[string]string{"status": "started"})
			if err != nil {
				log.Printf("Error writing JSON: %v", err)
			}

		case "stop":
			var containerToStop string
			if msg.ContainerID != "" {
				containerToStop = msg.ContainerID
			} else if latestContainerID != "" {
				containerToStop = latestContainerID
			} else {
				log.Println("No containers available to stop.")
				continue
			}

			err := stopContainer(cli, containerToStop)
			if err != nil {
				log.Printf("Error stopping container: %v", err)
				return
			}
			err = ws.WriteJSON(map[string]string{"status": "stopped"})
			if err != nil {
				log.Printf("Error writing JSON: %v", err)
			}

		case "delete":
			var containerToDelete string
			if msg.ContainerID != "" {
				containerToDelete = msg.ContainerID
			} else if latestContainerID != "" {
				containerToDelete = latestContainerID
			} else {
				log.Println("No containers available to delete.")
				continue
			}

			err = deleteContainer(cli, containerToDelete)
			if err != nil {
				log.Printf("Error deleting container: %v", err)
				return
			}
			err = ws.WriteJSON(map[string]string{"status": "deleted"})
			if err != nil {
				log.Printf("Error writing JSON: %v", err)
			}

		case "list":
			containers, err := listRunningContainers(cli)
			if err != nil {
				log.Printf("Error listing containers: %v", err)
				return
			}
			err = ws.WriteJSON(containers)
			if err != nil {
				log.Printf("Error writing JSON: %v", err)
			}

		case "exec":
			var containerToExec string
			if msg.ContainerID != "" {
				containerToExec = msg.ContainerID
			} else if latestContainerID != "" {
				containerToExec = latestContainerID
			} else {
				log.Println("No containers available to execute command.")
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
