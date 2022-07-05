package service

import (
	"bytes"
	"crypto/sha1"
	"example/bittorrent_in_go/model"
	"fmt"
	"os"
	"time"

	tm "github.com/buger/goterm"
)

// MaxBlockSize is the largest number of bytes a request can ask for
const MaxBlockSize = 65536

// MaxBacklog is the number of unfulfilled requests a client can have in its pipeline
const MaxBacklog = 5

type TorrentService struct {
	PeerID      [20]byte
	Torrent     *model.TorrentFile
	Clients     []*model.Client
	WorkQueue   chan *pieceWork
	ResultQueue chan *pieceResult
}

type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type pieceResult struct {
	index int
	buf   []byte
}

type pieceProgress struct {
	index      int
	client     *model.Client
	buf        []byte
	downloaded int
	requested  int
	backlog    int
}

func NewTorrentService(torrentPath string) (service *TorrentService) {

	service = new(TorrentService)

	copy(service.PeerID[:], "-TR0000-l9ik1xhfk7di")

	service.Torrent = model.MakeTorrentFile(torrentPath)

	service.WorkQueue = make(chan *pieceWork, len(service.Torrent.PieceHashes))
	service.ResultQueue = make(chan *pieceResult)

	return service
}

func (service *TorrentService) CreateClients() {

	peersList, err := service.Torrent.RequestPeers(service.PeerID, 54788)
	if err != nil {

		fmt.Println(err)
		return
	}

	clients_ch := make(chan *model.Client)

	for _, peer := range peersList {

		go model.NewClient(peer, service.Torrent.InfoHash, service.PeerID, clients_ch)
	}

	for index := 0; index < len(peersList); index++ {

		client := <-clients_ch

		if client == nil {

			continue
		}

		fmt.Printf("Successfully connected to %s.\n", client.Peer.ToString())

		service.Clients = append(service.Clients, client)
	}
}

func (service *TorrentService) CloseConnections() {

	for _, client := range service.Clients {

		client.Connection.Close()
	}
}

func (state *pieceProgress) readMessage() error {

	msg, err := state.client.Read() // blocking
	if err != nil {
		return err
	}

	if msg == nil {
		return nil
	}

	switch msg.ID {

	case model.MsgUnchoke:
		state.client.Choked = false

	case model.MsgChoke:
		state.client.Choked = true

	case model.MsgHave:
		index, err := msg.ParseHave()
		if err != nil {
			return err
		}
		state.client.Bitfield.HasPiece(index)

	case model.MsgPiece:
		n, err := msg.ParsePieceMessage(state.index, state.buf)
		if err != nil {
			return err
		}

		state.downloaded += n
		state.backlog--
	}

	return nil
}

func attemptDownloadPiece(client *model.Client, work *pieceWork) ([]byte, error) {
	state := pieceProgress{
		index:  work.index,
		client: client,
		buf:    make([]byte, work.length),
	}

	// Setting a deadline helps get unresponsive peers unstuck.
	// 30 seconds is more than enough time to download a 256 KB piece
	client.Connection.SetDeadline(time.Now().Add(30 * time.Second))
	defer client.Connection.SetDeadline(time.Time{}) // Disable the deadline

	for state.downloaded < work.length {

		// If unchoked, send requests until we have enough unfulfilled requests
		if !state.client.Choked {
			for state.backlog < MaxBacklog && state.requested < work.length {

				blockSize := MaxBlockSize

				// Last block might be shorter than the typical block
				if work.length-state.requested < blockSize {
					blockSize = work.length - state.requested
				}

				err := client.SendRequest(work.index, state.requested, blockSize)
				if err != nil {
					return nil, err
				}
				state.backlog++
				state.requested += blockSize
			}
		}

		err := state.readMessage()
		if err != nil {
			return nil, err
		}
	}

	return state.buf, nil
}

func checkIntegrity(work *pieceWork, buf []byte) error {

	hash := sha1.Sum(buf)
	if !bytes.Equal(hash[:], work.hash[:]) {
		return fmt.Errorf("index %d failed integrity check", work.index)
	}

	return nil
}

func (service *TorrentService) downloadWorker(clientIndex int) {

	client := service.Clients[clientIndex]

	client.SendUnchoke()
	client.SendInterested()

	for len(service.WorkQueue) > 0 {

		work := <-service.WorkQueue

		if !client.Bitfield.HasPiece(work.index) {

			service.WorkQueue <- work // Put piece back on the queue
			continue
		}

		buffer, err := attemptDownloadPiece(client, work)

		if err != nil {

			service.WorkQueue <- work // Put piece back on the queue
			continue
		}

		err = checkIntegrity(work, buffer)

		if err != nil {

			service.WorkQueue <- work // Put piece back on the queue
			continue
		}

		client.SendHave(work.index)
		service.ResultQueue <- &pieceResult{

			index: work.index,
			buf:   buffer,
		}
	}
}

func (service *TorrentService) writeToFile(buffer []byte) error {

	out, err := os.Create(service.Torrent.Name)

	if err != nil {

		return err
	}
	defer out.Close()

	_, err = out.Write(buffer)

	if err != nil {

		return err
	}

	return nil
}

func (service *TorrentService) Download() error {

	fmt.Printf("\nStarting download for %s...\n", service.Torrent.Name)

	for index, hash := range service.Torrent.PieceHashes {

		service.WorkQueue <- &pieceWork{index, hash, service.Torrent.PieceLength}
	}

	for clientIndex := range service.Clients {

		go service.downloadWorker(clientIndex)
	}

	// go service.downloadWorker(0)

	// Collect results into a buffer until full
	buf := make([]byte, service.Torrent.Length)
	donePieces := 0

	tm.Clear()
	tm.Flush()

	for donePieces < len(service.Torrent.PieceHashes) {

		res := <-service.ResultQueue
		begin, end := res.index*service.Torrent.PieceLength, (res.index+1)*service.Torrent.PieceLength

		copy(buf[begin:end], res.buf)
		donePieces++

		percent := float64(donePieces) / float64(len(service.Torrent.PieceHashes)) * 100

		tm.MoveCursor(1, 1)
		tm.Flush()

		fmt.Printf("(%0.2f%%) Downloaded piece #%-6d from %d peers", percent, res.index, len(service.Clients))
	}

	close(service.WorkQueue)

	return service.writeToFile(buf)
}
