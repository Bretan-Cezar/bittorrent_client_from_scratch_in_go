package model

import (
	"crypto/sha1"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type bencodeTrackerResp struct {
	Interval int
	Peers    string
}

func MakeTorrentFile(path string) (torrent *TorrentFile) {

	torrent = new(TorrentFile)

	bto, _ := readTorrent(path)

	torrent.Announce = bto.Announce
	torrent.Length = bto.Info.Length
	torrent.Name = bto.Info.Name
	torrent.PieceLength = bto.Info.PieceLength

	p := bto.Info.Pieces

	for p != "" {

		var newHash [20]byte

		copy(newHash[:], p[:20])

		torrent.PieceHashes = append(torrent.PieceHashes, newHash)

		p = p[20:]
	}

	torrent.InfoHash = sha1.Sum([]byte(bto.Info.Encoded))

	return
}

func (file *TorrentFile) buildTrackerURL(peerID [20]byte, port uint16) (string, error) {

	base, err := url.Parse(file.Announce)

	if err != nil {
		return "", err
	}

	params := url.Values{

		"info_hash":  []string{string(file.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(file.Length)},
	}

	base.RawQuery = params.Encode()

	return base.String(), nil
}

func (t *TorrentFile) RequestPeers(peerID [20]byte, port uint16) ([]Peer, error) {

	url, err := t.buildTrackerURL(peerID, port)

	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 15 * time.Second}

	resp, err := client.Get(url)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var respBinary []byte

	if resp.StatusCode == http.StatusOK {

		respBinary, err = io.ReadAll(resp.Body)

		if err != nil {
			return nil, err
		}

	} else {

		return nil, errors.New(resp.Status)
	}

	tokens := decode(string(respBinary))

	respData, _ := decodeItem(tokens)

	trackerResp := bencodeTrackerResp{

		Interval: respData.d["interval"].i,
		Peers:    respData.d["peers"].s,
	}

	if err != nil {
		return nil, err
	}

	return createPeersFromBinary([]byte(trackerResp.Peers))
}
