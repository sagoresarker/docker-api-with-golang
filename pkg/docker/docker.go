// /pkg/docker/docker.go

package docker

import (
	"bufio"
	"context"
	"log"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gorilla/websocket"
)

func ExecCommand(cli *client.Client, containerID string, command []string, options string, ws *websocket.Conn) error {
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
