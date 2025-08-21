package changedpi

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"os"
)

// 常量定义
const (
	pngHeader  = "\x89PNG\r\n\x1a\n"
	jpegHeader = "\xFF\xD8\xFF"

	// JPEG 相关常量
	jpegMarkerSOI  = 0xD8 // Start of Image
	jpegMarkerSOS  = 0xDA // Start of Scan
	jpegMarkerAPP0 = 0xE0 // JFIF 应用段
	jpegMarkerAPP1 = 0xE1 // EXIF 应用段

	// TIFF 标签
	tiffTagXResolution = 0x011A
	tiffTagYResolution = 0x011B

	// DPI 转换系数 (1 inch = 0.0254 meter)
	dpiToPpmFactor = 39.3700787
)

// ErrUnsupportedFormat 表示不支持的图片格式
var (
	ErrUnsupportedFormat  = errors.New("图片格式不支持")
	ErrInvalidJPEG        = errors.New("无效的JPEG文件")
	ErrInvalidJPEGSegment = errors.New("无效的JPEG段")
	ErrInvalidPNG         = errors.New("无效的PNG文件")
)

// ChangeDpi 修改图片的DPI
// inputPath: 输入图片路径
// outputPath: 输出图片路径
// dpi: 目标DPI值
func ChangeDpi(inputPath, outputPath string, dpi int) error {
	// 检查图片格式
	_, err := IsImage(inputPath)
	if err != nil {
		return ErrUnsupportedFormat
	}

	// 读取图片数据
	imgData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("读取图片失败: %w", err)
	}

	// 验证图片内部格式
	imgType, err := checkImageType(imgData)
	if err != nil {
		return err
	}

	// 根据图片类型修改DPI
	var newData []byte
	switch imgType {
	case JPEG, JPG:
		newData, err = changeDpiByJpeg(imgData, dpi)
	case PNG:
		newData, err = changeDpiByPng(imgData, dpi)
	default:
		return ErrUnsupportedFormat
	}

	if err != nil {
		return fmt.Errorf("修改DPI失败: %w", err)
	}

	// 写入输出文件
	if err = SaveBytes(outputPath, newData); err != nil {
		return fmt.Errorf("保存图片失败: %w", err)
	}

	return nil
}

// changeDpiByPng 修改PNG图片的DPI
func changeDpiByPng(data []byte, dpi int) ([]byte, error) {
	return insertOrReplacePhysChunk(data, float64(dpi))
}

// dpiToPpm 将DPI转换为像素/米
func dpiToPpm(dpi float64) uint32 {
	return uint32(dpi * dpiToPpmFactor)
}

// buildPhysChunk 构造PNG的pHYs chunk
func buildPhysChunk(dpi float64) []byte {
	ppm := dpiToPpm(dpi)
	data := make([]byte, 9)
	binary.BigEndian.PutUint32(data[0:4], ppm) // X分辨率
	binary.BigEndian.PutUint32(data[4:8], ppm) // Y分辨率
	data[8] = 1                                // 单位为米

	chunkType := []byte("pHYs")
	crc := crc32.ChecksumIEEE(append(chunkType, data...))

	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint32(len(data))) // 长度
	buf.Write(chunkType)                                    // 类型
	buf.Write(data)                                         // 数据
	binary.Write(&buf, binary.BigEndian, crc)               // CRC校验
	return buf.Bytes()
}

// insertOrReplacePhysChunk 插入或替换PNG的pHYs chunk
func insertOrReplacePhysChunk(data []byte, dpi float64) ([]byte, error) {
	if !bytes.HasPrefix(data, []byte(pngHeader)) {
		return nil, ErrInvalidPNG
	}

	physChunk := buildPhysChunk(dpi)
	physMarker := []byte("pHYs")
	var out bytes.Buffer

	// 查找pHYs的起始位置
	idx := bytes.Index(data, physMarker)
	if idx >= 0 {
		// 替换已有pHYs
		if idx < 4 {
			return nil, ErrInvalidPNG
		}
		start := idx - 4
		length := binary.BigEndian.Uint32(data[start:idx])
		end := idx + 4 + int(length) + 4
		if end > len(data) {
			return nil, ErrInvalidPNG
		}
		out.Write(data[:start])
		out.Write(physChunk)
		out.Write(data[end:])
	} else {
		// 插入到第一个IDAT前
		idatIdx := bytes.Index(data, []byte("IDAT"))
		if idatIdx < 0 || idatIdx < 4 {
			return nil, ErrInvalidPNG
		}
		insertPos := idatIdx - 4
		out.Write(data[:insertPos])
		out.Write(physChunk)
		out.Write(data[insertPos:])
	}

	return out.Bytes(), nil
}

