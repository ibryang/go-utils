package changedpi

import (
	"encoding/base64"
	"fmt"
	"strings"
)

var pngDataTable []uint32

func createPngDataTable() []uint32 {
	crcTable := make([]uint32, 256)
	for n := 0; n < 256; n++ {
		var c uint32 = uint32(n)
		for k := 0; k < 8; k++ {
			if c&1 != 0 {
				c = 0xedb88320 ^ (c >> 1)
			} else {
				c >>= 1
			}
		}
		crcTable[n] = c
	}
	return crcTable
}

func calcCrc(buf []byte) uint32 {
	var c uint32 = 0xffffffff
	if pngDataTable == nil {
		pngDataTable = createPngDataTable()
	}
	for _, b := range buf {
		c = pngDataTable[(c^uint32(b))&0xff] ^ (c >> 8)
	}
	return c ^ 0xffffffff
}

const (
	PNG  = "image/png"
	JPEG = "image/jpeg"
	JPG  = "image/jpg"
)

var (
	b64PhysSignature1 = "AAlwSFlz"
	b64PhysSignature2 = "AAAJcEhZ"
	b64PhysSignature3 = "AAAACXBI"

	_P = byte('p')
	_H = byte('H')
	_Y = byte('Y')
	_S = byte('s')
)

func ChangeDpi(base64Image string, dpi int) (string, error) {
	parts := strings.Split(base64Image, ",")
	format := parts[0]
	body := parts[1]

	var imgType string
	var headerLength int
	var overwritepHYs bool

	if strings.Contains(format, PNG) {
		imgType = PNG
		b64Index := detectPhysChunkFromDataUrl(body)
		if b64Index >= 0 {
			headerLength = (b64Index + 28 + 2) / 3 * 4
			overwritepHYs = true
		} else {
			headerLength = int(33.0 / 3.0 * 4.0)
		}
	} else if strings.Contains(format, JPEG) || strings.Contains(format, JPG) {
		imgType = JPEG
		headerLength = int(18.0 / 3.0 * 4.0)
	} else {
		return "", fmt.Errorf("unsupported image format: %s", format)
	}

	stringHeader := body[:headerLength]
	restOfData := body[headerLength:]

	headerBytes, err := base64.StdEncoding.DecodeString(stringHeader)
	if err != nil {
		return "", err
	}

	dataArray := make([]byte, len(headerBytes))
	for i, b := range headerBytes {
		dataArray[i] = b
	}

	finalArray, err := changeDpiOnArray(dataArray, dpi, imgType, overwritepHYs)
	if err != nil {
		return "", err
	}

	base64Header := base64.StdEncoding.EncodeToString(finalArray)
	return fmt.Sprintf("%s,%s%s", format, base64Header, restOfData), nil
}

func detectPhysChunkFromDataUrl(data string) int {
	b64index := strings.Index(data, b64PhysSignature1)
	if b64index == -1 {
		b64index = strings.Index(data, b64PhysSignature2)
	}
	if b64index == -1 {
		b64index = strings.Index(data, b64PhysSignature3)
	}
	return b64index
}

func searchStartOfPhys(data []byte) int {
	length := len(data) - 1
	for i := length; i >= 4; i-- {
		if data[i-4] == 9 && data[i-3] == _P &&
			data[i-2] == _H && data[i-1] == _Y &&
			data[i] == _S {
			return i - 3
		}
	}
	return -1
}

func changeDpiOnArray(dataArray []byte, dpi int, format string, overwritepHYs bool) ([]byte, error) {
	if format == JPEG {
		if len(dataArray) < 18 {
			return nil, fmt.Errorf("invalid JPEG data")
		}
		dataArray[13] = 1 // 1 pixel per inch or 2 pixel per cm
		dataArray[14] = byte(dpi >> 8)
		dataArray[15] = byte(dpi & 0xff)
		dataArray[16] = byte(dpi >> 8)
		dataArray[17] = byte(dpi & 0xff)
		return dataArray, nil
	}

	if format == PNG {
		// this multiplication is because the standard is dpi per meter.
		dpi = int(float64(dpi) * 39.3701)
		physChunk := make([]byte, 13)
		physChunk[0] = _P
		physChunk[1] = _H
		physChunk[2] = _Y
		physChunk[3] = _S
		physChunk[4] = byte(dpi >> 24)
		physChunk[5] = byte(dpi >> 16)
		physChunk[6] = byte(dpi >> 8)
		physChunk[7] = byte(dpi & 0xff)
		physChunk[8] = physChunk[4]
		physChunk[9] = physChunk[5]
		physChunk[10] = physChunk[6]
		physChunk[11] = physChunk[7]
		physChunk[12] = 1

		crc := calcCrc(physChunk)

		crcChunk := []byte{byte(crc >> 24), byte(crc >> 16), byte(crc >> 8), byte(crc)}
		if overwritepHYs {
			startingIndex := searchStartOfPhys(dataArray)
			if startingIndex == -1 {
				return nil, fmt.Errorf("pHYs chunk not found")
			}
			copy(dataArray[startingIndex:], physChunk)
			copy(dataArray[startingIndex+13:], crcChunk)
			return dataArray, nil
		} else {
			chunkLength := []byte{0, 0, 0, 9}
			finalHeader := make([]byte, 54)
			copy(finalHeader, dataArray[:33])
			copy(finalHeader[33:], chunkLength)
			copy(finalHeader[37:], physChunk)
			copy(finalHeader[50:], crcChunk)
			return finalHeader, nil
		}
	}

	return nil, fmt.Errorf("unsupported image format: %s", format)
}
