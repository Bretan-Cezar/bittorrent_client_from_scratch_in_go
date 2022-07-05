package model

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	MsgChoke         uint8 = 0
	MsgUnchoke       uint8 = 1
	MsgInterested    uint8 = 2
	MsgNotInterested uint8 = 3
	MsgHave          uint8 = 4
	MsgBitfield      uint8 = 5
	MsgRequest       uint8 = 6
	MsgPiece         uint8 = 7
	MsgCancel        uint8 = 8
)

type Message struct {
	ID      uint8
	Payload []byte
}

func (msg *Message) Serialize() (buffer []byte) {

	// Message string:
	// 4 bytes - L (message size, uint32 big-endian)
	// 1 byte - message ID (uint8)
	// L bytes - optional payload

	var payloadLength uint32 = uint32(len(msg.Payload)) + 1

	buffer = make([]byte, 4+payloadLength)

	binary.BigEndian.PutUint32(buffer[0:4], payloadLength)

	buffer[4] = byte(msg.ID)

	copy(buffer[5:], []byte(msg.Payload))

	return

}

func ReadMessage(r io.Reader) (msg *Message, err error) {

	lenBuffer := make([]byte, 4)

	// Reading message length from the ID onward
	_, err = io.ReadFull(r, lenBuffer)

	if err != nil {

		return
	}

	msgLength := binary.BigEndian.Uint32(lenBuffer)

	if msgLength == 0 {

		return
	}

	msgBuffer := make([]byte, msgLength)

	_, err = io.ReadFull(r, msgBuffer)

	if err != nil {

		return
	}

	id := uint8(msgBuffer[0])
	payload := msgBuffer[1:]

	msg = new(Message)
	msg.ID = id
	msg.Payload = payload

	return
}

func MakeRequestMessage(index, begin, length int) *Message {

	payload := make([]byte, 12)

	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))

	return &Message{ID: MsgRequest, Payload: payload}
}

func MakeHaveMessage(index int) *Message {

	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, uint32(index))

	return &Message{ID: MsgHave, Payload: payload}
}

func (msg *Message) ParsePieceMessage(index int, buf []byte) (int, error) {

	if msg.ID != MsgPiece {
		return 0, fmt.Errorf("expected PIECE (%d), got ID %d", MsgPiece, msg.ID)
	}

	if len(msg.Payload) < 8 {
		return 0, fmt.Errorf("payload too short. %d < 8", len(msg.Payload))
	}

	parsedIndex := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	if parsedIndex != index {
		return 0, fmt.Errorf("expected index %d, got %d", index, parsedIndex)
	}

	begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	if begin >= len(buf) {
		return 0, fmt.Errorf("begin offset too high. %d >= %d", begin, len(buf))
	}

	data := msg.Payload[8:]
	if begin+len(data) > len(buf) {
		return 0, fmt.Errorf("data too long (%d) for offset %d with length %d", len(data), begin, len(buf))
	}

	copy(buf[begin:], data)

	return len(data), nil
}

func (msg *Message) ParseHave() (int, error) {

	if msg.ID != MsgHave {
		return 0, fmt.Errorf("expected HAVE (%d), got ID %d", MsgHave, msg.ID)
	}

	if len(msg.Payload) != 4 {
		return 0, fmt.Errorf("expected payload length 4, got length %d", len(msg.Payload))
	}

	index := int(binary.BigEndian.Uint32(msg.Payload))

	return index, nil
}
