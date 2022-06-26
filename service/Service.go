package service

import (
	"bytes"
	"example/bittorrent_in_go/model"
	"fmt"
	"net"
	"time"
)

type TorrentService struct {
	PeerID      [20]byte
	Torrent     *model.TorrentFile
	Connections []net.Conn
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

func (service *TorrentService) completeHandshake(conn net.Conn) (*model.Handshake, error) {

	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	request := model.NewHandshake(service.Torrent.InfoHash, service.PeerID)

	_, err := conn.Write(request.Serialize())
	if err != nil {
		return nil, err
	}

	response, err := model.ReadHandshake(conn)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(response.InfoHash[:], service.Torrent.InfoHash[:]) {

		return nil, fmt.Errorf("expected infohash %x but got %x", response.InfoHash, service.Torrent.InfoHash)
	}

	return response, nil
}

func (service *TorrentService) EstablishHandshakes() {

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

		if conn == nil {

			continue
		}

		resp, err := service.completeHandshake(conn)

		if err != nil {

			continue
		}

		fmt.Printf("%v ; %p\n", conn, resp)

		service.Connections = append(service.Connections, conn)
	}
}

func (service *TorrentService) CloseConnections() {

	for _, conn := range service.Connections {

		conn.Close()
	}
}
