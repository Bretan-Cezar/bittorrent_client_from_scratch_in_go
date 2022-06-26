package service

import (
	"example/bittorrent_in_go/model"
	"fmt"
	"net"
	"time"
)

type TorrentService struct {
	PeerID  [20]byte
	Torrent *model.TorrentFile
}

func NewTorrentService(filePath string) (service *TorrentService) {

	service = new(TorrentService)

	copy(service.PeerID[:], "-TR0000-k8hj0wgej6ch")

	service.Torrent = model.MakeTorrentFile("debian-11.3.0-amd64-netinst.iso.torrent")

	return service
}

func connectToPeer(peer model.Peer, c chan net.Conn) {

	addr := peer.ToString()

	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)

	if err != nil {

		fmt.Println(err)
		c <- nil
	}

	c <- conn
}

func (service *TorrentService) Handshakes() {

	peersList, err := service.Torrent.RequestPeers(service.PeerID, 54788)

	conn_ch := make(chan net.Conn)

	if err != nil {

		fmt.Println(err)
		return
	}

	for _, peer := range peersList {

		go connectToPeer(peer, conn_ch)
	}

	for index := 0; index < len(peersList); index++ {

		conn := <-conn_ch
		fmt.Println(conn)

		if conn == nil {

			continue
		}

		conn.Close()
	}
}