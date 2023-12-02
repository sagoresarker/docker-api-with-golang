// /api/handlers.go

package api

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	Websocket "github.com/gorilla/websocket"
	"github.com/sagoresarker/docker-api-with-golang/pkg/docker"
	"github.com/sagoresarker/docker-api-with-golang/pkg/websocket"

	"github.com/docker/docker/client"
	"github.com/google/shlex"
)

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.Upgrader.Upgrade(w, r, nil)
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
			err = ws.WriteMessage(Websocket.TextMessage, []byte("__END__"))
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

		outputErr := docker.ExecCommand(cli, containerID, command, options, ws)
		if outputErr != nil {
			log.Printf("Error executing command: %v", outputErr)
			return
		}
	}
}
