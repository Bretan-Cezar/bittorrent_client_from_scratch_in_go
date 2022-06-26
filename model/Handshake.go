package model

import (
	"fmt"
	"io"
)

type Handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerID   [20]byte
}

func NewHandshake(infoHash, peerID [20]byte) *Handshake {

	return &Handshake{

		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

func (hs *Handshake) Serialize() []byte {

	// Handshake string:
	// 1 byte - L (length of protocol ID string in base 16) +
	// L bytes - protocol ID string +
	// 8 bytes - reserved +
	// 20 bytes - info hash string +
	// 20 bytes - peer ID string

	buf := make([]byte, 1+len(hs.Pstr)+8+20+20)

	buf[0] = byte(len(hs.Pstr))

	index := 1

	index += copy(buf[index:], hs.Pstr)
	index += copy(buf[index:], make([]byte, 8))
	index += copy(buf[index:], hs.InfoHash[:])
	index += copy(buf[index:], hs.PeerID[:])

	return buf
}

func ReadHandshake(r io.Reader) (*Handshake, error) {

	hexLength := make([]byte, 1)

	_, err := io.ReadFull(r, hexLength)
	if err != nil {
		return nil, err
	}

	protocolLength := int(hexLength[0])

	if protocolLength == 0 {
		err := fmt.Errorf("protocolLength cannot be 0")
		return nil, err
	}

	handshakeBuf := make([]byte, protocolLength+8+20+20)

	_, err = io.ReadFull(r, handshakeBuf)
	if err != nil {
		return nil, err
	}

	var infoHash, peerID [20]byte

	copy(infoHash[:], handshakeBuf[protocolLength+8:protocolLength+8+20])
	copy(peerID[:], handshakeBuf[protocolLength+8+20:])

	hs := Handshake{
		Pstr:     string(handshakeBuf[0:protocolLength]),
		InfoHash: infoHash,
		PeerID:   peerID,
	}

	return &hs, nil
}
