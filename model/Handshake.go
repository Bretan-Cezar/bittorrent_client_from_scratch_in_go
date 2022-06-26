package model

import (
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

func Read(r io.Reader) (*Handshake, error) {

	return nil, nil
}
