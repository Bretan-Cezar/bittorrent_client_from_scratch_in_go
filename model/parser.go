package model

import (
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

/*
	Custom torrent file parser inspired by:
	https://web.archive.org/web/20200105114449/https://effbot.org/zone/bencode.htm
*/

type decodedVariant struct {
	t string
	s string
	i int
	d map[string]decodedVariant
	l []decodedVariant
}

type bencodeInfo struct {
	Encoded     string
	Pieces      string
	PieceLength int
	Length      int
	Name        string
}

type bencodeTorrent struct {
	Announce     string
	Comment      string
	CreationDate int
	HttpSeeds    []string
	Info         *bencodeInfo
}

func tokenize(text string, c chan string) {

	expr1 := "[idel]"
	expr2 := "\\d+:"
	expr3 := "\\d+"

	re1, _ := regexp.Compile(expr1)
	re2, _ := regexp.Compile(expr2)
	re3, _ := regexp.Compile(expr3)

	for text != "" {

		m1 := string(re1.Find([]byte(text)))
		m2 := string(re2.Find([]byte(text)))
		m3 := string(re3.Find([]byte(text)))

		i1 := strings.Index(text, m1)
		i2 := strings.Index(text, m2)
		i3 := strings.Index(text, m3)

		if (i1 < i2 && i1 < i3) || (m2 == "" && m3 == "") {

			c <- m1
			text = text[1:]

		} else if i3 < i1 && i3 < i2 {

			c <- m3
			text = text[len(m3):]

		} else {

			if m2 != "" {

				c <- "s"

				lenStr := m2[:len(m2)-1]

				length, _ := strconv.Atoi(lenStr)

				if length != 0 {

					c <- text[len(lenStr)+1 : length+len(lenStr)+1]

				} else {

					c <- "<nil>"
				}

				text = text[length+len(m2):]

			} else {

				text = ""
			}

		}
	}

	c <- ""
}

func decode(bencode string) []string {

	ch := make(chan string)

	go tokenize(bencode, ch)

	var tokens []string

	token := "-"

	for token != "" {

		token = <-ch

		if token != "" {

			tokens = append(tokens, token)
		}
	}

	return tokens
}

func decodeItem(tokens []string) (data *decodedVariant, newTokens []string) {

	if tokens[0] == "i" {

		// integer: i <value> e
		data = new(decodedVariant)

		data.t = "i"
		data.i, _ = strconv.Atoi(tokens[1])

		newTokens = tokens[3:]

	} else if tokens[0] == "s" {

		// string: s <value>
		data = new(decodedVariant)

		data.t = "s"
		data.s = tokens[1]

		newTokens = tokens[2:]

	} else if tokens[0] == "l" || tokens[0] == "d" {

		// container: [l/d] <values> e

		data = new(decodedVariant)
		data.t = tokens[0]

		tokens = tokens[1:]

		switch data.t {

		case "l":

			for tokens[0] != "e" {

				var item *decodedVariant
				item, tokens = decodeItem(tokens)

				data.l = append(data.l, *item)
			}

		case "d":

			data.d = make(map[string]decodedVariant)

			for tokens[0] != "e" {

				var key *decodedVariant
				var value *decodedVariant

				key, tokens = decodeItem(tokens)
				value, tokens = decodeItem(tokens)

				data.d[key.s] = *value

			}
		}

		newTokens = tokens[1:]
	}

	return
}

func createObjects(tokens []string) *bencodeTorrent {

	data, _ := decodeItem(tokens)

	torrent := new(bencodeTorrent)

	torrent.Announce = data.d["announce"].s
	torrent.Comment = data.d["comment"].s
	torrent.CreationDate = data.d["creation date"].i

	seeds := data.d["httpseeds"].l

	for _, v := range seeds {

		torrent.HttpSeeds = append(torrent.HttpSeeds, v.s)
	}

	dataInfo := data.d["info"].d

	info := new(bencodeInfo)

	info.Length = dataInfo["length"].i
	info.Name = dataInfo["name"].s
	info.PieceLength = dataInfo["piece length"].i
	info.Pieces = dataInfo["pieces"].s

	torrent.Info = info

	return torrent
}

func readTorrent(path string) (torrent *bencodeTorrent, err error) {

	torrentBytes, err := ioutil.ReadFile(path)

	torrentString := string(torrentBytes)

	infoStartIndex := strings.Index(torrentString, "4:infod") + 6

	encodedInfo := torrentString[infoStartIndex : len(torrentString)-1]

	array := decode(torrentString)

	torrent = createObjects(array)

	torrent.Info.Encoded = encodedInfo

	return
}
