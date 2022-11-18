package model

import (
	"encoding/binary"
	"net"
	"strconv"
)

type Peer struct {
	IP   net.IP
	Port uint16
}

func createPeersFromBinary(peersBinary []byte) ([]Peer, error) {

	peersCount := len(peersBinary) / 6

	peers := make([]Peer, peersCount)

	for index := 0; index < peersCount; index++ {

		currentPeerBinary := peersBinary[:6]

		peers[index] = Peer{

			IP:   net.IP(currentPeerBinary[:4]),
			Port: binary.BigEndian.Uint16(currentPeerBinary[4:]),
		}

		peersBinary = peersBinary[6:]
	}

	return peers, nil
}

func (p Peer) String() string {

	return net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
}
