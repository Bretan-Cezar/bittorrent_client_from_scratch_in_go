package model

type Bitfield []byte

func (bf *Bitfield) HasPiece(index int) bool {

	byteIndex := index / 8
	offsetInByte := byte(index % 8)

	containingByte := (*bf)[byteIndex]

	// Clearing all the bits on its left
	containingByte <<= offsetInByte

	// Bringing the target bit in front
	containingByte >>= 7

	// The final byte should look like: (00000001)2 = (1)10
	return containingByte == 1
}

func (bf *Bitfield) MarkPiece(index int) {

	byteIndex := index / 8
	offsetInByte := byte(index % 8)

	containingByte := (*bf)[byteIndex]

	// E.g.: current byte = 10000010 ; offsetInByte = 3
	// 		 result: 10000010 OR 00010000 = 10010010

	containingByte |= byte(1 << (7 - offsetInByte))

	(*bf)[byteIndex] = containingByte
}
