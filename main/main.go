package main

import (
	"example/bittorrent_in_go/service"
	"os"
)

func main() {

	service := service.NewTorrentService(os.Args[1])

	service.CreateClients()

	service.Download()

	service.CloseConnections()
}
