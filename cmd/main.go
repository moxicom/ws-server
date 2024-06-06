package main

import "github.com/moxicom/ws-server/internal/server"

func main() {
	s := server.New()
	s.Run(":8080")

}
