package test

import (
	"example/bittorrent_in_go/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBitfield(t *testing.T) {

	bf := make(model.Bitfield, 4)

	bf[0] = 1   // 00000001
	bf[1] = 128 // 10000000
	bf[2] = 8   // 00001000
	bf[3] = 56  // 00011100

	indexes := map[int]bool{

		7: true, 8: true, 20: true, 26: true, 27: true, 28: true,
	}

	for index := 0; index < 32; index++ {

		_, ok := indexes[index]

		if ok {

			assert.Equal(t, bf.HasPiece(index), true)

		} else {

			assert.Equal(t, bf.HasPiece(index), false)
		}
	}

	bf.MarkPiece(15)

	assert.Equal(t, bf.HasPiece(15), true)
}
