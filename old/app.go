package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/google/shlex"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	latestContainerID string
)

func createContainer(cli *client.Client) (string, error) {
	ctx := context.Background()

	fmt.Println("Pulling ubuntu image...")
	reader, err := cli.ImagePull(ctx, "ubuntu", types.ImagePullOptions{})
	if err != nil {
		return "", err
	}
	io.Copy(io.Discard, reader)

	fmt.Println("Creating container...")
	resp, err := cli.ContainerCreate(
		ctx,
		&container.Config{
			Image:     "ubuntu",
			Cmd:       []string{"/bin/bash"},
			Tty:       true,
			OpenStdin: true,
		},
		nil, nil, nil, "",
	)
	if err != nil {
		return "", err
	}

	latestContainerID = resp.ID

	fmt.Printf("Container %s created!\n", resp.ID)

	fmt.Println("Starting container.......")

	err = cli.ContainerStart(ctx, latestContainerID, types.ContainerStartOptions{})

	if err != nil {
		return "", err
	}

	fmt.Printf("Started container! The Container ID is :%s\n", latestContainerID)

	return resp.ID, nil
}

func execCommand(cli *client.Client, command []string, ws *websocket.Conn) error {
	ctx := context.Background()

	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          command,
	}

	execIDResp, err := cli.ContainerExecCreate(ctx, latestContainerID, execConfig)
	if err != nil {
		return err
	}

	attachResp, err := cli.ContainerExecAttach(ctx, execIDResp.ID, types.ExecStartCheck{})
	if err != nil {
		return err
	}
	defer attachResp.Close()

	done := make(chan error)

	// Start a goroutine to read the output
	go func() {
		defer close(done)

		scanner := bufio.NewScanner(attachResp.Reader)
		for scanner.Scan() {
			line := scanner.Text()

			// Replace invalid UTF-8 characters
			line = strings.ToValidUTF8(line, "\uFFFD")

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

		err := ws.WriteMessage(websocket.TextMessage, []byte("__END__"))
		if err != nil {
			done <- err
			return
		}
	}()

	if err := <-done; err != nil {
		return err
	}

	return nil
}

func handleConnections(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err)
	}

	_, err = createContainer(cli)
	if err != nil {
		log.Fatal(err)
	}

	for {
		_, p, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
			return err
		}

		command, err := shlex.Split(string(p))
		if err != nil {
			log.Println(err)
			return err
		}

		err = execCommand(cli, command, ws)
		if err != nil {
			log.Println(err)
			return err
		}
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/ws", handleConnections)
	e.Start(":8080")
}
