package changedpi

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

type ImageType string

const (
	JPEG ImageType = ".jpeg"
	JPG  ImageType = ".jpg"
	PNG  ImageType = ".png"
)

func LoadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func LoadImageBytes(data []byte) image.Image {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}
	return img
}

func GetImageType(path string) (ImageType, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	_, _, err = image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	ext := strings.ToLower(filepath.Ext(path))
	return ImageType(ext), nil
}

func SaveImage(path string, img image.Image) error {
	ext := filepath.Ext(path)
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	if ext == ".jpeg" || ext == ".jpg" || ext == ".JPG" || ext == ".JPEG" {
		return jpeg.Encode(file, img, nil)
	}
	if ext == ".png" || ext == ".PNG" {
		return png.Encode(file, img)
	}
	return errors.New("图片格式不支持")
}

func SaveBytes(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

// 判断图片格式
func IsImage(path string) (ImageType, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == string(JPEG) || ext == string(JPG) || ext == string(PNG) {
		return ImageType(ext), nil
	}
	return "", errors.New("图片格式不支持")
}

// GetImageBounds 获取图片的边界
func GetImageBounds(img image.Image) (minX, minY, maxX, maxY int) {
	bounds := img.Bounds()
	return bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Max.Y
}

// GetContentBounds 获取图片中实际有内容的边界
func GetContentBounds(img image.Image) (minX, minY, maxX, maxY int) {
	bounds := img.Bounds()
	minX, minY = bounds.Max.X, bounds.Max.Y
	maxX, maxY = bounds.Min.X, bounds.Min.Y

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a > 0 { // 非完全透明
				if x < minX {
					minX = x
				}
				if y < minY {
					minY = y
				}
				if x > maxX {
					maxX = x
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}

	// 如果没有找到非透明像素，返回整个图像的边界
	if minX > maxX || minY > maxY {
		return bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Max.Y
	}

	return
}
