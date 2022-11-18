package model

import (
	"bytes"
	"fmt"
	"net"
	"time"
)

type Client struct {
	Connection net.Conn
	Choked     bool
	Bitfield   Bitfield
	Peer       Peer
	InfoHash   [20]byte
	PeerID     [20]byte
}

func connectToPeer(peer Peer) (net.Conn, error) {

	return net.DialTimeout("tcp", peer.String(), 3*time.Second)
}

func completeHandshake(conn net.Conn, infoHash [20]byte, peerID [20]byte) (*Handshake, error) {

	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	request := NewHandshake(infoHash, peerID)

	_, err := conn.Write(request.Serialize())

	if err != nil {
		return nil, err
	}

	return ReadHandshake(conn)
}

func recvBitfield(conn net.Conn) (Bitfield, error) {

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	msg, err := ReadMessage(conn)

	if err != nil {
		return nil, err
	}

	if msg.ID != MsgBitfield {

		return nil, fmt.Errorf("expected bitfield message (5) but got ID %d", msg.ID)
	}

	return msg.Payload, nil
}

func NewClient(peer Peer, infoHash [20]byte, peerID [20]byte, ch chan *Client) {

	conn, err := connectToPeer(peer)

	if err != nil {

		fmt.Println(err)
		ch <- nil
		return
	}

	response, err := completeHandshake(conn, infoHash, peerID)

	if err != nil {

		fmt.Println(err)
		ch <- nil
		return
	}

	if !bytes.Equal(response.InfoHash[:], infoHash[:]) {

		fmt.Printf("expected infohash %x but got %x\n", response.InfoHash, infoHash)
		ch <- nil
		return
	}

	bitfield, err := recvBitfield(conn)

	if err != nil {

		fmt.Println(err)
		ch <- nil
		return
	}

	ch <- &Client{

		Connection: conn,
		Choked:     true,
		Bitfield:   bitfield,
		Peer:       peer,
		InfoHash:   infoHash,
		PeerID:     peerID,
	}
}

func (c *Client) Read() (*Message, error) {

	return ReadMessage(c.Connection)
}

func (c *Client) SendRequest(index, begin, length int) error {

	req := MakeRequestMessage(index, begin, length)

	_, err := c.Connection.Write(req.Serialize())
	return err
}

func (c *Client) SendInterested() error {

	msg := Message{ID: MsgInterested}

	_, err := c.Connection.Write(msg.Serialize())
	return err
}

func (c *Client) SendNotInterested() error {

	msg := Message{ID: MsgNotInterested}

	_, err := c.Connection.Write(msg.Serialize())
	return err
}

func (c *Client) SendUnchoke() error {

	msg := Message{ID: MsgUnchoke}

	_, err := c.Connection.Write(msg.Serialize())
	return err
}

func (c *Client) SendHave(index int) error {

	msg := MakeHaveMessage(index)

	_, err := c.Connection.Write(msg.Serialize())
	return err
}
