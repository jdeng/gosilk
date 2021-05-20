package main

import (
	"encoding/binary"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"bytes"
	"github.com/jdeng/gosilk"
	"log"
)

type wavHeader struct {
	ChunkID       [4]byte
	ChunkSize     uint32
	Format        [4]byte
	Subchunk1ID   [4]byte
	Subchunk1Size uint32
	AudioFormat   uint16
	NumChannels   uint16
	SampleRate    uint32
	ByteRate      uint32
	BlockAlign    uint16
	BitsPerSample uint16
	Subchunk2ID   [4]byte
	Subchunk2Size uint32
}

const SAMPLE_RATE = 24000

func convert(buf []byte) ([]byte, error) {
	out, err := gosilk.Decode(buf, SAMPLE_RATE, true)
	if err != nil {
		return nil, err
	}

	datalen := uint32(len(out))
	hdr := wavHeader{
		ChunkID:       [4]byte{'R', 'I', 'F', 'F'},
		ChunkSize:     36 + datalen,
		Format:        [4]byte{'W', 'A', 'V', 'E'},
		Subchunk1ID:   [4]byte{'f', 'm', 't', ' '},
		Subchunk1Size: 16,
		AudioFormat:   1, //1 = PCM not compressed
		NumChannels:   1,
		SampleRate:    SAMPLE_RATE,
		ByteRate:      2 * SAMPLE_RATE,
		BlockAlign:    2,
		BitsPerSample: 16,
		Subchunk2ID:   [4]byte{'d', 'a', 't', 'a'},
		Subchunk2Size: datalen,
	}

	var b bytes.Buffer
	binary.Write(&b, binary.LittleEndian, hdr)
	b.Write(out)
	return b.Bytes(), nil
}

func main() {
	fname := os.Args[1]
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Fatal(err)
	}

	out, err := convert(data)
	if err != nil {
		log.Fatal(err)
	}

	oname := filepath.Base(fname)
	oname = strings.TrimSuffix(oname, path.Ext(oname)) + ".wav"
	ioutil.WriteFile(oname, out, 0644)
}
