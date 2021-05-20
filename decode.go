package gosilk

import (
	"bytes"
	"errors"
	lib "github.com/jdeng/gosilk/lib"
	"log"
	"modernc.org/libc"
	"unsafe"
)

type decoder struct {
	dec uintptr
	tls *libc.TLS
}

func newDecoder() *decoder {
	tls := libc.NewTLS()

	var size int32
	lib.XSKP_Silk_SDK_Get_Decoder_Size(tls, uintptr(unsafe.Pointer(&size)))

	dec := libc.Xmalloc(tls, uint64(size))
	ret := lib.XSKP_Silk_SDK_InitDecoder(tls, dec)
	if ret < 0 {
		libc.Xfree(tls, dec)
		return nil
	}

	return &decoder{tls: tls, dec: dec}
}

func freeDecoder(d *decoder) {
	libc.Xfree(d.tls, d.dec)
}

func (d *decoder) decodeFrame(sampleRate int32, payload []byte, buf []byte) (int, error) {
	c := &lib.SKP_SILK_SDK_DecControlStruct{FframesPerPacket: 1, FAPI_sampleRate: sampleRate}
	var used int16
	ret := lib.XSKP_Silk_SDK_Decode(d.tls, d.dec, uintptr(unsafe.Pointer(c)), 0, uintptr(unsafe.Pointer(&payload[0])), int32(len(payload)), uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&used)))
	if ret < 0 {
		return 0, errors.New("decode error")
	}

	return int(used) * 2, nil
}

func Decode(buf []byte, sampleRate int32, withHeader bool) ([]byte, error) {
	const MAXSIZE = 20 * 48 * 5 * 2 * 2
	var b bytes.Buffer
	out := make([]byte, MAXSIZE)

	if withHeader {
		const TAG = "#!SILK_V3"
		if len(buf) < 1+len(TAG) || buf[0] != 0x02 || bytes.Compare(buf[1:1+len(TAG)], []byte(TAG)) != 0 {
			return nil, errors.New("invalid header")
		}
		buf = buf[1+len(TAG):]
	}

	d := newDecoder()
	defer freeDecoder(d)

	for {
		if len(buf) < 2 {
			break
		}
		plen := int(uint16(buf[0]) + (uint16(buf[1]) << 8))

		if len(buf) < 2+plen {
			log.Printf("%d bytes expected but only %d available\n", int(plen), len(buf))
			break
		}
		payload := buf[2 : 2+plen]
		buf = buf[2+plen:]

		olen, _ := d.decodeFrame(sampleRate, payload, out)
		if olen < 0 {
			log.Printf("Failed to decode %d bytes\n", plen)
			break

		}
		b.Write(out[:olen])
	}

	return b.Bytes(), nil
}
