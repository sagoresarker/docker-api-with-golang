// /cmd/myapp/main.go

package main

import (
	"net/http"

	"github.com/sagoresarker/docker-api-with-golang/api"
)

func main() {
	http.HandleFunc("/ws", api.HandleConnections)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