// checkImageType 检查图片二进制数据的类型
func checkImageType(data []byte) (ImageType, error) {
	if bytes.HasPrefix(data, []byte(pngHeader)) {
		return PNG, nil
	}
	if bytes.HasPrefix(data, []byte(jpegHeader)) {
		if len(data) >= 3 && data[2] == 0xFF {
			return JPEG, nil
		}
	}
	return "", ErrUnsupportedFormat
}

// changeDpiByJpeg 修改JPEG图片的DPI（JFIF或EXIF）
func changeDpiByJpeg(data []byte, dpi int) ([]byte, error) {
	// 验证JPEG头
	if len(data) < 4 || data[0] != 0xFF || data[1] != jpegMarkerSOI {
		return nil, ErrInvalidJPEG
	}

	// 创建副本以避免修改原始数据
	result := make([]byte, len(data))
	copy(result, data)

	// 遍历所有段
	i := 2
	for i < len(result) {
		if result[i] != 0xFF {
			return nil, ErrInvalidJPEGSegment
		}

		marker := result[i+1]
		if marker == jpegMarkerSOS { // SOS段，图像数据开始
			break
		}

		// 段长度包含长度字段本身(2字节)
		if i+2 >= len(result) {
			return nil, ErrInvalidJPEGSegment
		}
		segLen := int(binary.BigEndian.Uint16(result[i+2:])) + 2
		if i+segLen > len(result) {
			return nil, ErrInvalidJPEGSegment
		}

		segEnd := i + segLen

		switch marker {
		case jpegMarkerAPP0: // APP0 (JFIF)
			if segLen >= 16 {
				// 修改JFIF中的X和Y分辨率
				binary.BigEndian.PutUint16(result[i+0x0C:], uint16(dpi))
				binary.BigEndian.PutUint16(result[i+0x0E:], uint16(dpi))
			}
		case jpegMarkerAPP1: // APP1 (EXIF)
			updateExifDpi(result, i, segLen, dpi)
		}

		i = segEnd
	}

	return result, nil
}

// updateExifDpi 更新EXIF中的DPI信息
func updateExifDpi(data []byte, offset, segLen, dpi int) {
	// 检查是否为EXIF段
	if segLen < 20 || offset+10 > len(data) {
		return
	}

	if string(data[offset+4:offset+10]) != "Exif\x00\x00" {
		return
	}

	// 找到TIFF头
	tiffStart := offset + 10
	if tiffStart+8 > len(data) {
		return
	}

	// 判断字节序
	var byteOrder binary.ByteOrder
	if data[tiffStart] == 'I' && data[tiffStart+1] == 'I' {
		byteOrder = binary.LittleEndian
	} else if data[tiffStart] == 'M' && data[tiffStart+1] == 'M' {
		byteOrder = binary.BigEndian
	} else {
		return // 无效的TIFF头
	}

	// 找到第一个IFD
	ifdOffset := byteOrder.Uint32(data[tiffStart+4:])
	ifdStart := int(tiffStart) + int(ifdOffset)
	if ifdStart+2 > len(data) {
		return
	}

	numEntries := byteOrder.Uint16(data[ifdStart:])
	pos := ifdStart + 2

	// 遍历IFD条目
	for j := 0; j < int(numEntries) && pos+12 <= len(data); j++ {
		tag := byteOrder.Uint16(data[pos:])
		if tag == tiffTagXResolution || tag == tiffTagYResolution {
			// 修改分辨率值
			valOffset := byteOrder.Uint32(data[pos+8:])
			valAddr := int(tiffStart) + int(valOffset)
			if valAddr+8 <= len(data) {
				byteOrder.PutUint32(data[valAddr:], uint32(dpi))
				byteOrder.PutUint32(data[valAddr+4:], 1)
			}
		}
		pos += 12
	}
}
