package main

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func main() {
	// Create a Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// Create a container
	containerID, err := createContainer(cli)
	if err != nil {
		panic(err)
	}

	// Start the container
	err = startContainer(cli, containerID)
	if err != nil {
		panic(err)
	}

	// Delete the container instance
	err = deleteInstance(cli, containerID)
	if err != nil {
		panic(err)
	}

	// Delete the container
	err = deleteContainer(cli, containerID)
	if err != nil {
		panic(err)
	}
}

// createContainer creates a Docker container with the specified image.
func createContainer(cli *client.Client) (string, error) {
	fmt.Println("Inside createContainer")
	resp, err := cli.ContainerCreate(context.Background(), &container.Config{
		Image: "nginx:latest",
		Cmd:   []string{"echo", "Whoooooooooo! I solved it! I'm a cool engineer!"},
		Tty:   false,
	}, nil, nil, nil, "")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Created container ID: %s\n", resp.ID)
	return resp.ID, nil
}

// startContainer starts a Docker container instance.
func startContainer(cli *client.Client, containerID string) error {
	fmt.Println("Inside startContainer")
	err := cli.ContainerStart(context.Background(), containerID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	fmt.Printf("Started container instance ID: %s\n", containerID)
	return nil
}

// deleteInstance stops and removes a Docker container instance.
func deleteInstance(cli *client.Client, containerID string) error {
	fmt.Println("Inside deleteInstance")
	err := cli.ContainerStop(context.Background(), containerID, container.StopOptions{})
	if err != nil {
		return err
	}

	fmt.Printf("Stopped container instance ID: %s\n", containerID)

	err = cli.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{})
	if err != nil {
		return err
	}

	fmt.Printf("(From Delete Instance) - Removed container instance ID: %s\n", containerID)
	return nil
}

// deleteContainer removes a Docker container.
func deleteContainer(cli *client.Client, containerID string) error {
	fmt.Println("Inside deleteContainer")
	err := cli.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{})
	if err != nil {
		return err
	}

	fmt.Printf("(From Delete Container) - Removed container ID: %s\n", containerID)
	return nil
}
