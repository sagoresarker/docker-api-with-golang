package main

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func createContainer(cli *client.Client) (string, error) {
	fmt.Println("Creating container...")
	resp, err := cli.ContainerCreate(context.Background(), &container.Config{
		Image: "alpine",
		Cmd:   []string{"echo", "Whoooooooooo! I solved it! I'm a cool engineer!"},
		Tty:   false,
	}, nil, nil, nil, "sagoresarker")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Container %s created!\n", resp.ID)
	return resp.ID, nil
}

func startContainer(cli *client.Client, containerID string) error {
	fmt.Println("Starting container.......")

	err := cli.ContainerStart(context.Background(), containerID, types.ContainerStartOptions{})

	if err != nil {
		panic(err)
	}

	fmt.Printf("Started container! The Container ID is :%s\n", containerID)
	return nil
}

func deleteContainer(cli *client.Client, containerID string) error {
	fmt.Println("Deleting container......")

	err := cli.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{})

	if err != nil {
		panic(err)
	}

	fmt.Printf("Deleted container! The Container ID is :%s\n", containerID)
	return nil
}

func main() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	containerID, err := createContainer(cli)
	if err != nil {
		panic(err)
	}

	err = startContainer(cli, containerID)
	if err != nil {
		panic(err)
	}

	err = deleteContainer(cli, containerID)
	if err != nil {
		panic(err)
	}
}
