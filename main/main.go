package main

import (
	"example/bittorrent_in_go/service"
	"os"
)

func main() {

	service := service.NewTorrentService(os.Args[1])

	service.EstablishHandshakes()

	service.CloseConnections()
}
